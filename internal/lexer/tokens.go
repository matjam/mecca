package lexer

//go:generate stringer -type TokenType -trimprefix TOKEN_
type TokenType int

const (
	TOKEN_TEXT TokenType = iota
	TOKEN_NL
	TOKEN_COMMAND_START
	TOKEN_COMMAND_END
	TOKEN_COMMAND
	TOKEN_COMMAND_ARG
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
