package main

import (
	"errors"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"github.com/cox96de/oomeventer/bpf"
)

func main() {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	log.Println("Waiting for events..")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	listener := bpf.NewListener()
	err := listener.Start()
	if err != nil {
		log.Fatalf("creating perf reader: %s", err)
	}
	defer listener.Close()
	go func() {
		for {
			e, err := listener.Poll()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				log.Printf("failed to read from perf buffer: %s", err)
				continue
			}
			log.Printf("pid: %d, filepath: %s, cgroup: %s", e.Pid, unix.ByteSliceToString(e.Comm[:]),
				unix.ByteSliceToString(e.CgroupName[:]))
			proc := &kube.Process{
				PID:        e.Pid,
				CgroupName: unix.ByteSliceToString(e.CgroupName[:]),
			}
			cmdlineContent, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", e.Pid))
			if err != nil {
				log.Printf("failed to read cmdline: %+v", err)
			} else {
				proc.CmdLine = strings.Split(string(cmdlineContent), "\x00")
			}
			environContent, err := os.ReadFile(fmt.Sprintf("/proc/%d/environ", e.Pid))
			if err != nil {
				log.Printf("failed to read environ: %+v", err)
			} else {
				proc.Environ = strings.Split(string(environContent), "\x00")
			}
			log.Printf("oom process: %+v", proc)
		}
		signals <- syscall.SIGTERM
	}()
	<-signals
}
