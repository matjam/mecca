package main

import (
	"fmt"
	"time" // added for timing

	"github.com/matjam/mecca"
)

func main() {
	// Create a new MECCA interpreter.
	interpreter := mecca.NewInterpreter()

	// Register a custom token for demonstration.
	interpreter.RegisterToken("user", userToken, 0)
	interpreter.RegisterToken("msgcount", msgCountToken, 1)

	// Example MECCA template using a registered token and standard tokens.
	template := `
[cls bold yellow]Welcome, [user]! [reset]
You have [lightblue msgcount 3 reset] new messages.
[locate 10 5 bold red]Hello, [#127]World![reset]
[reset][locate 12 0]
`

	for i := 0; i < 240; i++ {
		if i%24 == 0 {
			template += "\n"
		}
		template += fmt.Sprintf("[#%d]█", i+16)
	}

	start := time.Now() // start timing

	// Interpret the template.
	result := interpreter.Interpret(template)
	fmt.Println(result)

	// Print elapsed time.
	fmt.Printf("\nInterpretation took: %v\n", time.Since(start))
}

// userToken returns a sample user name.
func userToken(_ []string) string {
	// ...existing code...
	return "John Doe"
}

// msgCountToken returns a sample message count.
func msgCountToken(args []string) string {
	// ...existing code...
	return args[0]
}
