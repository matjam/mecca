package mecca

import (
	"fmt"
	"strconv" // new import for locate token
	"strings"

	"github.com/charmbracelet/lipgloss"
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

// Interpret processes the input string containing MECCA tokens and literal text,
// applies the current styling via lipgloss, and returns the rendered output.
func Interpret(input string) string {
	output := ""
	currentStyle := lipgloss.NewStyle()
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
		output += processToken(tokenContent, &currentStyle)
		input = input[start+end+1:]
	}
	return output
}

// Updated processToken function to wrap returned token text with the current style.
func processToken(content string, style *lipgloss.Style) string {
	parts := strings.Fields(content)
	result := ""
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		// Handle special tokens.
		switch strings.ToLower(part) {
		case "cls":
			// Clear screen and move cursor to top left.
			result += "\033[2J\033[H"
			continue
		case "blink": // Update blink token to use lipgloss's blink style.
			*style = style.Blink(true)
			continue
		case "bold":
			*style = style.Bold(true)
			continue
		case "reset": // New reset token: remove all styling.
			*style = lipgloss.NewStyle()
			continue
		case "locate": // New locate token: expects two arguments.
			if i+2 < len(parts) {
				col, err1 := strconv.Atoi(parts[i+1])
				row, err2 := strconv.Atoi(parts[i+2])
				if err1 == nil && err2 == nil {
					// ANSI escape sequence: CSI row;colH (adding 1 for 1-indexing)
					result += fmt.Sprintf("\033[%d;%dH", row+1, col+1)
				}
				i += 2
			}
			continue
		case "cr": // New [cr] token: moves the cursor to the beginning of the line.
			result += "\r"
			continue
		}
		// Handle color tokens.
		if isColorToken(part) {
			tokenLower := strings.ToLower(part)
			// If token starts with '#' use as is; else use mapping.
			if strings.HasPrefix(tokenLower, "#") {
				*style = style.Foreground(lipgloss.Color(tokenLower))
			} else if col, ok := colorMapping[tokenLower]; ok {
				*style = style.Foreground(col)
			}
			if i+2 < len(parts) && strings.ToLower(parts[i+1]) == "on" {
				bgToken := strings.ToLower(parts[i+2])
				if strings.HasPrefix(bgToken, "#") {
					*style = style.Background(lipgloss.Color(bgToken))
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
		if token, ok := GetToken(part); ok {
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
		// If token is unrecognized, simply do not output anything.
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
