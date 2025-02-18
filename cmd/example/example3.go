package main

import (
	"fmt"
	"strings"

	"github.com/matjam/mecca"
)

func example3() {
	// Create a new MECCA interpreter.
	interpreter := mecca.NewInterpreter(mecca.WithTemplateRoot("cmd/example"))

	// Define flag dimensions
	width := 80

	// Define stripe colors and their heights.
	type stripe struct {
		color string // token for color (foreground and background)
		rows  int    // number of rows for this stripe
	}

	// Define the flag stripes.
	stripes := []stripe{
		{"#FF69B4", 2}, // Hot Pink
		{"#FF0000", 2}, // Red
		{"#FF8E00", 2}, // Orange
		{"#FFFF00", 2}, // Yellow
		{"#008E00", 2}, // Green
		{"#00C0C0", 2}, // Turquoise
		{"#400098", 2}, // Indigo
		{"#8E008E", 2}, // Violet
	}

	var sb strings.Builder
	block := strings.Repeat("█", width)

	sb.WriteString("[cls][bold]For my friends who have different orientations, I present to you the flag of diversity!\n\n")

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

	interpreter.ExecString(sb.String(), nil)
}
