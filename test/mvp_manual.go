package main

import "fmt"

const (
	FlagTestConfig_Debug   uint64 = 1 << 0
	FlagTestConfig_Verbose uint64 = 1 << 1
)

type TestConfigFlags struct {
	flags uint64
}

func main() {
	cfg := TestConfigFlags{flags: 1} // Debug = true, Verbose = false

	if cfg.flags&FlagTestConfig_Debug != 0 {
		fmt.Println("Debug mode enabled")
	}

	if cfg.flags&FlagTestConfig_Verbose != 0 {
		fmt.Println("Verbose mode enabled")
	}

	debugActive := cfg.flags&FlagTestConfig_Debug != 0
	verboseActive := cfg.flags&FlagTestConfig_Verbose != 0
	fmt.Printf("Config: Debug=%t, Verbose=%t\n", debugActive, verboseActive)
}
