package main

import "fmt"

type Config struct {
	Debug			bool
	Verbose			bool
	Logging			bool
	FirstConvertedBool	bool
	SecondConvertedBool	bool
	ThirdConvertedBool	bool
}

var (
	TextTypePlain			= string([]byte{112, 108, 97, 105, 110})
	CustomTypeTest	LogType		= LogType([]byte{99, 117, 115, 116, 111, 109})
	CustomLevelTest	LogLevel	= 42

	StringVarTest	string
	StringVarTest2	string	= string([]byte{83, 111, 109, 101, 86, 97, 108, 117, 101})
	st1		string	= "ok"
)

type LogType string
type LogLevel int

const (
	// LogTypeDebug is the log type for debug messages.
	LogTypeDebug	LogType	= "debug"
	// LogTypeNotice is the log type for notice messages.
	LogTypeNotice	LogType	= "notice"
	// LogTypeInfo is the log type for informational messages.
	LogTypeInfo	LogType	= "info"
	// LogTypeWarn is the log type for warning messages.
	LogTypeWarn	LogType	= "warn"
	// LogTypeError is the log type for error messages.
	LogTypeError	LogType	= "error"
	// LogTypeFatal is the log type for fatal error messages.
	LogTypeFatal	LogType	= "fatal"
)

const (
	// LogLevelDebug 0
	LogLevelDebug	LogLevel	= iota
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
		fmt.Println(string([]byte{70, 105, 114, 115, 116, 67, 111, 110, 118, 101, 114, 116, 101, 100, 66, 111, 111, 108, 32, 105, 115, 32, 116, 114, 117, 101}))
	}
	if cfg.SecondConvertedBool {
		fmt.Println(string([]byte{83, 101, 99, 111, 110, 100, 67, 111, 110, 118, 101, 114, 116, 101, 100, 66, 111, 111, 108, 32, 105, 115, 32, 116, 114, 117, 101}))
	}
	if cfg.ThirdConvertedBool {
		fmt.Println(string([]byte{84, 104, 105, 114, 100, 67, 111, 110, 118, 101, 114, 116, 101, 100, 66, 111, 111, 108, 32, 105, 115, 32, 116, 114, 117, 101}))
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
		fmt.Println(string([]byte{70, 105, 114, 115, 116, 32, 97, 114, 103, 117, 109, 101, 110, 116, 32, 105, 115, 32, 116, 114, 117, 101}))
	}
	if second {
		fmt.Println(string([]byte{83, 101, 99, 111, 110, 100, 32, 97, 114, 103, 117, 109, 101, 110, 116, 32, 105, 115, 32, 116, 114, 117, 101}))
	}
	if third {
		fmt.Println(string([]byte{84, 104, 105, 114, 100, 32, 97, 114, 103, 117, 109, 101, 110, 116, 32, 105, 115, 32, 116, 114, 117, 101}))
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
		Debug:		true,
		Verbose:	false,
		Logging:	true,
	}

	fmt.Println(string([]byte{45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45}))
	fmt.Println(string([]byte{84, 101, 115, 116, 105, 110, 103, 32, 115, 116, 114, 105, 110, 103, 32, 111, 98, 102, 117, 115, 99, 97, 116, 101, 100, 32, 118, 97, 108, 117, 101, 115, 58}))
	testStringObfuscatedValues()

	fmt.Println(string([]byte{45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45}))
	fmt.Println(string([]byte{84, 101, 115, 116, 105, 110, 103, 32, 99, 111, 110, 115, 116, 32, 115, 116, 114, 105, 110, 103, 32, 111, 98, 102, 117, 115, 99, 97, 116, 101, 100, 32, 118, 97, 108, 117, 101, 115, 58}))
	testConstStringObfuscatedValues()

	fmt.Println(string([]byte{45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45}))
	fmt.Println(string([]byte{84, 101, 115, 116, 105, 110, 103, 32, 98, 111, 111, 108, 101, 97, 110, 32, 99, 111, 110, 118, 101, 114, 116, 101, 100, 32, 118, 97, 108, 117, 101, 115, 58}))

	if cfg.Debug {
		fmt.Println(string([]byte{68, 101, 98, 117, 103, 32, 109, 111, 100, 101, 32, 101, 110, 97, 98, 108, 101, 100}))
	}

	if cfg.Verbose {
		fmt.Println(string([]byte{86, 101, 114, 98, 111, 115, 101, 32, 109, 111, 100, 101, 32, 101, 110, 97, 98, 108, 101, 100}))
	}

	if cfg.Logging {
		fmt.Println(string([]byte{76, 111, 103, 103, 105, 110, 103, 32, 101, 110, 97, 98, 108, 101, 100}))
	}

	useConvertedBools(cfg)

	fmt.Println(string([]byte{45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45}))
	fmt.Println(string([]byte{84, 101, 115, 116, 105, 110, 103, 32, 98, 111, 111, 108, 101, 97, 110, 32, 99, 111, 110, 118, 101, 114, 116, 101, 100, 32, 118, 97, 108, 117, 101, 115, 58}))

	a, b, c := returnConvertedBools(cfg)
	fmt.Printf(string([]byte{82, 101, 116, 117, 114, 110, 101, 100, 32, 99, 111, 110, 118, 101, 114, 116, 101, 100, 32, 98, 111, 111, 108, 115, 58, 32, 37, 116, 44, 32, 37, 116, 44, 32, 37, 116, 10}), a, b, c)

	result := checkConvertedBools(cfg)
	fmt.Printf(string([]byte{67, 104, 101, 99, 107, 32, 99, 111, 110, 118, 101, 114, 116, 101, 100, 32, 98, 111, 111, 108, 115, 32, 114, 101, 115, 117, 108, 116, 58, 32, 37, 116, 10}), result)

	fmt.Println(string([]byte{45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45}))
	fmt.Println(string([]byte{84, 101, 115, 116, 105, 110, 103, 32, 98, 111, 111, 108, 101, 97, 110, 32, 114, 101, 99, 101, 105, 118, 101, 100, 32, 118, 97, 108, 117, 101, 115, 58}))

	receiveBoolArgs(cfg.FirstConvertedBool, cfg.SecondConvertedBool, cfg.ThirdConvertedBool)

	fmt.Printf(string([]byte{67, 111, 110, 102, 105, 103, 58, 32, 68, 101, 98, 117, 103, 61, 37, 116, 44, 32, 86, 101, 114, 98, 111, 115, 101, 61, 37, 116, 44, 32, 76, 111, 103, 103, 105, 110, 103, 61, 37, 116, 10}), cfg.Debug, cfg.Verbose, cfg.Logging)

	fmt.Println(string([]byte{45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45}))
	fmt.Println(string([]byte{84, 101, 115, 116, 105, 110, 103, 32, 115, 116, 114, 105, 110, 103, 32, 111, 98, 102, 117, 115, 99, 97, 116, 101, 100, 32, 118, 97, 108, 117, 101, 115, 58}))

	testStringObfuscatedValues()

	fmt.Println(string([]byte{45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45}))
	fmt.Println(string([]byte{84, 101, 115, 116, 105, 110, 103, 32, 99, 111, 110, 115, 116, 32, 115, 116, 114, 105, 110, 103, 32, 111, 98, 102, 117, 115, 99, 97, 116, 101, 100, 32, 118, 97, 108, 117, 101, 115, 58}))

	testConstStringObfuscatedValues()

	fmt.Println(string([]byte{45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45}))
	fmt.Println(string([]byte{69, 110, 100, 32, 111, 102, 32, 116, 101, 115, 116, 115, 46}))
}
