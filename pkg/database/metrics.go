package database

import (
    "time"
)

func InsertPodMetrics(podName string, namespace string, cpuRequest int64, cpuUsage int64, memRequest int64, memUsage int64, timestamp time.Time) error {
    db := GetDB()
    query := `
        INSERT INTO pod_metrics (
            id,
            timestamp,
            pod_name,
            namespace,
            cpu_request,
            cpu_usage,
            memory_request,
            memory_usage
        ) VALUES (
            gen_random_uuid(),
            $1, $2, $3, $4, $5, $6, $7
        );`
    _, err := db.Exec(query, timestamp, podName, namespace, cpuRequest, cpuUsage, memRequest, memUsage)
    return err
}