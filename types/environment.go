package types

type IEnvironment interface {
	CpuCount() int
	MemTotal() int
	Hostname() string
	Os() string
	Kernel() string
}
