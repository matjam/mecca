package lexer

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func jsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s[1 : len(s)-1]
}

func (t Token) String() string {
	return fmt.Sprintf("%v (line %v col %v): \"%v\"", t.Type, t.Line, t.Column, jsonEscape(t.Value))
}

type lexMode int

type Lexer struct {
	input        *bufio.Reader // the raw input for the template
	output       chan Token    // a channel holding the next available Token
	buffer       string        // a buffer for the current token
	bufferLine   int           // the line number where the buffer started
	bufferColumn int           // the column number where the buffer started
	tokenType    TokenType     // the type of token currently being processed
	line         int           // the current line number
	column       int           // the current column number
	err          error         // the last error encountered
	done 	   bool          // whether or not we've reached the end of the input
}

func NewLexer(r io.Reader) *Lexer {
	l := &Lexer{
		input:        bufio.NewReader(r),
		output:       make(chan Token),
		buffer:       "",
		bufferLine:   1,
		bufferColumn: 1,
		tokenType:         TOKEN_TEXT,
		line:         1,
		column:       1,
	}

	go func() {
		l.process()
	}()

	return l
}



// Lex returns the next token from the input.
func (l *Lexer) Lex() (Token, error) {
	eofToken := Token{
		Type:   TOKEN_EOF,
		Value:  "EOF",
		Line:   l.line,
		Column: l.column,
	}

	if l.err != nil {
		return eofToken, l.err
	}

	t, ok := <-l.output
	if !ok {
		return eofToken, errors.New("lexer: output channel closed")
	}

	return t, nil
}

// process is the entry point for the lexer. It reads the next rune
// from the input and decides what to do with it. We start in MODE_UNKNOWN and
// switch to MODE_TEXT or MODE_COMMAND depending on the input. Once we know
// which mode we're in, we call the appropriate function to process the input.
//
// We loop because we don't return until the input is exhausted, at which point
// we close the output channel and return.
func (l *Lexer) process() {
	for !l.done {
		r, err := l.next()
		if err != nil || l.err != nil {
			if l.buffer != "" {
				l.emit(TOKEN_TEXT)
			}
			l.reset().save().appendString("EOF").emit(TOKEN_EOF)

			close(l.output)
			l.done = true
			return
		}

		// effectively this is a state machine, which switches between modes
		// depending on the input..
		switch l.tokenType {
		case TOKEN_TEXT:
			l.processTextMode(r)
		case TOKEN_COMMAND:
			l.processCommandMode(r)
		case TOKEN_COMMAND_ARG:
			l.processCommandArgMode(r)
		}
	}
}

func (l *Lexer) nextLine() *Lexer {
	l.line++
	l.column = 1

	return l
}

func (l *Lexer) nextColumn() *Lexer {
	l.column++

	return l
}

func (l *Lexer) next() (rune, error) {
	r, _, err := l.input.ReadRune()
	if err != nil {
		return 0, err
	}

	return r, nil
}

// start starts a new mode
func (l *Lexer) start(m TokenType) *Lexer {
	l.tokenType = m

	return l
}

// emit sends the current token to the output channel
func (l *Lexer) emit(tokenType TokenType) *Lexer {
	if tokenType == TOKEN_TEXT && l.buffer == "" {
		// we don't want to emit empty text tokens
		return l
	}

	// depending on the command the tokenType may be different from the current one
	if tokenType == TOKEN_COMMAND {
		switch strings.ToLower(l.buffer) {
		case "for":
			tokenType = TOKEN_START_LOOP
		case "/for":
			tokenType = TOKEN_END_LOOP
		case "if":
			tokenType = TOKEN_START_COND
		case "/if":
			tokenType = TOKEN_END_COND
		case "reset":
			tokenType = TOKEN_RESET
		case "bold":
			tokenType = TOKEN_BOLD
		case "faint":
			tokenType = TOKEN_FAINT
		case "italic":
			tokenType = TOKEN_ITALIC
		case "underline":
			tokenType = TOKEN_UNDERLINE
		case "blink":
			tokenType = TOKEN_BLINK_SLOW
		case "blinkslow":	
			tokenType = TOKEN_BLINK_SLOW
		case "blinkrapid":
			tokenType = TOKEN_BLINK_RAPID
		case "reverse":
			tokenType = TOKEN_REVERSE
		case "crossedout":
			tokenType = TOKEN_CROSSED_OUT
		case "fg":
			tokenType = TOKEN_FG
		case "bg":
			tokenType = TOKEN_BG
		case "up":
			tokenType = TOKEN_CURSOR_UP
		case "down":
			tokenType = TOKEN_CURSOR_DOWN
		case "forward":
			tokenType = TOKEN_CURSOR_FORWARD
		case "backward":
			tokenType = TOKEN_CURSOR_BACKWARD
		case "position":
			tokenType = TOKEN_CURSOR_SET_POSITION
		case "clear":
			tokenType = TOKEN_SCREEN_CLEAR
		case "lineclear":
			tokenType = TOKEN_LINE_CLEAR
		case "no":
			tokenType = TOKEN_NO
		}
	}

	l.output <- Token{
		Type:   tokenType,
		Value:  l.buffer,
		Line:   l.bufferLine,
		Column: l.bufferColumn,
	}

	return l
}

func (l *Lexer) append(r rune) *Lexer {
	l.buffer += string(r)

	return l
}

func (l *Lexer) appendString(s string) *Lexer {
	for _, r := range s {
		l.append(r)
	}

	return l
}

// reset the buffer
func (l *Lexer) reset() *Lexer {
	l.buffer = ""

	return l
}

// save the current position in the input
func (l *Lexer) save() *Lexer {
	l.bufferLine = l.line
	l.bufferColumn = l.column

	return l
}

// peek the next rune
func (l *Lexer) peek() (rune, error) {
	r, _, err := l.input.ReadRune()
	if err != nil {
		l.err = err
		return 0, err
	}

	l.input.UnreadRune()

	return r, nil
}

func (l *Lexer) processTextMode(r rune) {
	switch r {
	case '\n':
		l.emit(TOKEN_TEXT)
		l.reset().save().append('\n').emit(TOKEN_NL).nextLine()
		l.reset().save().start(TOKEN_TEXT)
	case '[':
		// we may have either a COMMAND_START or a literal '['. Because we need to handle
		// the case where there are multiple '[' in a row, we need to peek ahead to see
		// what the next rune is.

		r, err := l.peek()
		if err != nil {
			// we've reached the end of the input and the last rune was a '['.
			l.appendString("[").emit(TOKEN_TEXT).nextColumn()
			l.reset().save().appendString("EOF").emit(TOKEN_EOF)
			l.done = true
			return
		}
	
		if r == '[' {
			_, err = l.next() // we eat the next rune, and because it was a '[' we know its a literal '['.
			if err != nil {	
				// somehow the peek was successful but we failed to read the rune we just peeked.
				// this is probably impossible, so we will panic.
				panic(fmt.Errorf("Unexpected error parsing end of input: %s", err))
			}
			l.nextColumn().nextColumn() // we need to increment the column twice because we ate two runes.
			l.appendString("[")
			return
		} 

		// we have a COMMAND_START
		l.emit(TOKEN_TEXT)
		l.reset().save().append('[').emit(TOKEN_COMMAND_START).nextColumn()
		l.reset().start(TOKEN_COMMAND)
	default:
		if l.buffer == "" {
			l.save()
		}
		l.append(r).nextColumn()
	}
}

func (l *Lexer) processCommandMode(r rune) {
	switch r {
	case ']':
		l.emit(TOKEN_COMMAND)
		l.reset().save().append(']').emit(TOKEN_COMMAND_END).nextColumn()
		l.reset().save().start(TOKEN_TEXT)
	case ' ':
		l.emit(TOKEN_COMMAND).nextColumn()
		l.reset().save().start(TOKEN_COMMAND_ARG)
	case '\n':
		l.err = errors.New(fmt.Sprintf("Unexpected newline in command '%s' at %v:%v", l.buffer, l.line, l.column))
	default:
		if l.buffer == "" {
			l.save()
		}
		l.append(r).nextColumn()
	}
}

// we only care about ']' and ' ' in command arg mode. Everything else is
// appended to the buffer.
func (l *Lexer) processCommandArgMode(r rune) {
	switch r {
	case ']':
		l.emit(TOKEN_COMMAND_ARG)
		l.reset().save().append(']').emit(TOKEN_COMMAND_END).nextColumn()
		l.reset().save().start(TOKEN_TEXT)
	case ' ':
		l.emit(TOKEN_COMMAND_ARG).nextColumn()
		l.reset().append(' ').reset().save()
	default:
		if l.buffer == "" {
			l.save()
		}
		l.append(r).nextColumn()
	}
}
