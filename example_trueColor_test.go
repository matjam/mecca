package mecca

import (
	"bytes"
	"fmt"
	"strings"
)

func Example_trueColor() {
	var buf bytes.Buffer
	interpreter := NewInterpreter(WithWriter(&buf))

	// MECCA supports true color (RGB hex codes) for precise color control
	var sb strings.Builder

	// Define colors using hex codes
	colors := []struct {
		hex  string
		name string
	}{
		{"#FF0000", "Red"},
		{"#00FF00", "Green"},
		{"#0000FF", "Blue"},
		{"#FF69B4", "Hot Pink"},
		{"#FFFF00", "Yellow"},
	}

	sb.WriteString("[bold]True Color Examples:[reset]\n\n")
	for _, c := range colors {
		sb.WriteString(fmt.Sprintf("[%s]%s[reset] - [%s on black]█[reset]\n", c.hex, c.name, c.hex))
	}

	interpreter.ExecString(sb.String(), nil)
	fmt.Print(buf.String())
}
