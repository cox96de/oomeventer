package kube

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ParseCgroupFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseCgroupFromReader(f)
}

func ParseCgroupFromReader(r io.Reader) (map[string]string, error) {
	var (
		cgroups = make(map[string]string)
		s       = bufio.NewScanner(r)
	)
	for s.Scan() {
		var (
			text  = s.Text()
			parts = strings.SplitN(text, ":", 3)
		)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid cgroup entry: %q", text)
		}
		for _, subs := range strings.Split(parts[1], ",") {
			if subs == "" {
				continue
			}
			cgroups[subs] = parts[2]
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return cgroups, nil
}

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
