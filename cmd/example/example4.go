package main

import (
	"fmt"
	"os"

	// added for timing
	"github.com/matjam/mecca"
)

func example4() {
	// Create a new MECCA interpreter.
	interpreter := mecca.NewInterpreter(os.Stdout)

	// Example MECCA template using a registered token and standard tokens.
	template := "[cls bold #255]Standard 4-bit Colors:[reset]\n"
	template += "(note, colors may vary depending on terminal)\n"

	for i := 0; i < 16; i++ {
		if i%8 == 0 {
			template += "\n"
		}
		template += fmt.Sprintf("[#%d] %3v ███", i, i)
	}

	template += "\n\n[bold #255]Standard 8-bit Colors:[reset]\n"

	for i := 0; i < 240; i++ {
		if i%6 == 0 {
			template += "\n"
		}
		template += fmt.Sprintf("[#255]%4v [#%d]███", i+16, i+16)
	}

	template += "\n"

	// Render the template.
	interpreter.Render(template)
}
