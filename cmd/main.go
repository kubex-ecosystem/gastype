package main

import (
	l "github.com/faelmori/logz"
	"os"
)

func main() {
	l.GetLogger("GasType")
	if err := RegX().Execute(); err != nil {
		os.Exit(1)
	}
}
