package monitor

import (
	"context"
	"net/http"
	"time"

	"uptime-monitor/internal/models"
)

const defaultTimeout = 10 * time.Second

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Monitor struct {
	client HTTPClient
	now    func() time.Time
}

func Hello() string {
	return "hello from monitor"
}

func New(client HTTPClient) Monitor {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}

	return Monitor{
		client: client,
		now:    time.Now,
	}
}

func (m Monitor) PerformChecks(ctx context.Context, urls []string) []models.CheckResult {
	results := make([]models.CheckResult, 0, len(urls))

	for _, url := range urls {
		results = append(results, m.Check(ctx, url))
	}

	return results
}

func (m Monitor) Check(ctx context.Context, url string) models.CheckResult {
	start := m.now()
	result := models.CheckResult{
		URL:       url,
		Timestamp: start.UTC().Format(time.RFC3339),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	resp, err := m.client.Do(req)
	result.LatencyMS = int(m.now().Sub(start).Milliseconds())
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.IsUp = resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest

	return result
}

func PerformChecks(urls []string) []models.CheckResult {
	return New(nil).PerformChecks(context.Background(), urls)
}
