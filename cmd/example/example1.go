package main

import (
	"fmt"
	"os"
	"time" // added for timing

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
	interpreter := mecca.NewInterpreter(os.Stdout)

	// Register a custom token for demonstration.
	interpreter.RegisterToken("user", ctx.userToken, 0)
	interpreter.RegisterToken("msgcount", ctx.msgCountToken, 1)

	// Example MECCA template using a registered token and standard tokens.
	template := `
[cls bold yellow]Welcome, [user]! [reset]
You have [lightblue msgcount 3 reset] new messages.
`

	for i := 0; i < 240; i++ {
		if i%24 == 0 {
			template += "\n"
		}
		template += fmt.Sprintf("[#%d]█", i+16)
	}

	start := time.Now() // start timing

	// Render the template.
	interpreter.Render(template)

	// Print elapsed time.
	fmt.Printf("\nInterpretation took: %v\n", time.Since(start))
}

// userToken returns a sample user name.
func (ctx ServerContext) userToken(_ []string) string {
	return ctx.user
}

// msgCountToken returns a sample message count.
func (ctx ServerContext) msgCountToken(args []string) string {
	return fmt.Sprintf("%d", ctx.msgCount)
}
