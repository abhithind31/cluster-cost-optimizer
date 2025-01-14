package kubernetes

import (
    "context"

    v1 "k8s.io/api/core/v1"
    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListPods(clientset *kubernetes.Clientset) ([]v1.Pod, error) {
    pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    return pods.Items, nil
}

func GetPodResourceRequests(pod v1.Pod) (cpuRequests int64, memRequests int64) {
    cpuRequests = 0
    memRequests = 0
    for _, container := range pod.Spec.Containers {
        cpuQuantity := container.Resources.Requests.Cpu()
        memQuantity := container.Resources.Requests.Memory()
        cpuRequests += cpuQuantity.MilliValue()
        memRequests += memQuantity.Value()
    }
    return
}