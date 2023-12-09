package app

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"

	"github.com/cox96de/oomeventer/bpf"
	"github.com/cox96de/oomeventer/kube"
)

type App struct {
	criClient  *kube.CRIClient
	kubeClient kubernetes.Interface
	namespaces map[string]bool
}

func NewApp(criClient *kube.CRIClient, namespaces []string) *App {
	return &App{criClient: criClient, namespaces: lo.SliceToMap(namespaces, func(item string) (string, bool) {
		return item, true
	})}
}

func (a *App) Run() {
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Panic(err)
	}
	objs := bpf.BPFObjects{}
	if err := bpf.LoadBPFObjects(&objs, nil); err != nil {
		log.Panicf("failed to load objects: %v", err)
	}
	defer objs.Close()
	kprobe, err := link.Kprobe("oom_kill_process", objs.KprobeOomKillProcess, nil)
	if err != nil {
		log.Panicf("failed to open tracepoint: %s", err)
	}
	defer kprobe.Close()
	log.Println("waiting for events.....")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	reader, err := ringbuf.NewReader(objs.BPFMaps.Events)
	if err != nil {
		log.Fatalf("creating perf reader: %s", err)
	}
	go func() {
		defer reader.Close()
		var e bpf.BPFEvent
		for {
			record, err := reader.Read()
			if err != nil {
				if errors.Is(err, perf.ErrClosed) {
					log.Printf("perf buffer closed")
					break
				}
				log.Printf("failed to read from perf buffer: %s", err)
				continue
			}
			// Parse the ring buffer entry into an event structure.
			if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &e); err != nil {
				log.Printf("failed to parse ring buffer event: %s", err)
				continue
			}
			go a.handleOOM(context.Background(), e.Pid)
		}
		signals <- syscall.SIGTERM
	}()
	<-signals
}

func (a *App) handleOOM(ctx context.Context, pid int32) {
	proc := &kube.Process{
		PID: pid,
	}
	cmdlineContent, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		log.Errorf("failed to read cmdline: %+v", err)
	} else {
		proc.CmdLine = strings.Split(string(cmdlineContent), "\x00")
	}
	environContent, err := os.ReadFile(fmt.Sprintf("/proc/%d/environ", pid))
	if err != nil {
		log.Errorf("failed to read environ: %+v", err)
	} else {
		proc.Environ = strings.Split(string(environContent), "\x00")
	}
	log.Infof("oom process: %+v", proc)
	cgroupPaths, err := kube.ParseCgroupFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		log.Errorf("failed to read cgroup: %+v", err)
		return
	}
	cgroupPath := cgroupPaths["memory"]
	containerID := kube.ExtractContainerIDFromCgroupPath(cgroupPath)
	if len(containerID) == 0 {
		log.Errorf("failed to extract container id from cgroup path: %s", cgroupPath)
		return
	}
	containerStatus, err := a.criClient.ContainerStatus(ctx, &runtimeapi.ContainerStatusRequest{
		ContainerId: containerID,
		Verbose:     false,
	})
	if err != nil {
		log.Errorf("failed to get container '%s' status: %+v", containerID, err)
		return
	}
	//podID := containerStatus.Info["sandboxID"]
	podNamespace := containerStatus.Status.Labels["io.kubernetes.pod.namespace"]
	podName := containerStatus.Status.Labels["io.kubernetes.pod.name"]
	log.Infof("OOM process in pod '%s/%s': %+v", podNamespace, podName, proc)
	_, err = a.kubeClient.EventsV1().Events(podNamespace).Create(context.Background(),
		&eventv1.Event{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "oomevent" + strconv.FormatInt(time.Now().Unix(), 10),
			},
			EventTime:           metav1.MicroTime{Time: time.Now()},
			Series:              nil,
			ReportingController: "kubernetes.io/kubelet",
			ReportingInstance:   podName + strconv.FormatInt(time.Now().Unix(), 10),
			Action:              "OOM",
			Reason:              "test-reason",
			Regarding: corev1.ObjectReference{
				Kind:      "Pod",
				Namespace: podNamespace,
				Name:      podName,
			},
			Related:         nil,
			Note:            "",
			Type:            "Warning",
			DeprecatedCount: 0,
		}, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("failed to send event: %+v", err)
	}
}
