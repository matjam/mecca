package mecca

import (
	"bytes"
	"fmt"
	"strings"
)

func Example_colors() {
	// Create a new MECCA interpreter.
	var buf bytes.Buffer
	interpreter := NewInterpreter(WithWriter(&buf))

	// Build a MECCA template demonstrating colors.
	var sb strings.Builder

	colors := []string{
		"red", "green", "blue", "yellow",
		"magenta", "cyan", "white",
	}

	// Only first eight colors can be used as background colors.
	allowedBgColors := []string{
		"black", "red", "green", "yellow",
		"blue", "magenta", "cyan", "white",
	}

	// Demonstrate foreground on background combinations.
	sb.WriteString("[bold]Foreground on Background Colors:[reset]\n")
	for _, fg := range colors[:3] { // Show first 3 for brevity
		for _, bg := range allowedBgColors[:4] { // Show first 4 backgrounds
			sb.WriteString(fmt.Sprintf("[%s on %s]%s/%s[reset] ", fg, bg, fg, bg))
		}
		sb.WriteString("\n")
	}

	// Interpret the MECCA template.
	interpreter.ExecString(sb.String(), nil)
	fmt.Print(buf.String())
}
