package mecca

import (
	"strings"
)

// TokenFunc defines a function for processing a MECCA token.
type TokenFunc func([]string) string

// Token holds the processing function and the expected number of arguments.
type Token struct {
	Func     TokenFunc
	ArgCount int
}

// tokenRegistry stores all registered tokens.
var tokenRegistry = make(map[string]Token)

// RegisterToken registers a token by name, its processing function, and the number of arguments it takes.
func RegisterToken(name string, fn TokenFunc, argCount int) {
	tokenRegistry[strings.ToLower(name)] = Token{
		Func:     fn,
		ArgCount: argCount,
	}
}

// GetToken retrieves a token by name.
func GetToken(name string) (Token, bool) {
	token, ok := tokenRegistry[strings.ToLower(name)]
	return token, ok
}
