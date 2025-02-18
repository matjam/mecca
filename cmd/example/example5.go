package main

import "github.com/matjam/mecca"

func example5() {
	// Create a new MECCA interpreter.
	interpreter := mecca.NewInterpreter(mecca.WithTemplateRoot("cmd/example"))

	// Example MECCA template loading an external ansi file.
	template := "[ansi pikachu.ans]\n[bold yellow]         " +
		"\"But this wasn't what I voted for!\" - Surprised Pikachu\n"

	// Render the template.
	interpreter.ExecString(template, nil)
}
