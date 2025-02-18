package main

import (
	"fmt"

	"github.com/matjam/mecca"
)

type ServerContext struct {
	user     string
	msgCount int
}

func example1() {
	ctx := ServerContext{
		user:     "Bob",
		msgCount: 3,
	}

	// Create a new MECCA interpreter.
	interpreter := mecca.NewInterpreter(mecca.WithTemplateRoot("cmd/example"))

	// Register a custom token for demonstration.
	interpreter.RegisterToken("user", ctx.userToken, 0)
	interpreter.RegisterToken("msgcount", ctx.msgCountToken, 1)

	// Example MECCA template using a registered token and standard tokens.
	template := `
[cls white]This is an example of a MECCA template using a registered token and standard tokens.

[bold yellow]Welcome, [user]! [reset]
You have [lightblue msgcount 3 reset] new messages.

[bold yellow][line 80 -]
There are multiple examples; you can execute them by running the following commands:

[bold cyan]go run ./cmd/example 1
go run ./cmd/example 2
go run ./cmd/example 3
go run ./cmd/example 4

This is an example of a string that was passed at execution time: [bold red foo reset]
`

	// Render the template.
	interpreter.ExecString(template, map[string]any{"foo": "bar"})
}

// userToken returns a sample user name.
func (ctx ServerContext) userToken(_ []string) string {
	return ctx.user
}

// msgCountToken returns a sample message count.
func (ctx ServerContext) msgCountToken(args []string) string {
	return fmt.Sprintf("%d", ctx.msgCount)
}
