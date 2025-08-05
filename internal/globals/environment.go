package globals

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	l "github.com/faelmori/logz"
	t "github.com/rafa-mori/gastype/interfaces"
)

type Environment struct {
	cpuCount int
	memTotal int
	hostname string
	os       string
	kernel   string
}

func NewEnvironment() t.IEnvironment { return &Environment{} }

func (e *Environment) CPUCount() int {
	if e.cpuCount == 0 {
		e.cpuCount = runtime.NumCPU()
	}
	return e.cpuCount
}

func (e *Environment) MemTotal() int {
	if e.memTotal == 0 {
		var mem syscall.Sysinfo_t
		err := syscall.Sysinfo(&mem)
		if err != nil {
			l.Error(fmt.Sprintf("Error getting memory info: %s", err.Error()), nil)
			return 0
		}
		totalRAM := mem.Totalram * uint64(mem.Unit) / (1024 * 1024) // Convertendo para MB
		e.memTotal = int(totalRAM)
	}
	return e.memTotal
}

func (e *Environment) Hostname() string {
	if e.hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			l.Error(fmt.Sprintf("Error getting hostname: %s", err.Error()), nil)
			return ""
		}
		e.hostname = hostname
	}
	return e.hostname
}

func (e *Environment) OS() string {
	if e.os == "" {
		e.os = runtime.GOOS
	}
	return e.os
}

func (e *Environment) Kernel() string {
	if e.kernel == "" {
		e.kernel = runtime.GOARCH
	}
	return e.kernel
}
