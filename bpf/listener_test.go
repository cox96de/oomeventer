package bpf

import (
	"context"
	"gotest.tools/v3/assert"
	"os/exec"
	"testing"
	"time"
)

func TestNewListener(t *testing.T) {
	listener := NewListener()
	err := listener.Start()
	assert.NilError(t, err)
	defer listener.Close()
	testImg := "python:3.7"
	_, _ = runCommand([]string{"docker", "pull", testImg})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	result := make(chan error)
	go func() {
		defer close(result)
		poll, err := listener.Poll()
		if err != nil {
			result <- err
			return
		}
		t.Logf("oom happend: %d", poll.Pid)
		assert.Assert(t, poll.Pid > 0)
	}()
	time.Sleep(time.Second)
	go func() {
		output, _ := runCommand([]string{"docker", "run", "--rm", "-m", "10m", testImg, "python", "-c", `a = "111111111111111"
while True:
    a += a
`})
		t.Logf("oom: %s", string(output))
	}()
	select {
	case <-ctx.Done():
		t.Fatal("timeout")
	case err := <-result:
		assert.NilError(t, err)
	}
}

func runCommand(args []string) ([]byte, error) {
	return exec.Command(args[0], args[1:]...).CombinedOutput()
}
