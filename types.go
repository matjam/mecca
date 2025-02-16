package mecca

import (
	"strings"
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

// tokenRegistry stores all registered tokens.
var tokenRegistry = make(map[string]Token)

// RegisterToken registers a token by name, its processing function, and the number of arguments it takes.
// The token name is case-insensitive.
func RegisterToken(name string, fn TokenFunc, argCount int) {
	tokenRegistry[strings.ToLower(name)] = Token{
		Func:     fn,
		ArgCount: argCount,
	}
}

// GetToken retrieves a registered token by name.
// It returns the Token and a boolean indicating whether it was found.
func GetToken(name string) (Token, bool) {
	token, ok := tokenRegistry[strings.ToLower(name)]
	return token, ok
}
