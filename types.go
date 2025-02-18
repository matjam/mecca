package mecca

import (
	"os"
	"strings"

	"github.com/charmbracelet/ssh"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TokenFunc defines a function for processing a MECCA token.
// It receives a slice of string arguments and returns the substitution string.
type TokenFunc func([]string) string

// Token holds a processing function and the expected number of arguments.
// It is used to encapsulate a token's functionality.
type Token struct {
	Func     TokenFunc // Func is the function that processes the token.
	ArgCount int       // ArgCount is the number of expected arguments.
}

// Interpreter is a MECCA interpreter that processes MECCA templates.
type Interpreter struct {
	templateRoot  string
	session       ssh.Session
	renderer      *lipgloss.Renderer
	output        *termenv.Output
	tokenRegistry map[string]Token
}

// Bridge Wish and Termenv so we can query for a user's terminal capabilities.
type sshOutput struct {
	ssh.Session
	tty *os.File
}

func (s *sshOutput) Write(p []byte) (int, error) {
	return s.Session.Write(p)
}

func (s *sshOutput) Read(p []byte) (int, error) {
	return s.Session.Read(p)
}

func (s *sshOutput) Fd() uintptr {
	return s.tty.Fd()
}

type sshEnviron struct {
	environ []string
}

func (s *sshEnviron) Getenv(key string) string {
	for _, v := range s.environ {
		if strings.HasPrefix(v, key+"=") {
			return v[len(key)+1:]
		}
	}
	return ""
}

func (s *sshEnviron) Environ() []string {
	return s.environ
}
