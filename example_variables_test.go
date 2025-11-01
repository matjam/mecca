package mecca

import (
	"bytes"
	"fmt"
)

func Example_variables() {
	var buf bytes.Buffer
	interpreter := NewInterpreter(WithWriter(&buf))

	// Variables can be passed at execution time to override or provide values
	template := `Hello, [name]!
You have [count] new messages.
Welcome to [system]!
`

	// Pass variables that will be substituted for [name], [count], and [system]
	interpreter.ExecString(template, map[string]any{
		"name":   "Alice",
		"count":  42,
		"system": "MECCA Interpreter",
	})
	fmt.Print(buf.String())
}
