// Package test_project provides a sample configuration for testing.
package main

import (
	"fmt"

	l "github.com/rafa-mori/logz"
)

type Config struct {
	Debug               bool
	Verbose             bool
	Logging             bool
	FirstConvertedBool  bool
	SecondConvertedBool bool
	ThirdConvertedBool  bool
}

var (
	TextTypePlain            = "plain"
	CustomTypeTest  LogType  = "custom"
	CustomLevelTest LogLevel = 42

	StringVarTest  string
	StringVarTest2 string = "SomeValue"
	st1            string = "ok"
)

type LogType string
type LogLevel int

const (
	// LogTypeDebug is the log type for debug messages.
	LogTypeDebug LogType = "debug"
	// LogTypeNotice is the log type for notice messages.
	LogTypeNotice LogType = "notice"
	// LogTypeInfo is the log type for informational messages.
	LogTypeInfo LogType = "info"
	// LogTypeWarn is the log type for warning messages.
	LogTypeWarn LogType = "warn"
	// LogTypeError is the log type for error messages.
	LogTypeError LogType = "error"
	// LogTypeFatal is the log type for fatal error messages.
	LogTypeFatal LogType = "fatal"
)

const (
	// LogLevelDebug 0
	LogLevelDebug LogLevel = iota
	// LogLevelNotice 1
	LogLevelNotice
	// LogLevelInfo 2
	LogLevelInfo
	// LogLevelSuccess 3
	LogLevelSuccess
	// LogLevelWarn 4
	LogLevelWarn
	// LogLevelError 5
	LogLevelError
)

func useConvertedBools(cfg Config) {
	if cfg.FirstConvertedBool {
		fmt.Println("FirstConvertedBool is true")
	}
	if cfg.SecondConvertedBool {
		fmt.Println("SecondConvertedBool is true")
	}
	if cfg.ThirdConvertedBool {
		fmt.Println("ThirdConvertedBool is true")
	}
}

func returnConvertedBools(cfg Config) (bool, bool, bool) {
	return cfg.FirstConvertedBool, cfg.SecondConvertedBool, cfg.ThirdConvertedBool
}

func checkConvertedBools(cfg Config) bool {
	return cfg.FirstConvertedBool && (cfg.SecondConvertedBool || !cfg.ThirdConvertedBool)
}

func receiveBoolArgs(first, second, third bool) {
	if first {
		fmt.Println("First argument is true")
	}
	if second {
		fmt.Println("Second argument is true")
	}
	if third {
		fmt.Println("Third argument is true")
	}
}

func testStringObfuscatedValues() {
	fmt.Println(StringVarTest)
	fmt.Println(StringVarTest2)
}

func testConstStringObfuscatedValues() {
	fmt.Println(st1)
	fmt.Println(TextTypePlain)
	fmt.Println(CustomTypeTest)
	fmt.Println(LogTypeDebug)
	fmt.Println(LogTypeNotice)
	fmt.Println(LogTypeInfo)
	fmt.Println(LogTypeWarn)
	fmt.Println(LogTypeError)
	fmt.Println(LogTypeFatal)

	fmt.Println(CustomLevelTest)
	fmt.Println(LogLevelDebug)
	fmt.Println(LogLevelNotice)
	fmt.Println(LogLevelInfo)
	fmt.Println(LogLevelSuccess)
	fmt.Println(LogLevelWarn)
	fmt.Println(LogLevelError)
}

func main() {
	cfg := Config{
		Debug:   true,
		Verbose: false,
		Logging: true,
	}

	fmt.Println("-------------------------------")
	fmt.Println("Testing string obfuscated values:")
	testStringObfuscatedValues()

	fmt.Println("-------------------------------")
	fmt.Println("Testing const string obfuscated values:")
	testConstStringObfuscatedValues()

	fmt.Println("-------------------------------")
	fmt.Println("Testing boolean converted values:")

	if cfg.Debug {
		fmt.Println("Debug mode enabled")
	}

	if cfg.Verbose {
		fmt.Println("Verbose mode enabled")
	}

	if cfg.Logging {
		fmt.Println("Logging enabled")
	}

	useConvertedBools(cfg)

	fmt.Println("-------------------------------")
	fmt.Println("Testing boolean converted values:")

	a, b, c := returnConvertedBools(cfg)
	fmt.Printf("Returned converted bools: %t, %t, %t\n", a, b, c)

	result := checkConvertedBools(cfg)
	fmt.Printf("Check converted bools result: %t\n", result)

	fmt.Println("-------------------------------")
	fmt.Println("Testing boolean received values:")

	receiveBoolArgs(cfg.FirstConvertedBool, cfg.SecondConvertedBool, cfg.ThirdConvertedBool)

	fmt.Printf("Config: Debug=%t, Verbose=%t, Logging=%t\n", cfg.Debug, cfg.Verbose, cfg.Logging)

	fmt.Println("-------------------------------")
	fmt.Println("Testing string obfuscated values:")

	testStringObfuscatedValues()

	fmt.Println("-------------------------------")
	fmt.Println("Testing const string obfuscated values:")

	testConstStringObfuscatedValues()

	fmt.Println("-------------------------------")
	fmt.Println("End of tests.")

	l.Info("All tests completed successfully.")
}
