package lexer

import (
	"errors"
	"strings"
)

//go:generate stringer -type TokenType -trimprefix TOKEN_
type TokenType int

const (
	TOKEN_TEXT TokenType = iota
	TOKEN_NL
	TOKEN_COMMAND_START
	TOKEN_COMMAND_END
	TOKEN_COMMAND_TEXT
	TOKEN_FUNC
	TOKEN_START_LOOP
	TOKEN_END_LOOP
	TOKEN_START_COND
	TOKEN_END_COND
	TOKEN_EOF
	TOKEN_RESET
	TOKEN_BOLD
	TOKEN_FAINT
	TOKEN_ITALIC
	TOKEN_UNDERLINE
	TOKEN_BLINK_SLOW
	TOKEN_BLINK_RAPID
	TOKEN_REVERSE
	TOKEN_CROSSED_OUT
	TOKEN_FG
	TOKEN_BG
	TOKEN_CURSOR_UP
	TOKEN_CURSOR_DOWN
	TOKEN_CURSOR_FORWARD
	TOKEN_CURSOR_BACKWARD
	TOKEN_CURSOR_NEXT_LINE
	TOKEN_CURSOR_PREV_LINE
	TOKEN_CURSOR_SET_POSITION
	TOKEN_SCREEN_CLEAR
	TOKEN_LINE_CLEAR
)

var tokenFromString = map[string]TokenType{
	"TEXT":                TOKEN_TEXT,
	"NL":                  TOKEN_NL,
	"COMMAND_START":       TOKEN_COMMAND_START,
	"COMMAND_END":         TOKEN_COMMAND_END,
	"COMMAND_TEXT":        TOKEN_COMMAND_TEXT,
	"FUNC":                TOKEN_FUNC,
	"START_LOOP":          TOKEN_START_LOOP,
	"END_LOOP":            TOKEN_END_LOOP,
	"START_COND":          TOKEN_START_COND,
	"END_COND":            TOKEN_END_COND,
	"EOF":                 TOKEN_EOF,
	"RESET":               TOKEN_RESET,
	"BOLD":                TOKEN_BOLD,
	"FAINT":               TOKEN_FAINT,
	"ITALIC":              TOKEN_ITALIC,
	"UNDERLINE":           TOKEN_UNDERLINE,
	"BLINK_SLOW":          TOKEN_BLINK_SLOW,
	"BLINK_RAPID":         TOKEN_BLINK_RAPID,
	"REVERSE":             TOKEN_REVERSE,
	"CROSSED_OUT":         TOKEN_CROSSED_OUT,
	"FG":                  TOKEN_FG,
	"BG":                  TOKEN_BG,
	"CURSOR_UP":           TOKEN_CURSOR_UP,
	"CURSOR_DOWN":         TOKEN_CURSOR_DOWN,
	"CURSOR_FORWARD":      TOKEN_CURSOR_FORWARD,
	"CURSOR_BACKWARD":     TOKEN_CURSOR_BACKWARD,
	"CURSOR_NEXT_LINE":    TOKEN_CURSOR_NEXT_LINE,
	"CURSOR_PREV_LINE":    TOKEN_CURSOR_PREV_LINE,
	"CURSOR_SET_POSITION": TOKEN_CURSOR_SET_POSITION,
	"SCREEN_CLEAR":        TOKEN_SCREEN_CLEAR,
	"LINE_CLEAR":          TOKEN_LINE_CLEAR,
}

var UnknownTokenError = errors.New("Unknown token")

func TokenFromString(s string) (TokenType, error) {
	if token, ok := tokenFromString[strings.ToUpper(s)]; ok {
		return token, nil
	} else {
		return 0, UnknownTokenError
	}
}
