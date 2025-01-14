package analysis

import (
    "fmt"
    "time"

    "github.com/abhithind31/cluster-cost-optimizer/pkg/cost"
    "github.com/abhithind31/cluster-cost-optimizer/pkg/database"
)

type Recommendation struct {
    PodName           string    `json:"pod_name"`
    Namespace         string    `json:"namespace"`
    RecommendedCPU    int64     `json:"recommended_cpu_request"`
    RecommendedMemory int64     `json:"recommended_memory_request"`
    PotentialSavings  float64   `json:"potential_savings"`
    Timestamp         time.Time `json:"timestamp"`
}

func GenerateRecommendations() error {
    db := database.GetDB()

    // Compute average usage over the past 24 hours
    query := `
        WITH usage_data AS (
            SELECT
                pod_name,
                namespace,
                AVG(cpu_usage) AS avg_cpu_usage,
                AVG(cpu_request) AS avg_cpu_request,
                AVG(memory_usage) AS avg_mem_usage,
                AVG(memory_request) AS avg_mem_request
            FROM pod_metrics
            WHERE timestamp > NOW() - INTERVAL '24 hours'
            GROUP BY pod_name, namespace
        )
        INSERT INTO recommendations (
            id,
            pod_name,
            namespace,
            recommended_cpu_request,
            recommended_memory_request,
            potential_savings,
            timestamp
        )
        SELECT
            gen_random_uuid(),
            pod_name,
            namespace,
            LEAST(avg_cpu_usage * 1.2, avg_cpu_request) AS recommended_cpu_request,
            LEAST(avg_mem_usage * 1.2, avg_mem_request) AS recommended_memory_request,
            ( (avg_cpu_request - LEAST(avg_cpu_usage * 1.2, avg_cpu_request)) * $1 ) +
            ( (avg_mem_request - LEAST(avg_mem_usage * 1.2, avg_mem_request)) * $2 ) AS potential_savings,
            NOW()
        FROM usage_data
        WHERE (avg_cpu_usage / NULLIF(avg_cpu_request, 0) < 0.5)
           OR (avg_mem_usage / NULLIF(avg_mem_request, 0) < 0.5)
        ON CONFLICT (pod_name, namespace) DO UPDATE SET
            recommended_cpu_request = EXCLUDED.recommended_cpu_request,
            recommended_memory_request = EXCLUDED.recommended_memory_request,
            potential_savings = EXCLUDED.potential_savings,
            timestamp = EXCLUDED.timestamp;
    `

    cpuPricePerMillicore := cost.CpuPricePerMillicorePerHour * 24  // Convert to daily price
    memPricePerByte := cost.MemPricePerBytePerHour * 24  // Convert to daily price

    _, err := db.Exec(query, cpuPricePerMillicore, memPricePerByte)
    if err != nil {
        return fmt.Errorf("failed to insert recommendations: %v", err)
    }
    return nil
}

func GetRecommendations() ([]Recommendation, error) {
    db := database.GetDB()
    query := `
        SELECT
            pod_name,
            namespace,
            recommended_cpu_request,
            recommended_memory_request,
            potential_savings,
            timestamp
        FROM recommendations
        WHERE timestamp > NOW() - INTERVAL '1 day';
    `
    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to query recommendations: %v", err)
    }
    defer rows.Close()

    var recommendations []Recommendation
    for rows.Next() {
        var rec Recommendation
        err := rows.Scan(&rec.PodName, &rec.Namespace, &rec.RecommendedCPU, &rec.RecommendedMemory, &rec.PotentialSavings, &rec.Timestamp)
        if err != nil {
            return nil, err
        }
        recommendations = append(recommendations, rec)
    }
    return recommendations, nil
}