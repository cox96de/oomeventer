package kube

import (
	"context"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func SendEvent(client kubernetes.Interface, ctx context.Context, proc *Process) error {
	createdEvent, err := client.EventsV1().Events("default").Create(context.Background(),
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
				Namespace: "default",
				Name:      podName,
			},
			Related:         nil,
			Note:            "",
			Type:            "Warning",
			DeprecatedCount: 0,
		}, metav1.CreateOptions{})
}
