package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <example>")
		fmt.Println("Available examples:")
		fmt.Println("  basic  - Run basic example")
		fmt.Println("  config - Run config example")
		fmt.Println("  sll - Run single letter level example")
		fmt.Println("  mis - Run mis usage example")
		return
	}

	switch os.Args[1] {
	case "basic":
		basicExample()
	case "config":
		configExample()
	case "sll":
		singleLetterLevel()
	case "mis":
		misUsage()
	default:
		fmt.Printf("Unknown example: %s\n", os.Args[1])
		fmt.Println("Available examples: basic, config")
	}
}
