package globals

import (
	"fmt"
	t "github.com/faelmori/gastype/types"
	l "github.com/faelmori/logz"
	"os"
	"runtime"
	"syscall"
)

type Environment struct {
	cpuCount int
	memTotal int
	hostname string
	os       string
	kernel   string
}

func NewEnvironment() t.IEnvironment {
	l.DebugCtx("Creating environment", nil)
	cpuCount := runtime.NumCPU()
	memTotal := syscall.Sysinfo_t{}.Totalram
	hostname, hostnameErr := os.Hostname()
	if hostnameErr != nil {
		l.ErrorCtx(fmt.Sprintf("Error getting hostname: %s", hostnameErr.Error()), nil)
		return nil
	}
	os := runtime.GOOS
	kernel := runtime.GOARCH
	return &Environment{
		cpuCount: cpuCount,
		memTotal: int(memTotal),
		hostname: hostname,
		os:       os,
		kernel:   kernel,
	}
}

func (e *Environment) CpuCount() int {
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
			l.ErrorCtx(fmt.Sprintf("Error getting memory info: %s", err.Error()), nil)
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
			l.ErrorCtx(fmt.Sprintf("Error getting hostname: %s", err.Error()), nil)
			return ""
		}
		e.hostname = hostname
	}
	return e.hostname
}

func (e *Environment) Os() string {
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
