package mecca

import (
	"fmt"
	"io"
	"os"
	"strconv" // new import for locate token
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Updated colorMapping to use ANSI 16-color codes.
var colorMapping = map[string]lipgloss.Color{
	"black":        lipgloss.Color("0"),
	"red":          lipgloss.Color("1"),
	"green":        lipgloss.Color("2"),
	"yellow":       lipgloss.Color("3"),
	"blue":         lipgloss.Color("4"),
	"magenta":      lipgloss.Color("5"),
	"cyan":         lipgloss.Color("6"),
	"white":        lipgloss.Color("7"),
	"lightblack":   lipgloss.Color("8"),
	"lightred":     lipgloss.Color("9"),
	"lightgreen":   lipgloss.Color("10"),
	"lightyellow":  lipgloss.Color("11"),
	"lightblue":    lipgloss.Color("12"),
	"lightmagenta": lipgloss.Color("13"),
	"lightcyan":    lipgloss.Color("14"),
	"lightwhite":   lipgloss.Color("15"),
}

// Updated isColorToken to check the colorMapping keys.
func isColorToken(s string) bool {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "#") {
		return true
	}
	_, exists := colorMapping[s]
	return exists
}

func (interpreter *Interpreter) InterpretFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return interpreter.Interpret(string(data)), nil
}

// Interpret processes the input string containing MECCA tokens and literal text,
// applies the current styling via lipgloss, and returns the rendered output.
func (interpreter *Interpreter) Interpret(input string) string {
	output := ""
	currentStyle := interpreter.renderer.NewStyle()
	for {
		start := strings.Index(input, "[")
		if start == -1 {
			// Split literal text on newline, style each line then add newline back.
			lines := strings.Split(input, "\n")
			for i, line := range lines {
				output += currentStyle.Render(line)
				if i < len(lines)-1 {
					output += "\n"
				}
			}
			break
		}
		// Process literal text before token.
		literal := input[:start]
		lines := strings.Split(literal, "\n")
		for i, line := range lines {
			output += currentStyle.Render(line)
			if i < len(lines)-1 {
				output += "\n"
			}
		}
		end := strings.Index(input[start:], "]")
		if end == -1 {
			// unmatched token, render remainder as-is.
			remainder := input[start:]
			lines = strings.Split(remainder, "\n")
			for i, line := range lines {
				output += currentStyle.Render(line)
				if i < len(lines)-1 {
					output += "\n"
				}
			}
			break
		}
		tokenContent := input[start+1 : start+end]
		output += interpreter.processToken(tokenContent, &currentStyle)
		input = input[start+end+1:]
	}
	return output
}

// Updated processToken function to wrap returned token text with the current style.
func (interpreter *Interpreter) processToken(content string, style *lipgloss.Style) string {
	parts := strings.Fields(content)
	result := ""
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		// Handle special tokens.
		switch strings.ToLower(part) {
		case "cls": // [cls] token: clear screen.
			result += ansi.EraseDisplay(2) + ansi.CursorPosition(1, 1)
			continue
		case "cleos": // [cleos] token: clear to end of screen.
			result += ansi.EraseDisplay(0)
			continue
		case "cleol": // [cleol] token: clear to end of line.
			result += ansi.EraseLine(0)
		case "blink": // [blink] token: blink text.
			*style = style.Blink(true)
			continue
		case "bright": // [bright] token: brighten text. (synonym for bold)
			fallthrough
		case "bold": // [bold] token: bold text.
			*style = style.Bold(true)
			continue
		case "underline": // [underline] token: underline text.
			*style = style.Underline(true)
			continue
		case "italic": // [italic] token: italicize text.
			*style = style.Italic(true)
			continue
		case "dim": // [dim] token: dim text.
			*style = style.Faint(true)
			continue
		case "reverse": // [reverse] token: reverse text.
			*style = style.Reverse(true)
			continue
		case "strike": // [strike] token: strike through text.
			*style = style.Strikethrough(true)
			continue
		case "reset": // [reset] token: remove all styling.
			*style = interpreter.renderer.NewStyle()
			continue
		case "locate": // [locate] token: expects two arguments.
			if i+2 < len(parts) {
				col, err1 := strconv.Atoi(parts[i+1])
				row, err2 := strconv.Atoi(parts[i+2])
				if err1 == nil && err2 == nil {
					// ANSI escape sequence: CSI row;colH (adding 1 for 1-indexing)
					result += ansi.CursorPosition(row+1, col+1)
				}
				i += 2
			}
			continue
		case "cr": // [cr] token: moves the cursor to the beginning of the line.
			result += "\r"
			continue
		case "lf": // [lf] token: moves the cursor to the next line.
			result += ansi.CursorNextLine(1)
			continue
		case "up": // [up] token: moves the cursor up one line.
			result += ansi.CursorUp(1)
			continue
		case "down": // [down] token: moves the cursor down one line.
			result += ansi.CursorDown(1)
			continue
		case "right": // [right] token: moves the cursor right one column.
			result += ansi.CursorForward(1)
			continue
		case "left": // [left] token: moves the cursor left one column.
			result += ansi.CursorBackward(1)
			continue
		case "savecursor": // [savecursor] token: saves the current cursor position.
			result += ansi.SaveCursor
			continue
		case "restorecursor": // [restorecursor] token: restores the saved cursor position.
			result += ansi.RestoreCursor
			continue
		case "line": // draws a line of a specified length using the specified character.
			if i+2 < len(parts) {
				length, err1 := strconv.Atoi(parts[i+1])
				char := parts[i+2]
				if err1 == nil {
					result += strings.Repeat(char, length)
				}
				i += 2
			}
			continue
		}
		// Colors can be specified one of three ways:
		// 1. By name (e.g., red, green, blue)
		// 2. By hex code (e.g., #ff0000, #00ff00, #0000ff)
		// 3. By ANSI color code, as a number (e.g., #63)
		if isColorToken(part) {
			tokenLower := strings.ToLower(part)
			if strings.HasPrefix(tokenLower, "#") {
				if len(tokenLower) == 7 {
					*style = style.Foreground(lipgloss.Color(tokenLower))
				} else {
					// ANSI color code
					if n, err := parseNumber(tokenLower[1:]); err == nil {
						*style = style.Foreground(lipgloss.Color(strconv.Itoa(n)))
					}
				}
			} else if col, ok := colorMapping[tokenLower]; ok {
				*style = style.Foreground(col)
			}
			if i+2 < len(parts) && strings.ToLower(parts[i+1]) == "on" {
				bgToken := strings.ToLower(parts[i+2])
				if strings.HasPrefix(bgToken, "#") {
					if len(bgToken) == 7 {
						*style = style.Background(lipgloss.Color(bgToken))
					} else {
						// ANSI color code
						if n, err := parseNumber(bgToken[1:]); err == nil {
							*style = style.Background(lipgloss.Color(strconv.Itoa(n)))
						}
					}
				} else if bg, ok := colorMapping[bgToken]; ok {
					*style = style.Background(bg)
				}
				i += 2
			}
			continue
		}
		// Handle UTF-8 tokens [U+xxxx]
		if strings.HasPrefix(part, "U+") && len(part) > 2 {
			if r, err := decodeUTF8Token(part[2:]); err == nil {
				result += style.Render(string(r))
				continue
			}
		} else if isNumber(part) {
			// Handle ASCII token from a number.
			if n, err := parseNumber(part); err == nil {
				result += style.Render(string(rune(n)))
				continue
			}
		}
		// Look up registered tokens.
		if token, ok := interpreter.GetToken(part); ok {
			if token.ArgCount > 0 && i+token.ArgCount < len(parts) {
				args := parts[i+1 : i+1+token.ArgCount]
				tokenOut := token.Func(args)
				result += style.Render(tokenOut)
				i += token.ArgCount
			} else {
				tokenOut := token.Func([]string{})
				result += style.Render(tokenOut)
			}
			continue
		}
		// If token is unrecognized, emit an error message inline.
		result += style.Render(fmt.Sprintf("[UNRECOGNIZED TOKEN \"%s\"]", part))
	}
	return result
}

// isNumber checks if a string consists solely of digits.
func isNumber(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// parseNumber converts a numeric string to an integer.
func parseNumber(s string) (int, error) {
	var n int
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// decodeUTF8Token converts a hexadecimal string to a rune.
func decodeUTF8Token(hexStr string) (rune, error) {
	var n int
	_, err := fmt.Sscanf(hexStr, "%x", &n)
	if err != nil {
		return 0, err
	}
	return rune(n), nil
}

func (interpreter *Interpreter) Render(input string) {
	io.WriteString(interpreter.writer, interpreter.Interpret(input))
}

func (interpreter *Interpreter) RenderFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	interpreter.Render(string(data))
	return nil
}
