package main

import (
	"fmt"
	"strings"

	"github.com/matjam/mecca"
)

func main() {
	// Create a new MECCA interpreter.
	interpreter := mecca.NewInterpreter()

	// Build a MECCA template demonstrating colors.
	var sb strings.Builder

	colors := []string{
		"black", "red", "green", "yellow",
		"blue", "magenta", "cyan", "white",
		"lightblack", "lightred", "lightgreen", "lightyellow",
		"lightblue", "lightmagenta", "lightcyan", "lightwhite",
	}

	// Only first eight colors can be used as background colors.
	allowedBgColors := []string{
		"black", "red", "green", "yellow",
		"blue", "magenta", "cyan", "white",
	}

	// Demonstrate foreground on background combinations.
	sb.WriteString("\n[bold]Foreground on Background Colors:\n ")
	for f, fg := range colors {
		sb.WriteString("\n[yellow on black]> ")

		for b, bg := range allowedBgColors {
			sb.WriteString(fmt.Sprintf("[%s on %s]%3v/%3v ", fg, bg, f, b))
		}
	}

	template := sb.String()
	// Interpret the MECCA template.
	result := interpreter.Interpret(template)
	fmt.Println(result)
}
