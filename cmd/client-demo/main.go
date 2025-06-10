package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cox96de/oomeventer/kube"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig string
	if home := os.Getenv("HOME"); home != "" {
		kubeconfig = home + "/.kube/config"
	}
	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, "kubeconfig file")
	flag.Parse()

	var config *rest.Config
	var err error
	if _, err = os.Stat(kubeconfig); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	watchlist := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"events",
		metav1.NamespaceAll,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		watchlist,
		&v1.Event{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				event := obj.(*v1.Event)
				fmt.Printf("Event Added: %s/%s: %s\n", event.Namespace, event.Name, event.Message)
				if event.ReportingController != kube.Component {
					return
				}
				processBody, ok := event.Annotations["process"]
				if !ok {
					return
				}
				pro := &kube.Process{}
				err := json.Unmarshal([]byte(processBody), &pro)
				if err != nil {
					fmt.Printf("json unmarshal error: %v\n", err)
					return
				}
				fmt.Printf("process: \n")
				fmt.Printf("  command: %s\n", pro.CmdLine)
				fmt.Printf("  envs: %s\n", pro.Environ)
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	close(stop)
}
