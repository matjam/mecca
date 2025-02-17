package mecca

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

type Interpreter struct {
	writer        io.Writer
	renderer      *lipgloss.Renderer
	tokenRegistry map[string]Token
}

func NewInterpreter(w io.Writer) *Interpreter {
	return &Interpreter{
		writer:        w,
		renderer:      lipgloss.NewRenderer(w),
		tokenRegistry: make(map[string]Token),
	}
}

func (i *Interpreter) RegisterToken(name string, fn TokenFunc, argCount int) {
	i.tokenRegistry[strings.ToLower(name)] = Token{
		Func:     fn,
		ArgCount: argCount,
	}
}

func (i *Interpreter) GetToken(name string) (Token, bool) {
	token, ok := i.tokenRegistry[strings.ToLower(name)]
	return token, ok
}
