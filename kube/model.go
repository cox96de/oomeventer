package kube

// Process represents a process information.
type Process struct {
	PID     int32    `json:"pid"`
	CmdLine []string `json:"cmdline"`
	Environ []string `json:"environ"`
}
