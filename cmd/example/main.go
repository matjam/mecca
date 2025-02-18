package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mectest <exanmple number>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "1":
		example1()
	case "2":
		example2()
	case "3":
		example3()
	case "4":
		example4()
	case "5":
		example5()
	default:
		fmt.Println("Usage: mectest <exanmple number>")
		os.Exit(1)
	}

	os.Exit(0)
}
