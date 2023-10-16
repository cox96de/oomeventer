package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/ringbuf"
	"golang.org/x/sys/unix"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

func main() {
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Panic(err)
	}
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Panicf("loading objects: %v", err)
	}
	defer objs.Close()
	kp, err := link.Kprobe("oom_kill_process", objs.KprobeOomKillProcess, nil)
	if err != nil {
		log.Panicf("opening tracepoint: %s", err)
	}
	defer kp.Close()
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	log.Println("Waiting for events..")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	reader, err := ringbuf.NewReader(objs.bpfMaps.Events)
	if err != nil {
		log.Fatalf("creating perf reader: %s", err)
	}
	go func() {
		defer reader.Close()
		var e bpfEvent
		for {
			record, err := reader.Read()
			if err != nil {
				if errors.Is(err, perf.ErrClosed) {
					break
				}
				log.Printf("failed to read from perf buffer: %s", err)
				continue
			}
			// Parse the ring buffer entry into an event structure.
			if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &e); err != nil {
				log.Printf("parsing perf event: %s", err)
				continue
			}
			log.Printf("pid: %d, filepath: %s, cgroup: %s", e.Pid, unix.ByteSliceToString(e.Comm[:]), unix.ByteSliceToString(e.CgroupName[:]))
		}
		signals <- syscall.SIGTERM
	}()
	<-signals
}
