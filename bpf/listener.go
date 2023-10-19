package bpf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"io"
	"log"
)

type Listener struct {
	objs   *BPFObjects
	kp     link.Link
	reader *ringbuf.Reader
}

func NewListener() *Listener {
	return &Listener{}
}

func (l *Listener) Start() error {
	if err := rlimit.RemoveMemlock(); err != nil {
		return err
	}
	objs := BPFObjects{}
	if err := LoadBPFObjects(&objs, nil); err != nil {
		log.Panicf("loading objects: %v", err)
	}
	l.objs = &objs
	kp, err := link.Kprobe("oom_kill_process", objs.KprobeOomKillProcess, nil)
	if err != nil {
		log.Panicf("opening tracepoint: %s", err)
	}
	l.kp = kp
	reader, err := ringbuf.NewReader(objs.BPFMaps.Events)
	if err != nil {
		return err
	}
	l.reader = reader
	return nil
}

func (l *Listener) Poll() (*BPFEvent, error) {
	record, err := l.reader.Read()
	if err != nil {
		if errors.Is(err, perf.ErrClosed) {
			return nil, io.EOF
		}
		return nil, err
	}
	var e BPFEvent
	// Parse the ring buffer entry into an event structure.
	if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (l *Listener) Close() error {
	l.reader.Close()
	l.kp.Close()
	l.objs.Close()
	return nil
}
