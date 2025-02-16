package main

import (
	"fmt"
	"strings"

	"github.com/matjam/mecca"
)

func main() {
	// Define flag dimensions
	width := 80

	// Define stripe colors and their heights.
	type stripe struct {
		color string // token for color (foreground and background)
		rows  int    // number of rows for this stripe
	}
	// LGBTQ+ Pride flag stripes.
	stripes := []stripe{
		{"#E40303", 3}, // Red
		{"#FF8C00", 3}, // Orange
		{"#FFED00", 3}, // Yellow
		{"#008026", 3}, // Green
		{"#004DFF", 3}, // Blue
		{"#750787", 3}, // Violet
	}

	var sb strings.Builder
	block := strings.Repeat("█", width)

	// Build the flag template.
	for _, s := range stripes {
		// Use [reset] to clear previous styling.
		// Set both foreground and background to same color for a full colored block.
		tokenPrefix := fmt.Sprintf("[reset][%s on %s]", s.color, s.color)
		for i := 0; i < s.rows; i++ {
			sb.WriteString(tokenPrefix)
			sb.WriteString(block)
			sb.WriteString("\n")
		}
	}

	template := sb.String()
	result := mecca.Interpret(template)
	fmt.Print(result)
}
