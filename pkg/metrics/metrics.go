package metrics

import (
    "context"

    metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
    metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodMetrics = metricsv1beta1.PodMetrics

func GetPodMetrics(metricsClient *metricsclientset.Clientset) ([]metricsv1beta1.PodMetrics, error) {
    podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses("").List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        return nil, err
    }
    return podMetricsList.Items, nil
}