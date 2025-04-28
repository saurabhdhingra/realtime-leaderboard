package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsStore holds API metrics information
type MetricsStore struct {
	RequestCount       map[string]int64
	ErrorCount         map[string]int64
	ResponseTimes      map[string]time.Duration
	RequestCountByPath map[string]map[string]int64
	mu                 sync.RWMutex
}

// Global metrics store
var Metrics = &MetricsStore{
	RequestCount:       make(map[string]int64),
	ErrorCount:         make(map[string]int64),
	ResponseTimes:      make(map[string]time.Duration),
	RequestCountByPath: make(map[string]map[string]int64),
}

// GetMetrics returns a copy of the current metrics
func (ms *MetricsStore) GetMetrics() gin.H {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Create deep copies of all metrics
	requestCount := make(map[string]int64)
	errorCount := make(map[string]int64)
	responseTimes := make(map[string]time.Duration)
	requestCountByPath := make(map[string]map[string]int64)

	for method, count := range ms.RequestCount {
		requestCount[method] = count
	}

	for method, count := range ms.ErrorCount {
		errorCount[method] = count
	}

	for method, duration := range ms.ResponseTimes {
		responseTimes[method] = duration
	}

	for method, paths := range ms.RequestCountByPath {
		methodMap := make(map[string]int64)
		for path, count := range paths {
			methodMap[path] = count
		}
		requestCountByPath[method] = methodMap
	}

	// Calculate average response times
	avgResponseTimes := make(map[string]float64)
	for method, totalTime := range responseTimes {
		count := requestCount[method]
		if count > 0 {
			avgResponseTimes[method] = float64(totalTime) / float64(count) / float64(time.Millisecond)
		}
	}

	// Total requests
	var totalRequests int64
	for _, count := range requestCount {
		totalRequests += count
	}

	// Total errors
	var totalErrors int64
	for _, count := range errorCount {
		totalErrors += count
	}

	return gin.H{
		"total_requests":         totalRequests,
		"total_errors":           totalErrors,
		"error_rate":             float64(totalErrors) / float64(totalRequests) * 100,
		"requests_by_method":     requestCount,
		"errors_by_method":       errorCount,
		"avg_response_time_ms":   avgResponseTimes,
		"requests_by_path":       requestCountByPath,
	}
}

// TrackRequest tracks a request in the metrics store
func (ms *MetricsStore) TrackRequest(method, path string, duration time.Duration, isError bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Track request count
	ms.RequestCount[method]++

	// Track error count if request resulted in error
	if isError {
		ms.ErrorCount[method]++
	}

	// Track response time
	ms.ResponseTimes[method] += duration

	// Track request count by path
	if ms.RequestCountByPath[method] == nil {
		ms.RequestCountByPath[method] = make(map[string]int64)
	}
	ms.RequestCountByPath[method][path]++
}

// MetricsMiddleware tracks request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		duration := time.Since(start)

		// Get request method and path
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		// Check if request resulted in error
		isError := c.Writer.Status() >= 400

		// Track request
		Metrics.TrackRequest(method, path, duration, isError)
	}
}

// MetricsHandler returns the current metrics
func MetricsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, Metrics.GetMetrics())
}

// PrintMetrics prints the current metrics to stdout
func PrintMetrics() {
	metrics := Metrics.GetMetrics()
	
	fmt.Println("\n=== API Metrics ===")
	fmt.Printf("Total Requests: %d\n", metrics["total_requests"])
	fmt.Printf("Total Errors: %d\n", metrics["total_errors"])
	fmt.Printf("Error Rate: %.2f%%\n", metrics["error_rate"])
	
	fmt.Println("\nRequests by Method:")
	for method, count := range metrics["requests_by_method"].(map[string]int64) {
		fmt.Printf("  %s: %d\n", method, count)
	}
	
	fmt.Println("\nAverage Response Time (ms):")
	for method, time := range metrics["avg_response_time_ms"].(map[string]float64) {
		fmt.Printf("  %s: %.2f ms\n", method, time)
	}
	
	fmt.Println("\nTop Paths:")
	for method, paths := range metrics["requests_by_path"].(map[string]map[string]int64) {
		fmt.Printf("  %s:\n", method)
		// Find top 5 paths
		type pathCount struct {
			path  string
			count int64
		}
		var topPaths []pathCount
		for path, count := range paths {
			topPaths = append(topPaths, pathCount{path, count})
		}
		// Sort by count (simple bubble sort for this small dataset)
		for i := 0; i < len(topPaths)-1; i++ {
			for j := 0; j < len(topPaths)-i-1; j++ {
				if topPaths[j].count < topPaths[j+1].count {
					topPaths[j], topPaths[j+1] = topPaths[j+1], topPaths[j]
				}
			}
		}
		// Print top 5 or all if less than 5
		count := 5
		if len(topPaths) < count {
			count = len(topPaths)
		}
		for i := 0; i < count; i++ {
			fmt.Printf("    %s: %d\n", topPaths[i].path, topPaths[i].count)
		}
	}
	
	fmt.Println("==================")
} 