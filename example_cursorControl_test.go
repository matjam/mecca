package mecca

import (
	"bytes"
	"fmt"
)

func Example_cursorControl() {
	var buf bytes.Buffer
	interpreter := NewInterpreter(WithWriter(&buf))

	// MECCA provides various cursor control tokens
	template := `[cls]Screen cleared!
[locate 5 10]Cursor moved to row 5, column 10[reset]
[bold]Cursor movement tokens:[reset]
[up]Moved up[down]Moved down[reset]
[right]Right[left]Left[reset]
[cr]Carriage return[lf]Line feed[reset]
`

	interpreter.ExecString(template, nil)
	fmt.Print(buf.String())
}
