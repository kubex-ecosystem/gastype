package main

import "fmt"

type Config struct {
	Debug   bool
	Verbose bool
	Logging bool
}

func main() {
	cfg := Config{
		Debug:   true,
		Verbose: false,
		Logging: true,
	}

	if cfg.Debug {
		fmt.Println("Debug mode enabled")
	}
	
	if cfg.Verbose {
		fmt.Println("Verbose mode enabled")
	}
	
	if cfg.Logging {
		fmt.Println("Logging enabled")
	}
	
	fmt.Printf("Config: Debug=%t, Verbose=%t, Logging=%t\n", cfg.Debug, cfg.Verbose, cfg.Logging)
}
