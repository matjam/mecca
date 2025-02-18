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

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/x/ansi"
	"github.com/creack/pty"
	"github.com/muesli/termenv"
	"golang.org/x/text/encoding/charmap"
)

// Updated colorMapping to use ANSI 16-color codes.
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

// Create a termenv.Output from the session.
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
		templateRoot:  ".",
		session:       nil,
		renderer:      lipgloss.NewRenderer(os.Stdout),
		output:        termenv.NewOutput(os.Stdout),
		tokenRegistry: make(map[string]Token),
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

// Session returns the SSH session associated with the interpreter. If the session
// is not set, Session returns nil.
func (i *Interpreter) Session() ssh.Session {
	return i.session
}

// RegisterToken registers a new token with the interpreter. The token name is
// case-insensitive. The token function is called with the token's arguments and
// should return the substitution string. You can use a type method as a token
// function, as shown in the example.
func (i *Interpreter) RegisterToken(name string, fn TokenFunc, argCount int) {
	if _, ok := i.tokenRegistry[strings.ToLower(name)]; ok {
		panic(fmt.Sprintf("token %s already registered", name))
	}

	i.tokenRegistry[strings.ToLower(name)] = Token{
		Func:     fn,
		ArgCount: argCount,
	}
}

// GetToken retrieves a token by name. The name is case-insensitive.
func (i *Interpreter) GetToken(name string) (Token, bool) {
	token, ok := i.tokenRegistry[strings.ToLower(name)]
	return token, ok
}

// ExecTemplate reads a template file from the template root directory and
// executes it. The filename is relative to the template root directory.
// Returns the rendered output or an error if the file cannot be read.
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

// RenderFile reads a file relative to the template root directory and renders it
// using the interpreter's writer. Returns an error if the file cannot be read.
func (interpreter *Interpreter) RenderTemplate(filename string, vars map[string]any) error {
	templatePath := path.Join(interpreter.templateRoot, filename)
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	output := interpreter.interpret(string(data), vars, []string{filename})

	io.WriteString(interpreter.session, output)
	return nil
}

// RenderString processes the input string containing MECCA tokens and renders it
// using the interpreter's writer. Any included templates are resolved relative to
// the template root directory specified when creating the interpreter.
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
	for {
		start := strings.Index(input, "[")
		if start == -1 {
			// Split literal text on newline, style each line then add newline back.
			lines := strings.Split(input, "\n")
			for i, line := range lines {
				output += currentStyle.Render(line)
				if i < len(lines)-1 {
					output += "\n"
				}
			}
			break
		}
		// Process literal text before token.
		literal := input[:start]
		lines := strings.Split(literal, "\n")
		for i, line := range lines {
			output += currentStyle.Render(line)
			if i < len(lines)-1 {
				output += "\n"
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
		output += interpreter.processToken(tokenContent, &currentStyle, vars, includes)
		input = input[start+end+1:]
	}
	return output
}

// processToken function to wrap returned token text with the current style.
func (interpreter *Interpreter) processToken(content string, style *lipgloss.Style, vars map[string]any, includes []string) string {
	parts := strings.Fields(content)
	result := ""
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
			continue
		case "locate": // [locate] token: expects two arguments.
			if i+2 < len(parts) {
				col, err1 := strconv.Atoi(parts[i+1])
				row, err2 := strconv.Atoi(parts[i+2])
				if err1 == nil && err2 == nil {
					// ANSI escape sequence: CSI row;colH (adding 1 for 1-indexing)
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
		case "include": // [include] token: expects one argument.
			if i+1 < len(parts) {
				filename := parts[i+1]

				// Prevent infinite recursion.
				for _, inc := range includes {
					if inc == filename {
						result += fmt.Sprintf("[ERROR: %s included recursively]", filename)
						return result
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
				result += style.Render(string(r))
				continue
			}
		} else if isNumber(part) {
			// Handle ASCII token from a number.
			if n, err := parseNumber(part); err == nil {
				result += style.Render(string(rune(n)))
				continue
			}
		}
		// Handle variable tokens.
		if vars != nil {
			if val, ok := vars[part]; ok {
				result += style.Render(fmt.Sprintf("%v", val))
				continue
			}
		}
		// Look up registered tokens.
		if token, ok := interpreter.GetToken(part); ok {
			if token.ArgCount > 0 && i+token.ArgCount < len(parts) {
				args := parts[i+1 : i+1+token.ArgCount]
				tokenOut := token.Func(args)
				result += style.Render(tokenOut)
				i += token.ArgCount
			} else {
				tokenOut := token.Func([]string{})
				result += style.Render(tokenOut)
			}
			continue
		}
		// If token is unrecognized, emit an error message inline.
		result += style.Render(fmt.Sprintf("[UNRECOGNIZED TOKEN \"%s\"]", part))
	}
	return result
}

// isNumber checks if a string consists solely of digits.
func isNumber(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// parseNumber converts a numeric string to an integer.
func parseNumber(s string) (int, error) {
	var n int
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// decodeUTF8Token converts a hexadecimal string to a rune.
func decodeUTF8Token(hexStr string) (rune, error) {
	var n int
	_, err := fmt.Sscanf(hexStr, "%x", &n)
	if err != nil {
		return 0, err
	}
	return rune(n), nil
}

// isColorToken to check the colorMapping keys.
func isColorToken(s string) bool {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "#") {
		return true
	}
	_, exists := colorMapping[s]
	return exists
}

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
