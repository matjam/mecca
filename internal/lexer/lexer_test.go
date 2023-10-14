package lexer

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {
	//        123456789012345678901234567890123456789012345678901234567890
	input := `Hello, this is [command]text mode [command arg1 arg2] again.`
	r := strings.NewReader(input)
	l := NewLexer(r)

	expectedTokens := []Token{
		{Type: TOKEN_TEXT, Value: "Hello, this is ", Line: 1, Column: 1},
		{Type: TOKEN_COMMAND_START, Value: "[", Line: 1, Column: 16},
		{Type: TOKEN_COMMAND, Value: "command", Line: 1, Column: 17},
		{Type: TOKEN_COMMAND_END, Value: "]", Line: 1, Column: 24},
		{Type: TOKEN_TEXT, Value: "text mode ", Line: 1, Column: 25},
		{Type: TOKEN_COMMAND_START, Value: "[", Line: 1, Column: 35},
		{Type: TOKEN_COMMAND, Value: "command", Line: 1, Column: 36},
		{Type: TOKEN_COMMAND_ARG, Value: "arg1", Line: 1, Column: 44},
		{Type: TOKEN_COMMAND_ARG, Value: "arg2", Line: 1, Column: 49},
		{Type: TOKEN_COMMAND_END, Value: "]", Line: 1, Column: 53},
		{Type: TOKEN_TEXT, Value: " again.", Line: 1, Column: 54},
		{Type: TOKEN_EOF, Value: "EOF", Line: 1, Column: 61},
	}

	for _, expectedToken := range expectedTokens {
		token, err := l.Lex()
		if err != nil && err != io.EOF {
			t.Errorf("Error lexing: %s", err)
		}

		fmt.Printf("%v\n", token)

		if token != expectedToken {
			t.Errorf("Expected %v got %v\n", expectedToken, token)
		}
	}
}

func TestLexerUnexpectedNewline(t *testing.T) {
	//		  123456789012345678901234567890123456789012345678901234567890
	input := "Hello a [command\nunexpected newline] whoops"
	r := strings.NewReader(input)
	l := NewLexer(r)

	type result struct {
		token Token
		err   bool
	}

	expectedResults := []result{
		{Token{Type: TOKEN_TEXT, Value: "Hello a ", Line: 1, Column: 1}, false},
		{Token{Type: TOKEN_COMMAND_START, Value: "[", Line: 1, Column: 9}, false},
		{Token{Type: TOKEN_TEXT, Value: "command", Line: 1, Column: 10}, false},
		{Token{Type: TOKEN_TEXT, Value: "", Line: 0, Column: 0}, true},
	}

	for _, expectedResult := range expectedResults {
		token, err := l.Lex()
		if err != nil && expectedResult.err == false {
			t.Errorf("Error lexing: %s", err)
		}
		if expectedResult.err == false && token != expectedResult.token {
			t.Errorf("Expected %v got %v\n", expectedResult.token, token)
		}
		if expectedResult.err == true && err == nil {
			t.Errorf("Expected error, got %v\n", token)
		}
	}
}

func TestLexerDoubleSquareBracket(t *testing.T) {
	// 	  	  123456789012345678901234567890123456789012345678901234567890
	input := "Hello a [[string with double square brackets whoops"
	r := strings.NewReader(input)
	l := NewLexer(r)

	expectedTokens := []Token{
		{Type: TOKEN_TEXT, Value: "Hello a [string with double square brackets whoops", Line: 1, Column: 1},
		{Type: TOKEN_EOF, Value: "EOF", Line: 1, Column: 52},
	}

	for _, expectedToken := range expectedTokens {
		token, err := l.Lex()
		if err != nil && err != io.EOF {
			t.Errorf("Error lexing: %s", err)
		}

		if token != expectedToken {
			t.Errorf("Expected %v got %v\n", expectedToken, token)
		}
	}
}

func TestLexerStartingWithCommand(t *testing.T) {
	// 	  	  123456789012345678901234567890123456789012345678901234567890
	input := "[command]A string starting with a command"
	r := strings.NewReader(input)
	l := NewLexer(r)

	expectedTokens := []Token{
		{Type: TOKEN_COMMAND_START, Value: "[", Line: 1, Column: 1},
		{Type: TOKEN_COMMAND, Value: "command", Line: 1, Column: 2},
		{Type: TOKEN_COMMAND_END, Value: "]", Line: 1, Column: 9},
		{Type: TOKEN_TEXT, Value: "A string starting with a command", Line: 1, Column: 10},
		{Type: TOKEN_EOF, Value: "EOF", Line: 1, Column: 42},
	}

	for _, expectedToken := range expectedTokens {
		token, err := l.Lex()
		if err != nil && err != io.EOF {
			t.Errorf("Error lexing: %s", err)
		}

		fmt.Printf("%v\n", token)

		if token != expectedToken {
			t.Errorf("Expected %v got %v\n", expectedToken, token)
		}
	}
}

func TestLexerStartingWithEscapedSquareBracket(t *testing.T) {
	// 	  	  123456789012345678901234567890123456789012345678901234567890
	input := "[[A string starting with a square bracket"
	r := strings.NewReader(input)
	l := NewLexer(r)

	expected := []Token{
		{Type: TOKEN_TEXT, Value: "[A string starting with a square bracket", Line: 1, Column: 1},
		{Type: TOKEN_EOF, Value: "EOF", Line: 1, Column: 42},
	}

	for _, expectedToken := range expected {
		token, err := l.Lex()
		if err != nil && err != io.EOF {
			t.Errorf("Error lexing: %s", err)
		}

		fmt.Printf("%v\n", token)

		if token != expectedToken {
			t.Errorf("Expected %v got %v\n", expectedToken, token)
		}
	}
}

func TestLexerTextEndingWithSingleBracket(t *testing.T) {
	// 	  	  123456789012345678901234567890123456789012345678901234567890
	input := "A string ending with a single bracket["
	r := strings.NewReader(input)
	l := NewLexer(r)

	expected := []Token{
		{Type: TOKEN_TEXT, Value: "A string ending with a single bracket", Line: 1, Column: 1},
		{Type: TOKEN_COMMAND_START, Value: "[", Line: 1, Column: 38},
		{Type: TOKEN_EOF, Value: "EOF", Line: 1, Column: 39},
	}

	for _, expectedToken := range expected {
		token, err := l.Lex()
		if err != nil && err != io.EOF {
			t.Errorf("Error lexing: %s", err)
			break
		}

		fmt.Printf("%v\n", token)

		if token != expectedToken {
			t.Errorf("Expected %v got %v\n", expectedToken, token)
		}
	}
}

func TestLexerMultiline(t *testing.T) {
	// 	  	  123456789012345678901234567890123456789012345678901234567890
	input := `A string
on multiple
lines
[command]`
	r := strings.NewReader(input)
	l := NewLexer(r)

	expected := []Token{
		{Type: TOKEN_TEXT, Value: "A string", Line: 1, Column: 1},
		{Type: TOKEN_NL, Value: "\n", Line: 1, Column: 9},
		{Type: TOKEN_TEXT, Value: "on multiple", Line: 2, Column: 1},
		{Type: TOKEN_NL, Value: "\n", Line: 2, Column: 12},
		{Type: TOKEN_TEXT, Value: "lines", Line: 3, Column: 1},
 		{Type: TOKEN_NL, Value: "\n", Line: 3, Column: 6},
		{Type: TOKEN_COMMAND_START, Value: "[", Line: 4, Column: 1},
		{Type: TOKEN_COMMAND, Value: "command", Line: 4, Column: 2},
		{Type: TOKEN_COMMAND_END, Value: "]", Line: 4, Column: 9},
		{Type: TOKEN_EOF, Value: "EOF", Line: 4, Column: 10},
	}
	
	for _, expectedToken := range expected {
		token, err := l.Lex()
		if err != nil && err != io.EOF {
			t.Errorf("Error lexing: %s", err)
			break
		}

		fmt.Printf("%v\n", token)

		if token != expectedToken {
			t.Errorf("Expected %v got %v\n", expectedToken, token)
		}
	}
}