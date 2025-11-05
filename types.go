package mecca

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/ssh"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TokenFunc defines a function type for processing a MECCA token.
// When a token is encountered in a MECCA template, the registered TokenFunc
// is called with the token's arguments and should return the string that will
// replace the token in the output.
//
// The args parameter contains the arguments passed to the token. For example,
// in [repeat 3 hello], the args would be ["3", "hello"] (or ["3", "hello"] if
// the token takes 2 arguments).
//
// Example:
//
//	interpreter.RegisterToken("greet", func(args []string) string {
//		if len(args) > 0 {
//			return "Hello, " + args[0] + "!"
//		}
//		return "Hello!"
//	}, 1)
type TokenFunc func([]string) string

// Token holds a processing function and metadata for a registered MECCA token.
// Tokens can be registered with the interpreter to provide custom functionality
// beyond the built-in tokens.
type Token struct {
	// Func is the function that processes the token and returns the substitution string.
	Func TokenFunc
	// ArgCount is the number of expected arguments that the token takes.
	// If a token is called with fewer arguments, it will receive an empty slice.
	ArgCount int
}

// Interpreter is a MECCA language interpreter that processes MECCA templates
// and renders them to terminal output. Each interpreter maintains its own
// template root directory, output configuration, and token registry.
type Interpreter struct {
	templateRoot     string             // Base directory for template file resolution
	session          ssh.Session        // SSH session if rendering to remote terminal
	renderer         *lipgloss.Renderer // Lipgloss renderer for styling
	output           *termenv.Output    // Terminal output for rendering
	reader           io.Reader          // Reader for interactive input (menus, etc.)
	tokenRegistry    map[string]Token   // Registry of custom registered tokens
	styleStack       []lipgloss.Style   // Stack for [save]/[load] style functionality
	menuOptions      map[string]string  // Map of option_id -> option text for current menu
	menuResponse     string             // The selected option_id from the menu
	inMenu           bool               // Whether we're currently building a menu
	capturingOption  bool               // Whether we're currently capturing option text
	currentOptionID  string             // The option ID being captured
	optionTextBuffer strings.Builder    // Buffer for the option text being captured

	// Interactive input and questionnaire data
	readlnResponse    string   // Last response from [readln]
	questionnaireData []string // Collected questionnaire responses
	answerOptional    bool     // Whether answers are optional ([ansopt] / [ansreq])

	// Flow control
	labels        map[string]int // Map of label names to positions in input
	shouldQuit    bool           // [quit] flag - exit current file
	shouldExit    bool           // [exit] flag - exit all files
	gotoTarget    string         // Target label for [goto]
	shouldGotoTop bool           // [top] flag - jump to top of file

	// File chaining
	callStack     []fileContext // Call stack for [link] (nested file execution)
	onExitFile    string        // File to execute on exit ([on exit] / [onexit])
	shouldDisplay bool          // [display] flag - stop processing current file
	displayFile   string        // File to display and then stop
	linkFile      string        // File to link (process and return)
	shouldLink    bool          // [link] flag - process file and return

	// Conditional display
	colorConditionStack []bool // Stack for [color]/[nocolor] nesting (true = show color, false = skip)

	// More system
	moreEnabled      bool   // Whether automatic more prompts are enabled
	currentLine      int    // Current line position in terminal
	terminalHeight   int    // Terminal height (lines), 0 = unknown
	moreResponse     string // Last response from [more] prompt (Y/n/=)
	lastMoreLine     int    // Line number where last more prompt was shown
	shouldHandleMore bool   // Flag to handle [more] prompt after flushing
}

// fileContext holds the state of a file being processed, used for [link] call stack
type fileContext struct {
	input      string           // Original input string
	vars       map[string]any   // Variables passed to this file
	includes   []string         // Include chain for recursion detection
	position   int              // Current position in input
	output     string           // Accumulated output so far
	style      lipgloss.Style   // Current style state
	styleStack []lipgloss.Style // Style stack state
}

// sshOutput bridges an SSH session with termenv's output interface, allowing
// the interpreter to query terminal capabilities for remote SSH sessions.
// This is an internal type used to adapt SSH sessions to termenv's requirements.
type sshOutput struct {
	ssh.Session
	tty *os.File // Local TTY for terminal capability queries
}

// Write writes data to the SSH session.
func (s *sshOutput) Write(p []byte) (int, error) {
	return s.Session.Write(p)
}

// Read reads data from the SSH session.
func (s *sshOutput) Read(p []byte) (int, error) {
	return s.Session.Read(p)
}

// Fd returns the file descriptor of the local TTY, required for termenv.
func (s *sshOutput) Fd() uintptr {
	return s.tty.Fd()
}

// sshEnviron provides environment variable access for SSH sessions.
// This is an internal type that adapts SSH session environment variables
// to termenv's environment interface.
type sshEnviron struct {
	environ []string // Environment variables as key=value pairs
}

// Getenv retrieves the value of an environment variable by key.
func (s *sshEnviron) Getenv(key string) string {
	for _, v := range s.environ {
		if strings.HasPrefix(v, key+"=") {
			return v[len(key)+1:]
		}
	}
	return ""
}

// Environ returns all environment variables as key=value pairs.
func (s *sshEnviron) Environ() []string {
	return s.environ
}
