package mecca

import (
	"bytes"
	"fmt"
)

type exampleServerContext struct {
	user     string
	msgCount int
}

// userToken returns a sample user name.
func (ctx exampleServerContext) userToken(_ []string) string {
	return ctx.user
}

// msgCountToken returns a sample message count.
func (ctx exampleServerContext) msgCountToken(args []string) string {
	return fmt.Sprintf("%d", ctx.msgCount)
}

func Example_basic() {
	ctx := exampleServerContext{
		user:     "Bob",
		msgCount: 3,
	}

	// Create a new MECCA interpreter.
	var buf bytes.Buffer
	interpreter := NewInterpreter(WithWriter(&buf))

	// Register a custom token for demonstration.
	interpreter.RegisterToken("user", ctx.userToken, 0)
	interpreter.RegisterToken("msgcount", ctx.msgCountToken, 1)

	// Example MECCA template using a registered token and standard tokens.
	template := `[bold yellow]Welcome, [user]! [reset]
You have [lightblue]3[reset] new messages.

[bold yellow][line 40 -][reset]
This is an example of a variable passed at execution time: [bold red][foo][reset]`

	// Render the template.
	interpreter.ExecString(template, map[string]any{"foo": "bar"})
	fmt.Print(buf.String())
}
