// Package interfaces provides interfaces for environment information.
package interfaces

type IEnvironment interface {
	CPUCount() int
	MemTotal() int
	Hostname() string
	OS() string
	Kernel() string
}
