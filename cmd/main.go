package main

import (
	l "github.com/faelmori/gastype/log"
	"os"
)

func main() {
	if err := RegX().Execute(); err != nil {
		l.Error("Error executing root command", err)
		os.Exit(1)
	}
}
