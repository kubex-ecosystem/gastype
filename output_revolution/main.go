package main

import "fmt"

const (
	FlagConfig_Debug	uint64	= 1 << 0
	FlagConfig_Verbose	uint64	= 1 << 1
	FlagConfig_Logging	uint64	= 1 << 2
)

type ConfigFlags struct {
	flags uint64
}

func main() {
	cfg := ConfigFlags{flags: 5,
	}

	if cfg.flags&FlagConfig_Debug != 0 {
		fmt.Println(string([]byte{68, 101, 98, 117, 103, 32, 109, 111, 100, 101, 32, 101, 110, 97, 98, 108, 101, 100}))
	}

	if cfg.flags&FlagConfig_Verbose != 0 {
		fmt.Println(string([]byte{86, 101, 114, 98, 111, 115, 101, 32, 109, 111, 100, 101, 32, 101, 110, 97, 98, 108, 101, 100}))
	}

	if cfg.flags&FlagConfig_Logging != 0 {
		fmt.Println(string([]byte{76, 111, 103, 103, 105, 110, 103, 32, 101, 110, 97, 98, 108, 101, 100}))
	}

	fmt.Printf(string([]byte{67, 111, 110, 102, 105, 103, 58, 32, 68, 101, 98, 117, 103, 61, 37, 116, 44, 32, 86, 101, 114, 98, 111, 115, 101, 61, 37, 116, 44, 32, 76, 111, 103, 103, 105, 110, 103, 61, 37, 116, 10}), (cfg.flags&FlagConfig_Debug) != 0, (cfg.flags&FlagConfig_Verbose) != 0, (cfg.flags&FlagConfig_Logging) != 0)
}
