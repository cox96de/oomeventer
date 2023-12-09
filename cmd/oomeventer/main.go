package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/cox96de/oomeventer/bpf"
	"github.com/cox96de/oomeventer/kube"
)

func main() {
	var criAddr string
	flag.StringVar(&criAddr, "cri-addr", "/run/containerd/containerd.sock", "cri address")
	flag.Parse()
	criClient, err := kube.NewCRIClient(criAddr)
	if err != nil {
		log.Fatalf("creating cri client: %s", err)
	}
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	log.Println("Waiting for events..")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	listener := bpf.NewListener()
	err = listener.Start()
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
				PID: e.Pid,
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

func getKubernetesClient(kubeConfigPath string) kubernetes.Interface {
	var (
		restConfig *rest.Config
		err        error
	)
	if len(kubeConfigPath) > 0 {
		// Use the given kube config.
		log.Printf("loading kube config from [%s]", kubeConfigPath)
		b, err := os.ReadFile(kubeConfigPath)
		if err != nil {
			log.Fatalf("failed to read kube config [%s]: %v", kubeConfigPath, err)
		}
		config, err := clientcmd.NewClientConfigFromBytes(b)
		if err != nil {
			log.Fatalf("failed to parse kube config [%s]: %v", kubeConfigPath, err)
		}
		restConfig, err = config.ClientConfig()
		if err != nil {
			log.Fatalf("failed to get rest config from kube config [%s]: %v", kubeConfigPath, err)
		}
	} else {
		// Use the in-cluster config.
		log.Printf("loading kube config from in-cluster files")
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("failed to get in-cluster config: %v", err)
		}
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("failed to create rest client: %v", err)
	}
	return client
}
