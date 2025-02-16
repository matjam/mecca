package main

import (
	"fmt"
	"time" // added for timing

	"github.com/matjam/mecca"
)

func main() {
	// Register a custom token for demonstration.
	mecca.RegisterToken("user", userToken, 0)
	mecca.RegisterToken("msgcount", msgCountToken, 1)

	// Example MECCA template using a registered token and standard tokens.
	template := `[cls bold yellow]Welcome, [user]! [reset]
You have [lightblue msgcount 3 reset] new messages.
[locate 10 5 bold red]Hello, World![reset]
`

	start := time.Now() // start timing

	// Interpret the template.
	result := mecca.Interpret(template)
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
