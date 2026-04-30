package monitor

import (
	"time"

	"uptime-monitor/internal/models"
)

func Hello() string {
	return "hello from monitor"
}

func PerformChecks(urls []string) []models.CheckResult {
	results := make([]models.CheckResult, 0, len(urls))
	now := time.Now().UTC().Format(time.RFC3339)

	for _, url := range urls {
		results = append(results, models.CheckResult{
			URL:        url,
			StatusCode: 200,
			IsUp:       true,
			LatencyMS:  1,
			Timestamp:  now,
		})
	}

	return results
}
