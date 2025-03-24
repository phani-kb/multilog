// Package main provides the entry point for the multilog application.
package main

import (
	"flag"
	"log"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yml", "Path to configuration file")
	flag.Parse()

	if err := run(*configPath); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(_ string) error {
	return nil
}
