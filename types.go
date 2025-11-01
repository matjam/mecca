package mecca

import (
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
	templateRoot  string              // Base directory for template file resolution
	session       ssh.Session         // SSH session if rendering to remote terminal
	renderer      *lipgloss.Renderer  // Lipgloss renderer for styling
	output        *termenv.Output     // Terminal output for rendering
	tokenRegistry map[string]Token    // Registry of custom registered tokens
	styleStack    []lipgloss.Style    // Stack for [save]/[load] style functionality
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
