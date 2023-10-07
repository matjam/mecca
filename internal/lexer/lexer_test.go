package lexer

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {
	input := "Hello, this is [command]text mode again."
	r := strings.NewReader(input)
	l := NewLexer(r)

	expectedTokens := []Token{
		{Type: TOKEN_TEXT, Value: "Hello, this is ", Line: 1, Column: 1},
		{Type: TOKEN_COMMAND_START, Value: "[", Line: 1, Column: 16},
		{Type: TOKEN_COMMAND_TEXT, Value: "command", Line: 1, Column: 17},
		{Type: TOKEN_COMMAND_END, Value: "]", Line: 1, Column: 24},
		{Type: TOKEN_TEXT, Value: "text mode again.", Line: 1, Column: 25},
		{Type: TOKEN_EOF, Value: "", Line: 1, Column: 41},
	}

	for _, expectedToken := range expectedTokens {
		token := l.NextToken()
		fmt.Printf("Got token %#v\n", token)
		if token != expectedToken {
			t.Errorf("Expected token %#v, got %#v", expectedToken, token)
		}
	}
}

func TestLexerWithBracket(t *testing.T) {
	input := "Hello, this is [[ text with an escaped square bracket."
	r := strings.NewReader(input)
	l := NewLexer(r)

	expectedTokens := []Token{
		{Type: TOKEN_TEXT, Value: "Hello, this is [ text with an escaped square bracket.", Line: 1, Column: 1},
	}

	for _, expectedToken := range expectedTokens {
		token := l.NextToken()
		fmt.Printf("Got token %#v\n", token)
		if token != expectedToken {
			t.Errorf("Expected token %#v, got %#v", expectedToken, token)
		}
	}
}

func TestLexerWithCommandAtStart(t *testing.T) {
	input := "[command]text"
	r := strings.NewReader(input)
	l := NewLexer(r)

	expectedTokens := []Token{
		{Type: TOKEN_COMMAND_START, Value: "[", Line: 1, Column: 1},
		{Type: TOKEN_COMMAND_TEXT, Value: "command", Line: 1, Column: 2},
		{Type: TOKEN_COMMAND_END, Value: "]", Line: 1, Column: 9},
		{Type: TOKEN_TEXT, Value: "text", Line: 1, Column: 10},
		{Type: TOKEN_EOF, Value: "", Line: 1, Column: 14},
	}

	for _, expectedToken := range expectedTokens {
		token := l.NextToken()
		fmt.Printf("Got token %#v\n", token)
		if token != expectedToken {
			t.Errorf("Expected token %#v, got %#v", expectedToken, token)
		}
	}
}

func TestLexerNoError(t *testing.T) {
	input, err := os.Open("testdata/test.mec")
	if err != nil {
		t.Errorf("Error opening test file: %s", err)
	}
	l := NewLexer(input)

	for {
		token := l.NextToken()
		fmt.Printf("Got token %v\n", token)
		if token.Type == TOKEN_EOF {
			break
		}
	}
}
