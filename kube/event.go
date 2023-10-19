package kube

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"log"
)

const component = "oom-eventer"

// SendOOMEvent sends an OOM event to the kubernetes API.
func SendOOMEvent(clientset kubernetes.Interface, proc *Process, namespace, podName string) error {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartEventWatcher(func(event *v1.Event) {
		log.Printf("recieve event from watcher: %+v", event)
	})
	eventBroadcaster.StartLogging(klog.Infof)
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: component})
	object, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	process, err := json.Marshal(proc)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal process")
	}
	eventRecorder.AnnotatedEventf(object, map[string]string{"process": string(process)}, v1.EventTypeWarning, "OOM", "subprocess oom")
	return nil
}
