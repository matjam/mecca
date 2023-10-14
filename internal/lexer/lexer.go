package lexer

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

const (
	MODE_UNKNOWN lexMode = iota
	MODE_TEXT
	MODE_BRACKET_LITERAL
	MODE_COMMAND
	MODE_COMMAND_ARG
	MODE_COMMAND_START
)

type Lexer struct {
	input        *bufio.Reader // the raw input for the template
	output       chan Token    // a channel holding the next available Token
	buffer       string        // a buffer for the current token
	bufferLine   int           // the line number where the buffer started
	bufferColumn int           // the column number where the buffer started
	mode         lexMode       // the current mode of the lexer
	line         int           // the current line number
	column       int           // the current column number
	err          error         // the last error encountered
}

func NewLexer(r io.Reader) *Lexer {
	l := &Lexer{
		input:        bufio.NewReader(r),
		output:       make(chan Token),
		buffer:       "",
		bufferLine:   1,
		bufferColumn: 1,
		mode:         MODE_UNKNOWN,
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
	if l.err != nil {
		return Token{}, l.err
	}

	t, ok := <-l.output
	if !ok {
		return Token{}, errors.New("lexer: output channel closed")
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
	for {
		r, err := l.next()
		if err != nil || l.err != nil {
			if l.buffer != "" {
				l.emit(TOKEN_TEXT)
			}
			l.reset().save().appendString("EOF").emit(TOKEN_EOF)

			close(l.output)
			return
		}

		// effectively this is a state machine, which switches between modes
		// depending on the input..
		switch l.mode {
		case MODE_UNKNOWN:
			l.processUnknownMode(r)
		case MODE_TEXT:
			l.processTextMode(r)
		case MODE_BRACKET_LITERAL:
			l.processTextMode(r)
		case MODE_COMMAND:
			l.processCommandMode(r)
		case MODE_COMMAND_ARG:
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
func (l *Lexer) start(m lexMode) *Lexer {
	l.mode = m

	return l
}

// emit sends the current token to the output channel
func (l *Lexer) emit(tokenType TokenType) *Lexer {
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

func (l *Lexer) reset() *Lexer {
	l.buffer = ""

	return l
}

func (l *Lexer) save() *Lexer {
	l.bufferLine = l.line
	l.bufferColumn = l.column

	return l
}

func (l *Lexer) peek() rune {
	r, _, err := l.input.ReadRune()
	if err != nil {
		return 0
	}

	l.input.UnreadRune()

	return r
}

func (l *Lexer) processUnknownMode(r rune) {
	switch r {
	case '[':
		peekedRune := l.peek()
		if peekedRune == '[' {
			// clear the buffer in case there's anything in it, save the current location
			// then discard the next character, as we know it is '['.
			_, _ = l.reset().save().next() // eat the next rune
			l.start(MODE_TEXT).appendString("[")
			l.nextColumn().nextColumn() // increment the column counter twice
		} else if peekedRune == rune(0) {
			l.err = errors.New(fmt.Sprintf("Unexpected EOF in unknown mode at %v:%v", l.line, l.column))
		} else {
			l.append(r).emit(TOKEN_COMMAND_START).nextColumn()
			l.reset().save().start(MODE_COMMAND)
		}
	default:
		l.reset().save().start(MODE_TEXT)
		l.processTextMode(r)
	}
}

func (l *Lexer) processTextMode(r rune) {
	switch r {
	case '\n':
		l.emit(TOKEN_TEXT)
		l.reset().save().append('\n').emit(TOKEN_NL).nextLine()
		l.reset().save().start(MODE_UNKNOWN)
	case '[':
		if l.peek() == '[' {
			l.appendString("[")
			l.nextColumn()
			l.start(MODE_BRACKET_LITERAL)
			return
		}
		if l.mode == MODE_BRACKET_LITERAL {
			l.nextColumn() // we eat this one
			l.start(MODE_TEXT)
			return
		}

		l.emit(TOKEN_TEXT)
		l.reset().save().append('[').emit(TOKEN_COMMAND_START).nextColumn()
		l.reset().start(MODE_COMMAND)
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
		l.reset().save().start(MODE_TEXT)
	case ' ':
		l.emit(TOKEN_COMMAND).nextColumn()
		l.reset().save().start(MODE_COMMAND_ARG)
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
		l.reset().save().start(MODE_UNKNOWN)
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
