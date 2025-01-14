package main

import (
    "fmt"
    "os"
    "time"

    mykube "github.com/abhithind31/cluster-cost-optimizer/pkg/kubernetes"
    "github.com/abhithind31/cluster-cost-optimizer/pkg/database"
	mymetrics "github.com/abhithind31/cluster-cost-optimizer/pkg/metrics"
    "github.com/abhithind31/cluster-cost-optimizer/pkg/analysis"
    "github.com/gin-gonic/gin"
    metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1beta1"
    k8sclient "k8s.io/client-go/kubernetes"
)

func main() {
    // Initialize database
    err := database.InitDB()
    if err != nil {
        panic(err.Error())
    }

    // Get Kubernetes client and config
    clientset, config, err := mykube.GetClientSet()
    if err != nil {
        panic(err.Error())
    }

    // Get Metrics client using the config
    metricsClientset, err := metricsclientset.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }

    // Start data collection
    go startDataCollection(clientset, metricsClientset)

    // Start analysis
    go startAnalysis()

    // Set up API routes
    router := gin.Default()
    router.GET("/api/recommendations", getRecommendations)

    // Start the server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    router.Run(":" + port)
}

func startDataCollection(clientset *k8sclient.Clientset, metricsClientset *metricsclientset.Clientset) {
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            collectAndStoreData(clientset, metricsClientset)
        }
    }
}

func collectAndStoreData(clientset *k8sclient.Clientset, metricsClientset *metricsclientset.Clientset) {
    // Fetch all pods
    pods, err := mykube.ListPods(clientset)
    if err != nil {
        fmt.Println("Error fetching pods:", err)
        return
    }

    // Fetch pod metrics
    podMetricsList, err := mymetrics.GetPodMetrics(metricsClientset)
    if err != nil {
        fmt.Println("Error fetching pod metrics:", err)
        return
    }

    // Map metrics by namespace/podName for quick access
    podMetricsMap := make(map[string]metricsapi.PodMetrics)
    for _, podMetric := range podMetricsList {
        key := fmt.Sprintf("%s/%s", podMetric.Namespace, podMetric.Name)
        podMetricsMap[key] = podMetric
    }

    // Current timestamp
    timestamp := time.Now().UTC()

    // Iterate over pods and collect data
    for _, pod := range pods {
        cpuRequests, memRequests := mykube.GetPodResourceRequests(pod)
        key := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

        // Get usage from metrics
        podMetric, exists := podMetricsMap[key]
        if !exists {
            // Metrics for this pod might not be available yet
            continue
        }

        cpuUsage := int64(0)
        memUsage := int64(0)
        for _, container := range podMetric.Containers {
            cpuQuantity := container.Usage.Cpu()
            memQuantity := container.Usage.Memory()
            cpuUsage += cpuQuantity.MilliValue()
            memUsage += memQuantity.Value()
        }

        // Insert metrics into the database
        err := database.InsertPodMetrics(
            pod.Name,
            pod.Namespace,
            cpuRequests,
            cpuUsage,
            memRequests,
            memUsage,
            timestamp,
        )
        if err != nil {
            fmt.Println("Error inserting pod metrics:", err)
            continue
        }
    }
}

func startAnalysis() {
    ticker := time.NewTicker(60 * time.Second) // Run analysis every hour
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            err := analysis.GenerateRecommendations()
            if err != nil {
                fmt.Println("Error generating recommendations:", err)
            }
        }
    }
}

func getRecommendations(c *gin.Context) {
    recommendations, err := analysis.GetRecommendations()
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, recommendations)
}