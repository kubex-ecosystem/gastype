package main

import "fmt"

// Simple test struct
type Config struct {
	Debug   bool
	Verbose bool
}

func main() {
	cfg := Config{Debug: true, Verbose: false}
	
	if cfg.Debug {
		fmt.Println("Debug mode enabled")
	}
	
	if cfg.Verbose {
		fmt.Println("Verbose mode enabled") 
	}
	
	fmt.Printf("Config: Debug=%t, Verbose=%t\n", cfg.Debug, cfg.Verbose)
}
