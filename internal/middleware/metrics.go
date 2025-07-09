package middleware

import (
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type EndpointStats struct {
	Count    int64           // количество запросов
	TotalDur time.Duration   // суммарное время выполнения
	MaxTime time.Duration
	MinTime time.Duration
}

type Metrics struct {
	TotalRequests int64
	Endpoints     map[string]*EndpointStats
	mu            sync.RWMutex // для потокобезопасности
}

var metrics = Metrics{
	Endpoints: make(map[string]*EndpointStats),
}

func MetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)
		if err != nil {
			c.Error(err)
		}

		duration := time.Since(start)
		endpoint := c.Path()

		metrics.mu.Lock()
		defer metrics.mu.Unlock()

		metrics.TotalRequests++

		if stats, exists := metrics.Endpoints[endpoint]; exists {
			stats.Count++
			stats.TotalDur += duration
			if duration > stats.MaxTime {
				stats.MaxTime = duration
			} else if duration < stats.MinTime {
				stats.MinTime = duration
			}
		} else {
			metrics.Endpoints[endpoint] = &EndpointStats{
				Count:    1,
				TotalDur: duration,
				MaxTime: duration,
				MinTime: duration,
			}
		}

		return nil
	}
}

func GetMetrics() map[string]interface{} {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	result := make(map[string]interface{})
	result["total_requests"] = metrics.TotalRequests

	endpointMetrics := make(map[string]interface{})
	for path, stats := range metrics.Endpoints {
		avgTime := time.Duration(int64(stats.TotalDur) / stats.Count)
		endpointMetrics[path] = map[string]interface{}{
			"count":       stats.Count,
			"total_time":  stats.TotalDur.String(),
			"avg_time":    avgTime.String(),
			"avg_time_ns": avgTime.Nanoseconds(),
			"min_time": stats.MinTime,
			"max_time": stats.MaxTime,
		}
	}

	result["endpoints"] = endpointMetrics
	return result
}