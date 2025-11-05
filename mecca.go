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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv" // new import for locate token
	"strings"
	"time"
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
		templateRoot:        ".",
		session:             nil,
		renderer:            lipgloss.NewRenderer(os.Stdout),
		output:              termenv.NewOutput(os.Stdout),
		reader:              nil,
		tokenRegistry:       make(map[string]Token),
		styleStack:          []lipgloss.Style{},
		menuOptions:         make(map[string]string),
		menuResponse:        "",
		inMenu:              false,
		capturingOption:     false,
		currentOptionID:     "",
		optionTextBuffer:    strings.Builder{},
		readlnResponse:      "",
		questionnaireData:   []string{},
		answerOptional:      false,
		labels:              make(map[string]int),
		shouldQuit:          false,
		shouldExit:          false,
		gotoTarget:          "",
		shouldGotoTop:       false,
		callStack:           []fileContext{},
		onExitFile:          "",
		shouldDisplay:       false,
		displayFile:         "",
		linkFile:            "",
		shouldLink:          false,
		colorConditionStack: []bool{},
		moreEnabled:         false,
		currentLine:         0,
		terminalHeight:      0,
		moreResponse:        "",
		lastMoreLine:        0,
		shouldHandleMore:    false,
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

// ReadlnResponse returns the last response from a [readln] token.
// Returns an empty string if no [readln] response has been recorded yet.
func (i *Interpreter) ReadlnResponse() string {
	return i.readlnResponse
}

// QuestionnaireData returns all collected questionnaire responses.
// This includes responses from [readln], [store], and any data added via [write].
// Returns a slice of strings where each entry is a questionnaire line.
func (i *Interpreter) QuestionnaireData() []string {
	return i.questionnaireData
}

// ClearQuestionnaireData clears all collected questionnaire data.
func (i *Interpreter) ClearQuestionnaireData() {
	i.questionnaireData = []string{}
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
	// Reset interactive state (but don't clear questionnaireData - it persists across calls)
	interpreter.readlnResponse = ""
	interpreter.shouldQuit = false
	interpreter.shouldExit = false
	interpreter.gotoTarget = ""
	interpreter.shouldGotoTop = false
	interpreter.colorConditionStack = []bool{}
	// Don't reset more system state - it persists across interpretations
	// But reset line counter for new interpretation
	interpreter.currentLine = 0
	interpreter.lastMoreLine = 0
	interpreter.shouldHandleMore = false

	// Get terminal height if not already set
	if interpreter.terminalHeight == 0 {
		interpreter.terminalHeight = interpreter.getTerminalHeight()
	}

	// Parse labels first (before processing tokens)
	interpreter.labels = parseLabels(input)
	originalInput := input
	position := 0
	skipToEndOfLine := false
	skipLine := false

	for position < len(originalInput) {
		remainingInput := originalInput[position:]
		start := strings.Index(remainingInput, "[")
		if start == -1 {
			// Split literal text on newline, style each line then add newline back.
			// Check if we should skip output due to color conditionals
			shouldSkip := false
			if len(interpreter.colorConditionStack) > 0 {
				for _, skip := range interpreter.colorConditionStack {
					if skip {
						shouldSkip = true
						break
					}
				}
			}

			lines := strings.Split(remainingInput, "\n")
			for i, line := range lines {
				if !shouldSkip {
					rendered := currentStyle.Render(line)
					output += rendered
					// Count lines for more system
					interpreter.currentLine++

					// Check for auto-more before adding newline
					if i < len(lines)-1 {
						if interpreter.checkAutoMore() {
							// Flush output before prompting
							if interpreter.session != nil {
								wish.WriteString(interpreter.session, output)
							} else {
								io.WriteString(interpreter.output, output)
							}
							output = ""

							// Handle more prompt
							interpreter.handleMorePrompt()

							// Check if user quit
							if interpreter.shouldQuit {
								return output
							}
						}
						output += "\n"
						interpreter.currentLine++
					}
				} else {
					// Still count lines even when skipping (for accurate positioning)
					interpreter.currentLine++
					if i < len(lines)-1 {
						interpreter.currentLine++
					}
				}
				// If we're capturing option text, accumulate the plain text
				if interpreter.capturingOption {
					interpreter.optionTextBuffer.WriteString(line)
					if i < len(lines)-1 {
						interpreter.optionTextBuffer.WriteString("\n")
					}
				}
			}
			break
		}
		// Check for escaped bracket ([[)
		if start+1 < len(remainingInput) && remainingInput[start+1] == '[' {
			// Process literal text before escaped bracket, including the first bracket
			// Check if we should skip output due to color conditionals
			shouldSkip := false
			if len(interpreter.colorConditionStack) > 0 {
				for _, skip := range interpreter.colorConditionStack {
					if skip {
						shouldSkip = true
						break
					}
				}
			}

			literal := remainingInput[:start+1] // Include first '['
			lines := strings.Split(literal, "\n")
			for i, line := range lines {
				if !shouldSkip {
					output += currentStyle.Render(line)
					interpreter.currentLine++
				} else {
					interpreter.currentLine++
				}
				if i < len(lines)-1 {
					if !shouldSkip {
						output += "\n"
					}
					interpreter.currentLine++
				}
			}
			// Skip the second '[' and continue
			position = position + start + 2
			continue
		}
		// Process literal text before token.
		// Check if we should skip output due to color conditionals
		shouldSkip := false
		if len(interpreter.colorConditionStack) > 0 {
			for _, skip := range interpreter.colorConditionStack {
				if skip {
					shouldSkip = true
					break
				}
			}
		}

		literal := remainingInput[:start]
		lines := strings.Split(literal, "\n")
		for i, line := range lines {
			if !shouldSkip {
				rendered := currentStyle.Render(line)
				output += rendered
				// Count lines for more system
				interpreter.currentLine++

				// Check for auto-more before adding newline
				if i < len(lines)-1 {
					if interpreter.checkAutoMore() {
						// Flush output before prompting
						if interpreter.session != nil {
							wish.WriteString(interpreter.session, output)
						} else {
							io.WriteString(interpreter.output, output)
						}
						output = ""

						// Handle more prompt
						interpreter.handleMorePrompt()

						// Check if user quit
						if interpreter.shouldQuit {
							return output
						}
					}
					output += "\n"
					interpreter.currentLine++
				}
			} else {
				// Still count lines even when skipping (for accurate positioning)
				interpreter.currentLine++
				if i < len(lines)-1 {
					interpreter.currentLine++
				}
			}
			// If we're capturing option text, accumulate the plain text
			if interpreter.capturingOption {
				interpreter.optionTextBuffer.WriteString(line)
				if i < len(lines)-1 {
					interpreter.optionTextBuffer.WriteString("\n")
				}
			}
		}
		end := strings.Index(input[start:], "]")
		if end == -1 {
			// unmatched token, render remainder as-is.
			// Check if we should skip output due to color conditionals
			shouldSkip := false
			if len(interpreter.colorConditionStack) > 0 {
				for _, skip := range interpreter.colorConditionStack {
					if skip {
						shouldSkip = true
						break
					}
				}
			}

			remainder := input[start:]
			lines := strings.Split(remainder, "\n")
			for i, line := range lines {
				if !shouldSkip {
					output += currentStyle.Render(line)
					interpreter.currentLine++
				} else {
					interpreter.currentLine++
				}
				if i < len(lines)-1 {
					if !shouldSkip {
						output += "\n"
					}
					interpreter.currentLine++
				}
			}
			break
		}
		tokenContent := remainingInput[start+1 : start+end]
		tokenPos := position + start

		// Check if this is a menuwait token before processing (we need to flush first)
		parts := parseFieldsWithQuotes(tokenContent)
		isMenuWait := len(parts) > 0 && strings.ToLower(parts[0]) == "menuwait"
		isReadln := len(parts) > 0 && strings.ToLower(parts[0]) == "readln"
		isEnter := len(parts) > 0 && strings.ToLower(parts[0]) == "enter"

		// Check for interactive tokens that need flushing
		if (isMenuWait || isReadln || isEnter) && interpreter.reader != nil {
			if interpreter.session != nil {
				wish.WriteString(interpreter.session, output)
			} else {
				io.WriteString(interpreter.output, output)
			}
			output = ""
		}

		tokenResult, shouldFlush, skipLineAfter := interpreter.processToken(tokenContent, &currentStyle, vars, includes, output, tokenPos, originalInput)

		// Handle [choice] conditional - skip rest of line if condition not met
		if skipLineAfter {
			skipToEndOfLine = true
		}

		// If we should skip to end of line
		if skipToEndOfLine || skipLine {
			// Find the rest of the line and skip to next newline
			lineEnd := strings.Index(remainingInput[start+end+1:], "\n")
			if lineEnd == -1 {
				// No newline found, we're at the end
				break
			}
			// Skip to after the newline
			position = tokenPos + end + 1 + lineEnd + 1
			skipLine = false
			skipToEndOfLine = false
			continue
		}

		// Check if we should skip output due to color conditionals
		shouldSkipOutput := false
		if len(interpreter.colorConditionStack) > 0 {
			for _, skip := range interpreter.colorConditionStack {
				if skip {
					shouldSkipOutput = true
					break
				}
			}
		}

		if !shouldSkipOutput {
			output += tokenResult
			// Count newlines in token result for more system
			interpreter.currentLine += strings.Count(tokenResult, "\n")
		}

		// If we need to flush output after processing
		if shouldFlush {
			if interpreter.session != nil {
				wish.WriteString(interpreter.session, output)
			} else {
				io.WriteString(interpreter.output, output)
			}
			output = ""

			// Handle [more] prompt if needed
			if interpreter.shouldHandleMore {
				interpreter.handleMorePrompt()
				interpreter.shouldHandleMore = false

				// Check if user quit
				if interpreter.shouldQuit {
					return output
				}
			}
		} else {
			// Check for auto-more after accumulating output (but before flushing)
			// Only check if we have enough content and more is enabled
			if interpreter.checkAutoMore() && len(output) > 0 {
				// Flush output before prompting
				if interpreter.session != nil {
					wish.WriteString(interpreter.session, output)
				} else {
					io.WriteString(interpreter.output, output)
				}
				output = ""

				// Handle more prompt
				interpreter.handleMorePrompt()

				// Check if user quit
				if interpreter.shouldQuit {
					return output
				}
			}
		}

		// Handle [quit] - exit current file
		if interpreter.shouldQuit {
			break
		}

		// Handle [exit] - exit all files
		if interpreter.shouldExit {
			// Clear call stack and exit
			interpreter.callStack = []fileContext{}
			break
		}

		// Handle [display] - stop processing current file after displaying target
		// Check this before [link] since [display] should stop immediately
		if interpreter.shouldDisplay && interpreter.displayFile != "" {
			// Process the display file
			displayOutput := interpreter.processDisplayFile(interpreter.displayFile, vars, includes)
			output += displayOutput
			interpreter.shouldDisplay = false
			interpreter.displayFile = ""
			// Stop processing current file
			break
		}

		// Handle [link] - process file and continue with current file
		if interpreter.shouldLink && interpreter.linkFile != "" {
			// Check call stack depth (max 8 levels)
			if len(interpreter.callStack) >= 8 {
				output += "[ERROR: link nesting too deep (max 8 levels)]"
				interpreter.shouldLink = false
				interpreter.linkFile = ""
			} else {
				// Save current context to call stack (before processing link)
				ctx := fileContext{
					input:      originalInput,
					vars:       vars,
					includes:   includes,
					position:   position,
					output:     output,
					style:      currentStyle,
					styleStack: make([]lipgloss.Style, len(interpreter.styleStack)),
				}
				copy(ctx.styleStack, interpreter.styleStack)
				interpreter.callStack = append(interpreter.callStack, ctx)

				// Process linked file
				linkedOutput := interpreter.processLinkFile(interpreter.linkFile, vars, includes)
				output += linkedOutput

				// Restore context from call stack (but keep the accumulated output)
				if len(interpreter.callStack) > 0 {
					ctx = interpreter.callStack[len(interpreter.callStack)-1]
					interpreter.callStack = interpreter.callStack[:len(interpreter.callStack)-1]

					originalInput = ctx.input
					vars = ctx.vars
					includes = ctx.includes
					// Don't restore position - we continue from after the [link] token
					// Don't restore output - we've accumulated the linked output
					currentStyle = ctx.style
					interpreter.styleStack = ctx.styleStack
				}

				interpreter.shouldLink = false
				interpreter.linkFile = ""
				// Continue processing from after the [link] token
				// (position will be updated at end of loop iteration)
			}
		}

		// Handle [top] - jump to top of file
		if interpreter.shouldGotoTop {
			position = 0
			interpreter.shouldGotoTop = false
			// Re-parse labels when jumping to top
			interpreter.labels = parseLabels(originalInput)
			continue
		}

		// Handle [goto] - jump to label
		if interpreter.gotoTarget != "" {
			labelPos, ok := interpreter.labels[interpreter.gotoTarget]
			if ok {
				position = labelPos
				interpreter.gotoTarget = ""
				continue
			}
			// Label not found, continue normally
			interpreter.gotoTarget = ""
		}

		// Check if we hit a newline - reset skipLine if we did
		if strings.Contains(remainingInput[:start+end+1], "\n") {
			skipLine = false
		}

		position = tokenPos + end + 1
	}

	// If we're still capturing an option at the end, save it
	if interpreter.capturingOption && interpreter.currentOptionID != "" {
		interpreter.menuOptions[interpreter.currentOptionID] = strings.TrimSpace(interpreter.optionTextBuffer.String())
		interpreter.capturingOption = false
		interpreter.currentOptionID = ""
		interpreter.optionTextBuffer.Reset()
	}

	// Handle [on exit] - execute exit file if set
	if interpreter.onExitFile != "" {
		exitOutput := interpreter.processDisplayFile(interpreter.onExitFile, vars, includes)
		output += exitOutput
		interpreter.onExitFile = ""
	}

	return output
}

// parseLabels scans the input and builds a map of label names to their positions.
// Labels are defined as [/labelname] or [label labelname].
func parseLabels(input string) map[string]int {
	labels := make(map[string]int)
	pos := 0

	for pos < len(input) {
		start := strings.Index(input[pos:], "[")
		if start == -1 {
			break
		}
		start += pos

		// Check for escaped bracket
		if start+1 < len(input) && input[start+1] == '[' {
			pos = start + 2
			continue
		}

		end := strings.Index(input[start:], "]")
		if end == -1 {
			break
		}
		end += start

		tokenContent := input[start+1 : end]
		parts := parseFieldsWithQuotes(tokenContent)

		if len(parts) > 0 {
			firstPart := strings.ToLower(parts[0])
			// Check for [/labelname] format
			if strings.HasPrefix(firstPart, "/") && len(firstPart) > 1 {
				labelName := strings.ToLower(firstPart[1:])
				labels[labelName] = end + 1 // Position after the closing bracket
			} else if firstPart == "label" && len(parts) > 1 {
				// Check for [label labelname] format
				labelName := strings.ToLower(parts[1])
				labels[labelName] = end + 1
			}
		}

		pos = end + 1
	}

	return labels
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
//   - Interactive tokens (e.g., [readln], [enter], [goto], [choice], [quit], [exit])
//   - Custom registered tokens
//   - Variable substitutions
//   - ASCII/UTF-8 code tokens (e.g., [65] for 'A', [U+00A9] for '©')
//
// The style parameter is modified in-place as tokens are processed. The function
// returns the rendered output string, a boolean indicating whether output should be
// flushed (for interactive tokens), and a boolean indicating whether to skip the
// rest of the line (for [choice]).
func (interpreter *Interpreter) processToken(content string, style *lipgloss.Style, vars map[string]any, includes []string, accumulatedOutput string, tokenPos int, originalInput string) (string, bool, bool) {
	parts := parseFieldsWithQuotes(content)
	result := ""
	shouldFlush := false
	skipLineAfter := false

	// Helper to check if we should skip output due to color conditionals
	shouldSkipOutput := func() bool {
		if len(interpreter.colorConditionStack) > 0 {
			// If any level in the stack says to skip, skip
			for _, skip := range interpreter.colorConditionStack {
				if skip {
					return true
				}
			}
		}
		return false
	}

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
		case "bell": // [bell] token: beep (ASCII 07).
			result += "\a"
			continue
		case "bs": // [bs] token: backspace (ASCII 08).
			result += "\b"
			continue
		case "tab": // [tab] token: tab character (ASCII 09).
			result += "\t"
			continue
		case "pause": // [pause] token: pause for half a second.
			time.Sleep(500 * time.Millisecond)
			continue
		case "comment": // [comment <c>] token: comment - ignore content.
			// Skip processing the rest of the token content
			// Everything after "comment" is ignored
			if i+1 < len(parts) {
				// Skip the comment text
				i = len(parts) - 1 // Skip to end
			}
			continue
		case "repeat": // [repeat <c>[<n>]] token: repeat character <c> for <n> times.
			if i+1 < len(parts) {
				char := parts[i+1]
				count := 1 // Default to 1 if no count specified

				// Check if there's a count in the next part (e.g., [repeat = 15] or [repeat]=[15])
				// The count can be specified as a number after the character
				if i+2 < len(parts) {
					if n, err := strconv.Atoi(parts[i+2]); err == nil {
						count = n
						i++
					}
				}

				// If character is a single character, use it; otherwise use first character
				if len(char) > 0 {
					// Use first character of the string (handles multi-char inputs)
					charToRepeat := char[0]
					result += strings.Repeat(string(charToRepeat), count)
				}
				i++
			}
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
		case "include": // [include] token: expects one argument.
			if i+1 < len(parts) {
				filename := parts[i+1]

				// Prevent infinite recursion.
				for _, inc := range includes {
					if inc == filename {
						result += fmt.Sprintf("[ERROR: %s included recursively]", filename)
						return result, shouldFlush, skipLineAfter
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
		case "display": // [display <f>] token: display file without returning control.
			if i+1 < len(parts) {
				filename := parts[i+1]
				// Set flag to display file and stop processing current file
				interpreter.shouldDisplay = true
				interpreter.displayFile = filename
				i++
			}
			continue
		case "link": // [link <f>] token: display file with return to caller.
			if i+1 < len(parts) {
				filename := parts[i+1]
				// Set flag to link file - actual processing happens in interpret loop
				interpreter.shouldLink = true
				interpreter.linkFile = filename
				i++
			}
			continue
		case "bg": // [BG <c>] token: set background color only.
			if i+1 < len(parts) && !shouldSkipOutput() {
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
		case "fg": // [FG <c>] token: set foreground color only.
			if i+1 < len(parts) && !shouldSkipOutput() {
				fgToken := strings.ToLower(parts[i+1])
				if strings.HasPrefix(fgToken, "#") {
					if len(fgToken) == 7 {
						*style = style.Foreground(lipgloss.Color(fgToken))
					} else {
						// ANSI color code
						if n, err := parseNumber(fgToken[1:]); err == nil {
							*style = style.Foreground(lipgloss.Color(strconv.Itoa(n)))
						}
					}
				} else if fg, ok := colorMapping[fgToken]; ok {
					*style = style.Foreground(fg)
				}
				i++
			}
			continue
		case "on": // [on exit <f>] or [on <color>] token.
			if i+1 < len(parts) {
				nextPart := strings.ToLower(parts[i+1])
				if nextPart == "exit" && i+2 < len(parts) {
					// [on exit <f>] - set exit file handler
					filename := parts[i+2]
					interpreter.onExitFile = filename
					i += 2
				} else {
					// [on <color>] - existing background color logic
					if !shouldSkipOutput() {
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
					}
					i++
				}
			}
			continue
		case "onexit": // [onexit <f>] token: synonym for [on exit <f>].
			if i+1 < len(parts) {
				filename := parts[i+1]
				interpreter.onExitFile = filename
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
		case "readln": // [readln] or [readln <desc>] token: read a line of input from user.
			if interpreter.reader == nil {
				result += style.Render("[ERROR: no reader configured, use WithReader() option]")
				continue
			}

			shouldFlush = true

			// Read a line from the reader
			scanner := bufio.NewScanner(interpreter.reader)
			if scanner.Scan() {
				interpreter.readlnResponse = scanner.Text()

				// Store in questionnaire data if description provided
				if i+1 < len(parts) {
					desc := parts[i+1]
					interpreter.questionnaireData = append(interpreter.questionnaireData, fmt.Sprintf("%s: %s", desc, interpreter.readlnResponse))
					i++
				} else {
					interpreter.questionnaireData = append(interpreter.questionnaireData, interpreter.readlnResponse)
				}
			} else {
				interpreter.readlnResponse = ""
				if !interpreter.answerOptional {
					// If answer required and nothing entered, add empty entry
					if i+1 < len(parts) {
						desc := parts[i+1]
						interpreter.questionnaireData = append(interpreter.questionnaireData, fmt.Sprintf("%s: ", desc))
						i++
					}
				}
			}
			continue
		case "enter": // [enter] token: wait for Enter key press.
			if interpreter.reader == nil {
				result += style.Render("[ERROR: no reader configured, use WithReader() option]")
				continue
			}

			shouldFlush = true

			// Display prompt and wait for Enter
			prompt := style.Render("Press ENTER to continue")
			result += prompt

			// Read until newline
			buf := make([]byte, 1)
			for {
				n, err := interpreter.reader.Read(buf)
				if err != nil || n == 0 {
					break
				}
				if buf[0] == '\n' || buf[0] == '\r' {
					// Handle \r\n sequence
					if buf[0] == '\r' {
						// Peek at next byte
						peek := make([]byte, 1)
						if n2, _ := interpreter.reader.Read(peek); n2 > 0 && peek[0] != '\n' {
							// Put it back somehow? Actually we can't unread, so just continue
						}
					}
					break
				}
			}
			continue
		case "choice": // [choice <c>] token: conditional - display rest of line only if response matches.
			if i+1 < len(parts) {
				expectedChar := strings.ToLower(parts[i+1])
				// Check if menu response or readln response matches
				response := ""
				if interpreter.menuResponse != "" {
					response = interpreter.menuResponse
				} else if interpreter.readlnResponse != "" {
					// For readln, check if first character matches (case-insensitive)
					if len(interpreter.readlnResponse) > 0 {
						response = strings.ToLower(string(interpreter.readlnResponse[0]))
					}
				}

				if strings.ToLower(response) != expectedChar {
					// Condition not met - skip rest of line
					skipLineAfter = true
				}
				i++
			}
			continue
		case "ifentered": // [ifentered <s>] token: conditional - display rest of line only if [readln] response matches exactly.
			if i+1 < len(parts) {
				expectedString := strings.ToLower(parts[i+1])
				// Check if readln response matches exactly (case-insensitive)
				response := ""
				if interpreter.readlnResponse != "" {
					response = strings.ToLower(interpreter.readlnResponse)
				}

				if response != expectedString {
					// Condition not met - skip rest of line
					skipLineAfter = true
				}
				i++
			}
			continue
		case "top": // [top] token: jump to top of current file.
			interpreter.shouldGotoTop = true
			continue
		case "goto": // [goto <label>] token: jump to label.
			if i+1 < len(parts) {
				labelName := strings.ToLower(parts[i+1])
				interpreter.gotoTarget = labelName
				i++
			}
			continue
		case "jump": // [jump <label>] token: synonym for [goto].
			if i+1 < len(parts) {
				labelName := strings.ToLower(parts[i+1])
				interpreter.gotoTarget = labelName
				i++
			}
			continue
		case "quit": // [quit] token: exit current file.
			interpreter.shouldQuit = true
			continue
		case "exit": // [exit] token: exit all files.
			interpreter.shouldExit = true
			continue
		case "label": // [label <name>] token: define a label (position marker).
			// Labels are already parsed by parseLabels(), so we just ignore this token
			// during processing (it doesn't output anything)
			if i+1 < len(parts) {
				i++
			}
			continue
		case "ansopt": // [ansopt] token: make answers optional.
			interpreter.answerOptional = true
			continue
		case "ansreq": // [ansreq] token: make answers required.
			interpreter.answerOptional = false
			continue
		case "store": // [store] or [store <desc>] token: store menu response in questionnaire.
			if interpreter.menuResponse != "" {
				if i+1 < len(parts) {
					desc := parts[i+1]
					interpreter.questionnaireData = append(interpreter.questionnaireData, fmt.Sprintf("%s: %s", desc, interpreter.menuResponse))
					i++
				} else {
					interpreter.questionnaireData = append(interpreter.questionnaireData, interpreter.menuResponse)
				}
			}
			continue
		case "write": // [write <text>] token: write line to questionnaire data.
			if i+1 < len(parts) {
				text := parts[i+1]
				interpreter.questionnaireData = append(interpreter.questionnaireData, text)
				i++
			}
			continue
		case "color": // [color] token: display following text only if ANSI enabled.
		case "colour": // [colour] token: Canadian spelling of [color].
			// Check if ANSI is supported
			hasColor := interpreter.output.ColorProfile() != termenv.Ascii
			// Push false if color is NOT supported (skip output), true if color IS supported (show output)
			interpreter.colorConditionStack = append(interpreter.colorConditionStack, !hasColor)
			continue
		case "nocolor": // [nocolor] token: display following text only if ANSI disabled.
		case "nocolour": // [nocolour] token: Canadian spelling of [nocolor].
			// Check if ANSI is supported
			hasColor := interpreter.output.ColorProfile() != termenv.Ascii
			// Push true if color IS supported (skip output), false if color is NOT supported (show output)
			interpreter.colorConditionStack = append(interpreter.colorConditionStack, hasColor)
			continue
		case "endcolor": // [endcolor] token: end color-conditional block.
		case "endcolour": // [endcolour] token: Canadian spelling of [endcolor].
			// Pop the color condition stack
			if len(interpreter.colorConditionStack) > 0 {
				interpreter.colorConditionStack = interpreter.colorConditionStack[:len(interpreter.colorConditionStack)-1]
			}
			continue
		case "copy": // [copy <f>] token: copy file contents directly to output (no parsing).
			if i+1 < len(parts) && !shouldSkipOutput() {
				filename := parts[i+1]
				data := path.Join(interpreter.templateRoot, filename)

				if fileData, err := os.ReadFile(data); err == nil {
					result += string(fileData)
				} else {
					result += fmt.Sprintf("[ERROR: %v]", err)
				}
				i++
			}
			continue
		case "more": // [more] token: display "More [Y,n,=]?" prompt.
			if interpreter.reader == nil {
				result += style.Render("[ERROR: no reader configured, use WithReader() option]")
				continue
			}

			shouldFlush = true
			// Note: handleMorePrompt will be called after output is flushed
			// Store that we need to handle more prompt
			interpreter.shouldHandleMore = true
			continue
		case "moreon": // [moreon] token: enable automatic more prompts.
			interpreter.moreEnabled = true
			continue
		case "moreoff": // [moreoff] token: disable automatic more prompts.
			interpreter.moreEnabled = false
			continue
		}

		// Check for [/labelname] format (labels starting with /)
		if strings.HasPrefix(part, "/") && len(part) > 1 {
			// This is a label definition, already parsed by parseLabels()
			// Just ignore it during processing
			continue
		}

		// Continue with other token processing
		// Colors can be specified one of three ways:
		// 1. By name (e.g., red, green, blue)
		// 2. By hex code (e.g., #ff0000, #00ff00, #0000ff)
		// 3. By ANSI color code, as a number (e.g., #63)
		if isColorToken(part) && !shouldSkipOutput() {
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
				if !shouldSkipOutput() {
					result += style.Render(text)
				}
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
				if !shouldSkipOutput() {
					result += style.Render(text)
				}
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
				if !shouldSkipOutput() {
					result += style.Render(text)
				}
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
				if !shouldSkipOutput() {
					result += style.Render(text)
				}
				i += token.ArgCount
			} else {
				text = token.Func([]string{})
				if !shouldSkipOutput() {
					result += style.Render(text)
				}
			}
			// If capturing option text, capture the plain text
			if interpreter.capturingOption {
				interpreter.optionTextBuffer.WriteString(text)
			}
			continue
		}
		// If token is unrecognized, emit an error message inline.
		if !shouldSkipOutput() {
			result += style.Render(fmt.Sprintf("[UNRECOGNIZED TOKEN \"%s\"]", part))
		}
	}
	return result, shouldFlush, skipLineAfter
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

// processDisplayFile loads and processes a file for [display] token.
// This is used by both [display] and [on exit] handlers.
func (interpreter *Interpreter) processDisplayFile(filename string, vars map[string]any, includes []string) string {
	templatePath := path.Join(interpreter.templateRoot, filename)
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Sprintf("[ERROR: %v]", err)
	}

	// Prevent infinite recursion
	for _, inc := range includes {
		if inc == filename {
			return fmt.Sprintf("[ERROR: %s included recursively]", filename)
		}
	}

	// Add to includes chain
	newIncludes := append(includes, filename)

	return interpreter.interpret(string(data), vars, newIncludes)
}

// processLinkFile loads and processes a file for [link] token.
// The file is processed and control returns to the caller.
func (interpreter *Interpreter) processLinkFile(filename string, vars map[string]any, includes []string) string {
	return interpreter.processDisplayFile(filename, vars, includes)
}

// getTerminalHeight attempts to get the terminal height in lines.
// Returns 24 as default if unable to determine.
func (interpreter *Interpreter) getTerminalHeight() int {
	// Try to get terminal size from termenv
	if interpreter.output != nil {
		// termenv doesn't have direct height access, but we can check environment
		// For SSH sessions, we can try to get from session
		if interpreter.session != nil {
			sshPty, _, _ := interpreter.session.Pty()
			if sshPty.Window.Height > 0 {
				return int(sshPty.Window.Height)
			}
		}
	}
	// Default to 24 lines (standard terminal size)
	return 24
}

// checkAutoMore checks if we should automatically prompt for more when enabled.
// Returns true if we should prompt, false otherwise.
func (interpreter *Interpreter) checkAutoMore() bool {
	if !interpreter.moreEnabled {
		return false
	}

	// Prompt if we're within 2 lines of the bottom
	// (leave room for the prompt itself)
	if interpreter.terminalHeight > 0 && interpreter.currentLine >= interpreter.terminalHeight-2 {
		// Don't prompt again on the same line
		if interpreter.currentLine > interpreter.lastMoreLine {
			return true
		}
	}
	return false
}

// handleMorePrompt displays the "More [Y,n,=]?" prompt and waits for user input.
// Returns the response character (Y, n, or =) or empty string.
func (interpreter *Interpreter) handleMorePrompt() string {
	if interpreter.reader == nil {
		return ""
	}

	// Display prompt
	prompt := "More [Y,n,=]? "
	if interpreter.session != nil {
		wish.WriteString(interpreter.session, prompt)
	} else {
		io.WriteString(interpreter.output, prompt)
	}

	// Read a single character
	buf := make([]byte, 1)
	n, err := interpreter.reader.Read(buf)
	if err != nil || n == 0 {
		return ""
	}

	inputChar := strings.ToLower(string(buf[0]))

	// Handle valid responses
	if inputChar == "y" || inputChar == "n" || inputChar == "=" {
		interpreter.moreResponse = inputChar
		interpreter.lastMoreLine = interpreter.currentLine

		// Handle responses:
		// Y = continue (clear screen and reset line counter)
		// n = quit/stop
		// = = show one more line
		if inputChar == "y" {
			// Clear screen and reset
			if interpreter.session != nil {
				wish.WriteString(interpreter.session, ansi.EraseDisplay(2)+ansi.CursorPosition(1, 1))
			} else {
				io.WriteString(interpreter.output, ansi.EraseDisplay(2)+ansi.CursorPosition(1, 1))
			}
			interpreter.currentLine = 0
			interpreter.lastMoreLine = 0
		} else if inputChar == "n" {
			// Quit - set flag to stop processing
			interpreter.shouldQuit = true
		}
		// = just continues, no special action needed

		return inputChar
	}

	// Invalid input, treat as 'n' (quit)
	interpreter.moreResponse = "n"
	interpreter.shouldQuit = true
	return "n"
}
