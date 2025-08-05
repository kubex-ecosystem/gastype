package main

import "fmt"

type TestConfig struct {
	Debug   bool
	Verbose bool
}

func main() {
	cfg := TestConfig{Debug: true, Verbose: false}

	if cfg.Debug {
		fmt.Println("Debug mode enabled")
	}

	if cfg.Verbose {
		fmt.Println("Verbose mode enabled")
	}

	fmt.Printf("Config: Debug=%t, Verbose=%t\n", cfg.Debug, cfg.Verbose)
}
