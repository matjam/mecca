// Package mecca provides a MECCA language interpreter for creating terminal-based
// user interfaces. MECCA (originally designed for Maximus BBS software) is a simple,
// easy-to-learn templating language designed for non-programmers to create interactive
// terminal content.
//
// This implementation provides a subset of the original MECCA language, adapted for
// modern Go applications using the lipgloss library for rendering. It's designed to
// work with the bubbletea framework in modern BBS software.
//
// The interpreter processes MECCA templates containing tokens enclosed in square
// brackets, such as [red], [bold], [user], etc. Tokens can be used for colors, text
// styling, cursor movement, file inclusion, and custom functionality through
// registered tokens.
//
// Basic Usage:
//
//	interpreter := mecca.NewInterpreter()
//	interpreter.ExecString("[bold][red]Hello, World![reset]", nil)
//
// With custom tokens:
//
//	interpreter := mecca.NewInterpreter()
//	interpreter.RegisterToken("user", func(args []string) string {
//		return "John Doe"
//	}, 0)
//	interpreter.ExecString("Welcome, [user]!", nil)
//
// See the README.md file for complete documentation of all supported tokens.
package mecca

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv" // new import for locate token
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/x/ansi"
	"github.com/creack/pty"
	"github.com/muesli/termenv"
	"golang.org/x/text/encoding/charmap"
)

// colorMapping maps color names to their ANSI 16-color codes for use with lipgloss.
// Supported colors include the standard 8 colors and their light variants.
var colorMapping = map[string]lipgloss.Color{
	"black":        lipgloss.Color("0"),
	"red":          lipgloss.Color("1"),
	"green":        lipgloss.Color("2"),
	"yellow":       lipgloss.Color("3"),
	"blue":         lipgloss.Color("4"),
	"magenta":      lipgloss.Color("5"),
	"cyan":         lipgloss.Color("6"),
	"white":        lipgloss.Color("7"),
	"lightblack":   lipgloss.Color("8"),
	"lightred":     lipgloss.Color("9"),
	"lightgreen":   lipgloss.Color("10"),
	"lightyellow":  lipgloss.Color("11"),
	"lightblue":    lipgloss.Color("12"),
	"lightmagenta": lipgloss.Color("13"),
	"lightcyan":    lipgloss.Color("14"),
	"lightwhite":   lipgloss.Color("15"),
}

// outputFromSession creates a termenv.Output from an SSH session.
// This bridges the SSH session with termenv's output capabilities, allowing
// the interpreter to query terminal capabilities and render output appropriately
// for the remote terminal. The function uses unsafe mode since we already know
// the session is a TTY.
func outputFromSession(sess ssh.Session) *termenv.Output {
	sshPty, _, _ := sess.Pty()
	_, tty, err := pty.Open()
	if err != nil {
		log.Fatal(err)
	}
	o := &sshOutput{
		Session: sess,
		tty:     tty,
	}
	environ := sess.Environ()
	environ = append(environ, fmt.Sprintf("TERM=%s", sshPty.Term))
	e := &sshEnviron{environ: environ}
	// We need to use unsafe mode here because the ssh session is not running
	// locally and we already know that the session is a TTY.
	return termenv.NewOutput(o, termenv.WithUnsafe(), termenv.WithEnvironment(e))
}

// NewInterpreter creates a new MECCA interpreter with a default template root
// directory and output writer. The template root directory is the current
// working directory, and the output writer is os.Stdout. You can customize the
// template root directory and output writer by passing options to NewInterpreter.
func NewInterpreter(options ...func(*Interpreter)) *Interpreter {
	interpreter := &Interpreter{
		templateRoot:     ".",
		session:          nil,
		renderer:         lipgloss.NewRenderer(os.Stdout),
		output:           termenv.NewOutput(os.Stdout),
		reader:           nil,
		tokenRegistry:    make(map[string]Token),
		styleStack:       []lipgloss.Style{},
		menuOptions:      make(map[string]string),
		menuResponse:     "",
		inMenu:           false,
		capturingOption:  false,
		currentOptionID:  "",
		optionTextBuffer: strings.Builder{},
	}

	for _, option := range options {
		option(interpreter)
	}

	return interpreter
}

// WithTemplateRoot is a functional option to set the template root directory for
// NewInterpreter. The template root directory is the base directory for all
// template files. If a template includes another template, the included template
// is resolved relative to the template root directory. The default template root
// directory is the current working directory.
func WithTemplateRoot(root string) func(*Interpreter) {
	return func(i *Interpreter) {
		i.templateRoot = root
	}
}

// WithSession is a functional option to set the SSH session for the interpreter.
// SSH sessions are from github.com/charmbracelet/ssh. If a session is set, the
// interpreter uses the session's output for rendering. If a session is not set,
// the interpreter uses os.Stdout as the output writer.
func WithSession(sess ssh.Session) func(*Interpreter) {
	return func(i *Interpreter) {
		i.session = sess
		i.output = outputFromSession(sess)
		i.renderer = lipgloss.NewRenderer(sess)

		i.renderer.SetOutput(i.output)
	}
}

// WithWriter is a functional option to set the output writer for the interpreter.
// The output writer is used to render the output of the interpreter. The default
// output writer is os.Stdout. You can use WithWriter to set a different output
// writer, such as a file or buffer.
func WithWriter(w io.Writer) func(*Interpreter) {
	return func(i *Interpreter) {
		i.output = termenv.NewOutput(w)
		i.renderer = lipgloss.NewRenderer(i.output)
	}
}

// WithReader is a functional option to set the input reader for the interpreter.
// The input reader is used for interactive features such as menus. When a [menuwait]
// token is encountered, the interpreter reads from this reader to get user input.
// The default reader is nil, which means interactive features will not work unless
// a reader is provided.
func WithReader(r io.Reader) func(*Interpreter) {
	return func(i *Interpreter) {
		i.reader = r
	}
}

// Session returns the SSH session associated with the interpreter. If the session
// is not set, Session returns nil.
func (i *Interpreter) Session() ssh.Session {
	return i.session
}

// MenuResponse returns the selected option ID from the most recent menu interaction.
// This should be called after processing a template that contains a [menuwait] token.
// Returns an empty string if no menu response has been recorded yet.
func (i *Interpreter) MenuResponse() string {
	return i.menuResponse
}

// RegisterToken registers a new custom token with the interpreter. The token
// name is case-insensitive. When the token is encountered in a MECCA template,
// the provided function is called with the token's arguments and should return
// the substitution string.
//
// The argCount parameter specifies how many arguments the token expects. If a
// token is used with fewer arguments than specified, the function will receive
// an empty slice.
//
// You can use a type method as a token function. Registered tokens can be
// overridden by variables passed at execution time with the same name.
//
// This function will panic if a token with the same name is already registered.
//
// Example:
//
//	type Server struct {
//		user string
//	}
//	server := &Server{user: "Alice"}
//	interpreter.RegisterToken("user", server.userToken, 0)
//	// Now [user] in templates will be replaced with "Alice"
func (i *Interpreter) RegisterToken(name string, fn TokenFunc, argCount int) {
	if _, ok := i.tokenRegistry[strings.ToLower(name)]; ok {
		panic(fmt.Sprintf("token %s already registered", name))
	}

	i.tokenRegistry[strings.ToLower(name)] = Token{
		Func:     fn,
		ArgCount: argCount,
	}
}

// GetToken retrieves a registered token by name. The name is case-insensitive.
// Returns the token and a boolean indicating whether the token was found.
// This is primarily used internally but can be useful for introspection.
func (i *Interpreter) GetToken(name string) (Token, bool) {
	token, ok := i.tokenRegistry[strings.ToLower(name)]
	return token, ok
}

// ExecTemplate reads a template file from the template root directory and
// executes it, returning the rendered output as a string. The filename is
// resolved relative to the template root directory set when creating the
// interpreter.
//
// Returns the rendered output string and an error if the file cannot be read.
// Unlike RenderTemplate, this method returns the output rather than writing
// it directly to the output writer, making it useful when you need to process
// or modify the output before displaying it.
//
// Example:
//
//	output, err := interpreter.ExecTemplate("welcome.mec", map[string]any{"user": "Alice"})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Print(output)
func (interpreter *Interpreter) ExecTemplate(filename string, vars map[string]any) (string, error) {
	templatePath := path.Join(interpreter.templateRoot, filename)
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	return interpreter.interpret(string(data), vars, []string{filename}), nil
}

// ExecString processes the input string containing MECCA tokens and executes it.
// The output is written to the interpreter's writer, and any included templates
// are resolved relative to the template root directory specified when creating
// the interpreter.
func (interpreter *Interpreter) ExecString(input string, vars map[string]any) {
	io.WriteString(interpreter.output, interpreter.interpret(input, vars, []string{}))
}

// RenderTemplate reads a template file from the template root directory and
// renders it using the interpreter's output writer. The filename is resolved
// relative to the template root directory set when creating the interpreter.
//
// Returns an error if the file cannot be read. The function handles both SSH
// sessions and regular output writers, so it's safe to call even when the
// interpreter was created without an SSH session.
//
// Example:
//
//	interpreter := mecca.NewInterpreter(mecca.WithTemplateRoot("templates"))
//	err := interpreter.RenderTemplate("welcome.mec", map[string]any{"user": "Alice"})
func (interpreter *Interpreter) RenderTemplate(filename string, vars map[string]any) error {
	templatePath := path.Join(interpreter.templateRoot, filename)
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	output := interpreter.interpret(string(data), vars, []string{filename})

	if interpreter.session != nil {
		wish.WriteString(interpreter.session, output)
		return nil
	}

	io.WriteString(interpreter.output, output)
	return nil
}

// RenderString processes the input string containing MECCA tokens and renders it
// using the interpreter's output writer. Any included templates are resolved
// relative to the template root directory specified when creating the interpreter.
//
// Unlike ExecString, this method uses wish.WriteString when an SSH session is
// present, which provides better handling for SSH connections. For non-SSH usage,
// ExecString and RenderString behave identically.
//
// The vars parameter can be used to pass variables that will override registered
// tokens with the same name.
func (interpreter *Interpreter) RenderString(input string, vars map[string]any) {
	if interpreter.session != nil {
		wish.WriteString(interpreter.session, interpreter.interpret(input, vars, []string{}))
		return
	}

	io.WriteString(interpreter.output, interpreter.interpret(input, vars, []string{}))
}

// interpret processes the input string containing MECCA tokens and literal text,
// applies the current styling via lipgloss, and returns the rendered output.
func (interpreter *Interpreter) interpret(input string, vars map[string]any, includes []string) string {
	output := ""
	currentStyle := interpreter.renderer.NewStyle()
	// Reset style stack for each interpretation
	interpreter.styleStack = []lipgloss.Style{}
	// Reset menu state for each interpretation
	interpreter.menuOptions = make(map[string]string)
	interpreter.menuResponse = ""
	interpreter.inMenu = false
	interpreter.capturingOption = false
	interpreter.currentOptionID = ""
	interpreter.optionTextBuffer.Reset()
	for {
		start := strings.Index(input, "[")
		if start == -1 {
			// Split literal text on newline, style each line then add newline back.
			lines := strings.Split(input, "\n")
			for i, line := range lines {
				rendered := currentStyle.Render(line)
				output += rendered
				// If we're capturing option text, accumulate the plain text
				if interpreter.capturingOption {
					interpreter.optionTextBuffer.WriteString(line)
				}
				if i < len(lines)-1 {
					output += "\n"
					if interpreter.capturingOption {
						interpreter.optionTextBuffer.WriteString("\n")
					}
				}
			}
			break
		}
		// Check for escaped bracket ([[)
		if start+1 < len(input) && input[start+1] == '[' {
			// Process literal text before escaped bracket, including the first bracket
			literal := input[:start+1] // Include first '['
			lines := strings.Split(literal, "\n")
			for i, line := range lines {
				output += currentStyle.Render(line)
				if i < len(lines)-1 {
					output += "\n"
				}
			}
			// Skip the second '[' and continue
			input = input[start+2:]
			continue
		}
		// Process literal text before token.
		literal := input[:start]
		lines := strings.Split(literal, "\n")
		for i, line := range lines {
			rendered := currentStyle.Render(line)
			output += rendered
			// If we're capturing option text, accumulate the plain text
			if interpreter.capturingOption {
				interpreter.optionTextBuffer.WriteString(line)
			}
			if i < len(lines)-1 {
				output += "\n"
				if interpreter.capturingOption {
					interpreter.optionTextBuffer.WriteString("\n")
				}
			}
		}
		end := strings.Index(input[start:], "]")
		if end == -1 {
			// unmatched token, render remainder as-is.
			remainder := input[start:]
			lines = strings.Split(remainder, "\n")
			for i, line := range lines {
				output += currentStyle.Render(line)
				if i < len(lines)-1 {
					output += "\n"
				}
			}
			break
		}
		tokenContent := input[start+1 : start+end]
		// Check if this is a menuwait token before processing (we need to flush first)
		parts := parseFieldsWithQuotes(tokenContent)
		isMenuWait := len(parts) > 0 && strings.ToLower(parts[0]) == "menuwait"

		// If it's menuwait, flush accumulated output first before processing
		if isMenuWait && interpreter.reader != nil {
			if interpreter.session != nil {
				wish.WriteString(interpreter.session, output)
			} else {
				io.WriteString(interpreter.output, output)
			}
			output = ""
		}

		tokenResult, shouldFlush := interpreter.processToken(tokenContent, &currentStyle, vars, includes, output)
		output += tokenResult

		// If we need to flush output after processing (shouldn't happen with menuwait since we flush before)
		if shouldFlush {
			if interpreter.session != nil {
				wish.WriteString(interpreter.session, output)
			} else {
				io.WriteString(interpreter.output, output)
			}
			output = ""
		}
		input = input[start+end+1:]
	}

	// If we're still capturing an option at the end, save it
	if interpreter.capturingOption && interpreter.currentOptionID != "" {
		interpreter.menuOptions[interpreter.currentOptionID] = strings.TrimSpace(interpreter.optionTextBuffer.String())
		interpreter.capturingOption = false
		interpreter.currentOptionID = ""
		interpreter.optionTextBuffer.Reset()
	}

	return output
}

// processToken processes a single token's content, updating the style state and
// returning the rendered output. Token content may contain multiple space-separated
// tokens, which are processed in order.
//
// Supported tokens include:
//   - Color tokens (e.g., [red], [#FF0000], [lightblue on white])
//   - Style tokens (e.g., [bold], [underline], [blink])
//   - Cursor control tokens (e.g., [locate 5 10], [up], [down])
//   - File inclusion tokens (e.g., [include file.mec], [ansi file.ans])
//   - Menu tokens (e.g., [menu], [option a "Option A"], [menuwait])
//   - Custom registered tokens
//   - Variable substitutions
//   - ASCII/UTF-8 code tokens (e.g., [65] for 'A', [U+00A9] for '©')
//
// The style parameter is modified in-place as tokens are processed. The function
// returns the rendered output string for this token and a boolean indicating whether
// output should be flushed (for interactive tokens like [menuwait]).
func (interpreter *Interpreter) processToken(content string, style *lipgloss.Style, vars map[string]any, includes []string, accumulatedOutput string) (string, bool) {
	parts := parseFieldsWithQuotes(content)
	result := ""
	shouldFlush := false
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		// Handle special tokens.
		switch strings.ToLower(part) {
		case "cls": // [cls] token: clear screen.
			result += ansi.EraseDisplay(2) + ansi.CursorPosition(1, 1)
			continue
		case "cleos": // [cleos] token: clear to end of screen.
			result += ansi.EraseDisplay(0)
			continue
		case "cleol": // [cleol] token: clear to end of line.
			result += ansi.EraseLine(0)
		case "blink": // [blink] token: blink text.
			*style = style.Blink(true)
			continue
		case "steady": // [steady] token: cancel blink attribute.
			*style = style.Blink(false)
			continue
		case "bright": // [bright] token: brighten text. (synonym for bold)
			fallthrough
		case "bold": // [bold] token: bold text.
			*style = style.Bold(true)
			continue
		case "underline": // [underline] token: underline text.
			*style = style.Underline(true)
			continue
		case "italic": // [italic] token: italicize text.
			*style = style.Italic(true)
			continue
		case "dim": // [dim] token: dim text.
			*style = style.Faint(true)
			continue
		case "reverse": // [reverse] token: reverse text.
			*style = style.Reverse(true)
			continue
		case "strike": // [strike] token: strike through text.
			*style = style.Strikethrough(true)
			continue
		case "reset": // [reset] token: remove all styling.
			*style = interpreter.renderer.NewStyle()
			// If we were capturing an option, end the capture and save it
			if interpreter.capturingOption && interpreter.currentOptionID != "" {
				interpreter.menuOptions[interpreter.currentOptionID] = strings.TrimSpace(interpreter.optionTextBuffer.String())
				interpreter.capturingOption = false
				interpreter.currentOptionID = ""
				interpreter.optionTextBuffer.Reset()
			}
			continue
		case "save": // [save] token: save current color and style.
			interpreter.styleStack = append(interpreter.styleStack, *style)
			continue
		case "load": // [load] token: restore saved color and style.
			if len(interpreter.styleStack) > 0 {
				*style = interpreter.styleStack[len(interpreter.styleStack)-1]
				interpreter.styleStack = interpreter.styleStack[:len(interpreter.styleStack)-1]
			}
			continue
		case "locate": // [locate] token: expects two arguments (row, column).
			if i+2 < len(parts) {
				row, err1 := strconv.Atoi(parts[i+1])
				col, err2 := strconv.Atoi(parts[i+2])
				if err1 == nil && err2 == nil {
					// ANSI escape sequence: CSI row;colH (adding 1 for 1-indexing)
					// README says: [locate <row> <column>]
					result += ansi.CursorPosition(row+1, col+1)
				}
				i += 2
			}
			continue
		case "cr": // [cr] token: moves the cursor to the beginning of the line.
			result += "\r"
			continue
		case "lf": // [lf] token: moves the cursor to the next line.
			result += ansi.CursorNextLine(1)
			continue
		case "up": // [up] token: moves the cursor up one line.
			result += ansi.CursorUp(1)
			continue
		case "down": // [down] token: moves the cursor down one line.
			result += ansi.CursorDown(1)
			continue
		case "right": // [right] token: moves the cursor right one column.
			result += ansi.CursorForward(1)
			continue
		case "left": // [left] token: moves the cursor left one column.
			result += ansi.CursorBackward(1)
			continue
		case "savecursor": // [savecursor] token: saves the current cursor position.
			result += ansi.SaveCursor
			continue
		case "restorecursor": // [restorecursor] token: restores the saved cursor position.
			result += ansi.RestoreCursor
			continue
		case "line": // draws a line of a specified length using the specified character.
			if i+2 < len(parts) {
				length, err1 := strconv.Atoi(parts[i+1])
				char := parts[i+2]
				if err1 == nil {
					result += strings.Repeat(char, length)
				}
				i += 2
			}
			continue
		case "on": // [ON <color>] token: set background color only (expects color argument).
			if i+1 < len(parts) {
				bgToken := strings.ToLower(parts[i+1])
				if strings.HasPrefix(bgToken, "#") {
					if len(bgToken) == 7 {
						*style = style.Background(lipgloss.Color(bgToken))
					} else {
						// ANSI color code
						if n, err := parseNumber(bgToken[1:]); err == nil {
							*style = style.Background(lipgloss.Color(strconv.Itoa(n)))
						}
					}
				} else if bg, ok := colorMapping[bgToken]; ok {
					*style = style.Background(bg)
				}
				i++
			}
			continue
		case "include": // [include] token: expects one argument.
			if i+1 < len(parts) {
				filename := parts[i+1]

				// Prevent infinite recursion.
				for _, inc := range includes {
					if inc == filename {
						result += fmt.Sprintf("[ERROR: %s included recursively]", filename)
						return result, shouldFlush
					}
				}

				if output, err := interpreter.ExecTemplate(filename, vars); err == nil {
					result += output
				} else {
					result += fmt.Sprintf("[ERROR: %v]", err)
				}
				i++
			}
			continue
		case "ansi": // [ansi] token: expects one argument, the ANSI file to include.
			if i+1 < len(parts) {
				filename := parts[i+1]
				data := path.Join(interpreter.templateRoot, filename)

				if output, err := os.ReadFile(data); err == nil {
					result += string(output)
				} else {
					result += fmt.Sprintf("[ERROR: %v]", err)
				}

				i++
			}
			continue
		case "ansiconvert": // [ansiconvert] token: expects two arguments, the ANSI file to convert and the input charset.
			if i+2 < len(parts) {
				filename := parts[i+1]
				charset := parts[i+2]
				data := path.Join(interpreter.templateRoot, filename)

				if output, err := os.ReadFile(data); err == nil {
					result += convertFromCharset(string(output), charset)
				} else {
					result += fmt.Sprintf("[ERROR: %v]", err)
				}

				i += 2
			}
			continue
		case "menu": // [menu] token: starts a new menu, clearing any existing options.
			interpreter.inMenu = true
			interpreter.menuOptions = make(map[string]string)
			continue
		case "option": // [option option_id] token: starts capturing option text until [reset].
			if i+1 < len(parts) {
				optionID := strings.ToLower(parts[i+1])

				// Validate option_id: must be a single alphanumeric character
				if len(optionID) == 1 && isValidOptionID(optionID) {
					// If we were already capturing an option, save it first
					if interpreter.capturingOption && interpreter.currentOptionID != "" {
						interpreter.menuOptions[interpreter.currentOptionID] = strings.TrimSpace(interpreter.optionTextBuffer.String())
						interpreter.optionTextBuffer.Reset()
					}
					// Start capturing this option
					interpreter.capturingOption = true
					interpreter.currentOptionID = optionID
					// Display the option ID in the current style
					result += style.Render(strings.ToUpper(optionID))
				} else {
					result += style.Render(fmt.Sprintf("[ERROR: invalid option_id %s, must be single alphanumeric character]", parts[i+1]))
				}
				i++
			}
			continue
		case "menuwait": // [menuwait] token: waits for user input and reads the selected option.
			if interpreter.reader == nil {
				result += style.Render("[ERROR: no reader configured, use WithReader() option]")
				continue
			}

			// Signal that output should be flushed before waiting for input
			shouldFlush = true

			// Read a single character from the reader
			// This will happen after the accumulated output is flushed
			buf := make([]byte, 1)
			n, err := interpreter.reader.Read(buf)
			if err != nil || n == 0 {
				interpreter.menuResponse = ""
				continue
			}

			// Convert to lowercase for case-insensitive matching
			inputChar := strings.ToLower(string(buf[0]))

			// Check if the input matches any option ID
			if _, ok := interpreter.menuOptions[inputChar]; ok {
				interpreter.menuResponse = inputChar
			} else {
				interpreter.menuResponse = ""
			}
			continue
		}
		// Colors can be specified one of three ways:
		// 1. By name (e.g., red, green, blue)
		// 2. By hex code (e.g., #ff0000, #00ff00, #0000ff)
		// 3. By ANSI color code, as a number (e.g., #63)
		if isColorToken(part) {
			tokenLower := strings.ToLower(part)
			if strings.HasPrefix(tokenLower, "#") {
				if len(tokenLower) == 7 {
					*style = style.Foreground(lipgloss.Color(tokenLower))
				} else {
					// ANSI color code
					if n, err := parseNumber(tokenLower[1:]); err == nil {
						*style = style.Foreground(lipgloss.Color(strconv.Itoa(n)))
					}
				}
			} else if col, ok := colorMapping[tokenLower]; ok {
				*style = style.Foreground(col)
			}
			if i+2 < len(parts) && strings.ToLower(parts[i+1]) == "on" {
				bgToken := strings.ToLower(parts[i+2])
				if strings.HasPrefix(bgToken, "#") {
					if len(bgToken) == 7 {
						*style = style.Background(lipgloss.Color(bgToken))
					} else {
						// ANSI color code
						if n, err := parseNumber(bgToken[1:]); err == nil {
							*style = style.Background(lipgloss.Color(strconv.Itoa(n)))
						}
					}
				} else if bg, ok := colorMapping[bgToken]; ok {
					*style = style.Background(bg)
				}
				i += 2
			}
			continue
		}
		// Handle UTF-8 tokens [U+xxxx]
		if strings.HasPrefix(part, "U+") && len(part) > 2 {
			if r, err := decodeUTF8Token(part[2:]); err == nil {
				text := string(r)
				result += style.Render(text)
				// If capturing option text, capture the plain text
				if interpreter.capturingOption {
					interpreter.optionTextBuffer.WriteString(text)
				}
				continue
			}
		} else if isNumber(part) {
			// Handle ASCII token from a number.
			if n, err := parseNumber(part); err == nil {
				text := string(rune(n))
				result += style.Render(text)
				// If capturing option text, capture the plain text
				if interpreter.capturingOption {
					interpreter.optionTextBuffer.WriteString(text)
				}
				continue
			}
		}
		// Handle variable tokens.
		if vars != nil {
			if val, ok := vars[part]; ok {
				text := fmt.Sprintf("%v", val)
				result += style.Render(text)
				// If capturing option text, capture the plain text
				if interpreter.capturingOption {
					interpreter.optionTextBuffer.WriteString(text)
				}
				continue
			}
		}
		// Look up registered tokens.
		if token, ok := interpreter.GetToken(part); ok {
			var text string
			if token.ArgCount > 0 && i+token.ArgCount < len(parts) {
				args := parts[i+1 : i+1+token.ArgCount]
				text = token.Func(args)
				result += style.Render(text)
				i += token.ArgCount
			} else {
				text = token.Func([]string{})
				result += style.Render(text)
			}
			// If capturing option text, capture the plain text
			if interpreter.capturingOption {
				interpreter.optionTextBuffer.WriteString(text)
			}
			continue
		}
		// If token is unrecognized, emit an error message inline.
		result += style.Render(fmt.Sprintf("[UNRECOGNIZED TOKEN \"%s\"]", part))
	}
	return result, shouldFlush
}

// parseFieldsWithQuotes parses a string into fields, respecting quoted arguments.
// Quoted strings (e.g., "hello world") are treated as a single field, allowing
// arguments to contain spaces. Escaped quotes within quoted strings are handled
// correctly (e.g., "say \"hello\"").
//
// This function supports the MECCA language feature where token arguments can be
// quoted if they contain spaces. Unquoted arguments are split on whitespace as
// usual.
//
// Example: `token "hello world" arg2` parses to ["token", "hello world", "arg2"]
func parseFieldsWithQuotes(s string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false
	bytes := []byte(s)
	i := 0

	for i < len(bytes) {
		if bytes[i] == '"' {
			if !inQuotes {
				inQuotes = true
			} else {
				// End of quoted string
				if current.Len() > 0 {
					fields = append(fields, current.String())
					current.Reset()
				}
				inQuotes = false
			}
			i++
		} else if bytes[i] == '\\' && inQuotes && i+1 < len(bytes) && bytes[i+1] == '"' {
			// Escaped quote inside quoted string
			current.WriteByte('"')
			i += 2 // Skip both backslash and quote
		} else if bytes[i] == ' ' || bytes[i] == '\t' {
			if !inQuotes {
				if current.Len() > 0 {
					fields = append(fields, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(bytes[i])
			}
			i++
		} else {
			// Decode UTF-8 rune
			r, size := utf8.DecodeRune(bytes[i:])
			if r != utf8.RuneError {
				current.WriteRune(r)
			} else {
				current.WriteByte(bytes[i])
			}
			i += size
		}
	}

	if current.Len() > 0 {
		fields = append(fields, current.String())
	}

	return fields
}

// isNumber checks if a string consists solely of ASCII digits (0-9).
// This is used to identify ASCII code tokens like [65] which should be
// interpreted as ASCII character codes rather than token names.
func isNumber(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// parseNumber converts a numeric string to an integer. This is a simple
// implementation that manually converts each digit, avoiding the overhead of
// strconv.Atoi for the common case of small numbers used in MECCA tokens.
func parseNumber(s string) (int, error) {
	var n int
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// decodeUTF8Token converts a hexadecimal string to a rune. This is used to
// process UTF-8 code tokens like [U+00A9] which should be converted to the
// corresponding Unicode character (in this case, the copyright symbol ©).
func decodeUTF8Token(hexStr string) (rune, error) {
	var n int
	_, err := fmt.Sscanf(hexStr, "%x", &n)
	if err != nil {
		return 0, err
	}
	return rune(n), nil
}

// isColorToken checks if a string is a valid color token. A valid color token
// is either a recognized color name (e.g., "red", "lightblue") or a hex color
// code starting with "#" (e.g., "#FF0000" for true color, "#202" for 256-color).
func isColorToken(s string) bool {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "#") {
		return true
	}
	_, exists := colorMapping[s]
	return exists
}

// isValidOptionID checks if a string is a valid menu option ID.
// A valid option ID is a single alphanumeric character (a-z, A-Z, 0-9).
func isValidOptionID(s string) bool {
	if len(s) != 1 {
		return false
	}
	c := s[0]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// convertFromCharset converts text from a specified character encoding to UTF-8.
// Currently supports CP437 (Code Page 437), which was commonly used in DOS
// BBS systems. This is used by the [ansiconvert] token to convert ANSI art files
// from their original encoding to UTF-8 for display.
func convertFromCharset(input string, charset string) string {
	var b bytes.Buffer
	switch strings.ToLower(charset) {
	case "cp437":
		// Convert from CP437 to UTF-8.
		cp437Reader := charmap.CodePage437.NewDecoder().Reader(strings.NewReader(input))
		io.Copy(&b, cp437Reader)
	default:
		// No conversion.
		b.WriteString(fmt.Sprintf("[ERROR: unsupported charset %s]", charset))
	}
	return b.String()
}
