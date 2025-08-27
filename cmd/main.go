package main

import (
	"github.com/rafa-mori/gastype/internal/module"
	gl "github.com/rafa-mori/gastype/internal/module/logger"
)

// main initializes the logger and creates a new GoBE instance.
func main() {
	if err := module.RegX().Command().Execute(); err != nil {
		gl.Log("fatal", err.Error())
	}
}
