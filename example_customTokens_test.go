package mecca

import (
	"bytes"
	"fmt"
	"strings"
)

func Example_customTokens() {
	var buf bytes.Buffer
	interpreter := NewInterpreter(WithWriter(&buf))

	// Register custom tokens that can be used in templates
	interpreter.RegisterToken("greet", func(args []string) string {
		if len(args) > 0 {
			return "Hello, " + args[0] + "!"
		}
		return "Hello!"
	}, 1)

	interpreter.RegisterToken("repeat", func(args []string) string {
		if len(args) < 2 {
			return ""
		}
		count := 0
		for _, c := range args[0] {
			count = count*10 + int(c-'0')
		}
		return strings.Repeat(args[1], count)
	}, 2)

	interpreter.RegisterToken("uppercase", func(args []string) string {
		if len(args) > 0 {
			return strings.ToUpper(args[0])
		}
		return ""
	}, 1)

	// Use the custom tokens in a template
	template := `[greet World][reset]
[repeat 5 *][reset]
[uppercase hello world][reset]
`

	interpreter.ExecString(template, nil)
	fmt.Print(buf.String())
}
