package kube

import (
	"context"
	"testing"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func TestNewCRIClient(t *testing.T) {
	client, err := NewCRIClient("/run/containerd/containerd.sock")
	if err != nil {
		panic(err)
	}
	status, err := client.ContainerStatus(context.Background(), &runtimeapi.ContainerStatusRequest{
		ContainerId: "a59e5b84dfa59",
		Verbose:     false,
	})
	if err != nil {
		panic(err)
	}
	t.Logf("%+v", status)
}
