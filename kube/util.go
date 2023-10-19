package kube

import (
	"github.com/containerd/cgroups"
	"path/filepath"
	"strings"
)

// ParseCgroupFile parses the cgroup file and returns the unified cgroup path.
func ParseCgroupFile(path string) (string, error) {
	// `runc` library contains similar logic.
	cgroups, unified, err := cgroups.ParseCgroupFileUnified(path)
	if err != nil {
		return "", err
	}
	if unified != "" {
		return unified, nil
	}
	return cgroups["name=systemd"], nil
}

// ExtractContainerIDFromCgroupPath extracts the container ID from the cgroup path.
// The logic is reference from https://github.com/kubernetes/kubernetes/blob/0c645922edcc06adff43c70c02fb56751364bbb5/pkg/kubelet/stats/cri_stats_provider.go#L955
func ExtractContainerIDFromCgroupPath(cgroupPath string) string {
	// case0 == cgroupfs: "/kubepods/burstable/pod2fc932ce-fdcc-454b-97bd-aadfdeb4c340/9be25294016e2dc0340dd605ce1f57b492039b267a6a618a7ad2a7a58a740f32"
	id := filepath.Base(cgroupPath)

	// case1 == systemd: "/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod2fc932ce_fdcc_454b_97bd_aadfdeb4c340.slice/cri-containerd-aaefb9d8feed2d453b543f6d928cede7a4dbefa6a0ae7c9b990dd234c56e93b9.scope"
	// trim anything before the final '-' and suffix .scope
	systemdSuffix := ".scope"
	if strings.HasSuffix(id, systemdSuffix) {
		id = strings.TrimSuffix(id, systemdSuffix)
		components := strings.Split(id, "-")
		if len(components) > 1 {
			id = components[len(components)-1]
		}
	}
	return id
}
