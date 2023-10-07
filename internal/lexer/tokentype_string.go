// Code generated by "stringer -type TokenType -trimprefix TOKEN_"; DO NOT EDIT.

package lexer

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TOKEN_TEXT-0]
	_ = x[TOKEN_NL-1]
	_ = x[TOKEN_COMMAND_START-2]
	_ = x[TOKEN_COMMAND_END-3]
	_ = x[TOKEN_COMMAND_TEXT-4]
	_ = x[TOKEN_FUNC-5]
	_ = x[TOKEN_START_LOOP-6]
	_ = x[TOKEN_END_LOOP-7]
	_ = x[TOKEN_START_COND-8]
	_ = x[TOKEN_END_COND-9]
	_ = x[TOKEN_EOF-10]
	_ = x[TOKEN_RESET-11]
	_ = x[TOKEN_BOLD-12]
	_ = x[TOKEN_FAINT-13]
	_ = x[TOKEN_ITALIC-14]
	_ = x[TOKEN_UNDERLINE-15]
	_ = x[TOKEN_BLINK_SLOW-16]
	_ = x[TOKEN_BLINK_RAPID-17]
	_ = x[TOKEN_REVERSE-18]
	_ = x[TOKEN_CROSSED_OUT-19]
	_ = x[TOKEN_FG-20]
	_ = x[TOKEN_BG-21]
	_ = x[TOKEN_CURSOR_UP-22]
	_ = x[TOKEN_CURSOR_DOWN-23]
	_ = x[TOKEN_CURSOR_FORWARD-24]
	_ = x[TOKEN_CURSOR_BACKWARD-25]
	_ = x[TOKEN_CURSOR_NEXT_LINE-26]
	_ = x[TOKEN_CURSOR_PREV_LINE-27]
	_ = x[TOKEN_CURSOR_SET_POSITION-28]
	_ = x[TOKEN_SCREEN_CLEAR-29]
	_ = x[TOKEN_LINE_CLEAR-30]
}

const _TokenType_name = "TEXTNLCOMMAND_STARTCOMMAND_ENDCOMMAND_TEXTFUNCSTART_LOOPEND_LOOPSTART_CONDEND_CONDEOFRESETBOLDFAINTITALICUNDERLINEBLINK_SLOWBLINK_RAPIDREVERSECROSSED_OUTFGBGCURSOR_UPCURSOR_DOWNCURSOR_FORWARDCURSOR_BACKWARDCURSOR_NEXT_LINECURSOR_PREV_LINECURSOR_SET_POSITIONSCREEN_CLEARLINE_CLEAR"

var _TokenType_index = [...]uint16{0, 4, 6, 19, 30, 42, 46, 56, 64, 74, 82, 85, 90, 94, 99, 105, 114, 124, 135, 142, 153, 155, 157, 166, 177, 191, 206, 222, 238, 257, 269, 279}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}