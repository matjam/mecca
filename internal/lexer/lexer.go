package lexer

import (
	"bufio"
	"fmt"
	"io"
)

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("%v (%v:%v): \"%v\"", t.Type, t.Line, t.Column, t.Value)
}

type Lexer struct {
	input        *bufio.Reader
	mode         int
	line         int
	column       int
	currentToken *Token // the token that is currently being lexed
	nextToken    *Token // a token that should be returned on the next call to NextToken()
}

const (
	TEXT_MODE = iota
	COMMAND_MODE
)

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		input:  bufio.NewReader(r),
		mode:   TEXT_MODE,
		line:   1,
		column: 1,
	}
}

func (l *Lexer) emitCurrentToken() Token {
	token := l.currentToken
	l.currentToken = nil
	return *token
}

func (l *Lexer) emitNextToken() Token {
	token := l.nextToken
	l.nextToken = nil
	return *token
}

func (l *Lexer) NextToken() Token {
	var err error
	var finishedToken Token
	var currentRune rune

	// return a pending token that was already lexed
	if l.nextToken != nil {
		return l.emitNextToken()
	}

	// read the runes from the input until we find the beginning of the next token.
	// We then set the currentToken to the new token and return the token that was
	// previously being lexed.
	for {
		currentRune, _, err = l.input.ReadRune()
		if err != nil {
			if l.currentToken == nil {
				return Token{Type: TOKEN_EOF, Value: "", Line: l.line, Column: l.column}
			}

			l.nextToken = &Token{Type: TOKEN_EOF, Value: "", Line: l.line, Column: l.column}
			return l.emitCurrentToken()
		}

		currentLine, currentColumn := l.line, l.column

		// increment our line/column counters
		if currentRune == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}

		// if we're in text mode, we need to check if we've found the beginning of a command
		if l.mode == TEXT_MODE {
			if currentRune == '[' {
				// peek ahead to see if the next character is also a '['
				peekedRune, _, err := l.input.ReadRune()
				if err != nil {
					l.nextToken = &Token{Type: TOKEN_EOF, Value: "", Line: l.line, Column: l.column}
					l.currentToken.Value += string(currentRune)
					return l.emitCurrentToken()
				}
				if peekedRune == '[' {
					// we've found an escaped square bracket, so we need to add it to the current token
					if l.currentToken == nil {
						l.currentToken = &Token{Type: TOKEN_TEXT, Value: "", Line: currentLine, Column: currentColumn}
					}

					l.currentToken.Value += string(currentRune)

					continue
				}
				// unread the peeked rune so that it can be read again on the next call to NextToken()
				err = l.input.UnreadRune()
				if err != nil {
					panic(err)
				}

				l.mode = COMMAND_MODE
				if l.currentToken != nil {
					finishedToken = *l.currentToken
					l.currentToken = nil
					l.nextToken = &Token{Type: TOKEN_COMMAND_START, Value: "[", Line: currentLine, Column: currentColumn}
					return finishedToken
				}
				return Token{Type: TOKEN_COMMAND_START, Value: "[", Line: currentLine, Column: currentColumn}
			}

			if l.currentToken == nil {
				l.currentToken = &Token{Type: TOKEN_TEXT, Value: "", Line: currentLine, Column: currentColumn}
			}

			l.currentToken.Value += string(currentRune)
		} else if l.mode == COMMAND_MODE {
			if currentRune == ']' {
				l.mode = TEXT_MODE
				if l.currentToken != nil {
					finishedToken = *l.currentToken
					l.currentToken = nil
					l.nextToken = &Token{Type: TOKEN_COMMAND_END, Value: "]", Line: currentLine, Column: currentColumn}
					return finishedToken
				}
				return Token{Type: TOKEN_COMMAND_END, Value: "]", Line: currentLine, Column: currentColumn}
			}

			// So we don't actually know what kind of token we are parsing right now, so
			// we need to

			if l.currentToken == nil {
				l.currentToken = &Token{Type: TOKEN_COMMAND_TEXT, Value: "", Line: currentLine, Column: currentColumn}
			}

			l.currentToken.Value += string(currentRune)
		} else {
			panic("Unknown lexer mode")
		}
	}
}

func (l *Lexer) processCommandText() Token {
	// process input until we find the end of the command, and return a token specific to the command.

	var buffer string
	currentLine, currentColumn := l.line, l.column

	for {
		r, _, err := l.input.ReadRune()
		if err != nil {
			l.nextToken = &Token{Type: TOKEN_EOF, Value: "", Line: l.line, Column: l.column}
			return Token{Type: TOKEN_COMMAND_TEXT, Value: buffer, Line: currentLine, Column: currentColumn}
		}

		// increment our line/column counters
		if r == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}

		if r == ']' || r == ' ' {
			tokenType, err := TokenFromString(buffer)
			if err != nil {
				return Token{Type: TOKEN_COMMAND_TEXT, Value: buffer, Line: currentLine, Column: currentColumn}
			}

			return Token{Type: tokenType, Value: buffer, Line: currentLine, Column: currentColumn}
		}

		buffer += string(r)
	}
}
