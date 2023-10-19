package kube

type Process struct {
	PID        int32    `json:"pid"`
	CgroupName string   `json:"cgroup_name"`
	CmdLine    []string `json:"cmdline"`
	Environ    []string `json:"environ"`
}
