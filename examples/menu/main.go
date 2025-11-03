package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/matjam/mecca"
)

func main() {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	// Set template root to the examples/menu directory
	// This assumes running from the project root (e.g., "go run examples/menu/main.go")
	templateRoot := filepath.Join(wd, "examples/menu")

	// Create interpreter with stdin as reader for menu input
	interpreter := mecca.NewInterpreter(
		mecca.WithTemplateRoot(templateRoot),
		mecca.WithReader(os.Stdin),
	)

	// Execute the menu template
	interpreter.ExecString(`[include menu.mec]`, nil)

	// Get and display the selected option
	response := interpreter.MenuResponse()
	if response == "" {
		// Use interpreter to render error message with styling
		interpreter.ExecString("\n[bold][red]No valid option selected or no input received.[reset]\n", nil)
		os.Exit(1)
	}

	// Map option IDs to descriptions for display
	optionMap := map[string]string{
		"a": "View Messages",
		"b": "Read Email",
		"c": "File Downloads",
		"d": "User Directory",
		"e": "Chat Room",
		"q": "Quit / Logout",
	}

	description, ok := optionMap[response]
	if !ok {
		description = "Unknown option"
	}

	// Use interpreter to render the response message with styling
	msg := fmt.Sprintf("\n[bold][green]You selected: %s (%s)[reset]\n\n", strings.ToUpper(response), description)
	interpreter.ExecString(msg, nil)
}
