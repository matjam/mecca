package mecca

import (
	"bytes"
	"fmt"
)

func Example_styles() {
	var buf bytes.Buffer
	interpreter := NewInterpreter(WithWriter(&buf))

	// Demonstrate various text style tokens
	template := `[bold]Bold text[reset]
[underline]Underlined text[reset]
[italic]Italic text[reset]
[dim]Dim text[reset]
[reverse]Reversed text[reset]
[strike]Strikethrough text[reset]
[blink]Blinking text[steady] (now steady)[reset]
[red]Red [on white]on white[reset] background
`

	interpreter.ExecString(template, nil)
	fmt.Print(buf.String())
}
