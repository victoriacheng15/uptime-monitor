package models

import "time"

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type CheckResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	IsUp       bool   `json:"is_up"`
	LatencyMS  int    `json:"latency_ms"`
	Timestamp  string `json:"timestamp"`
	Error      string `json:"error"`
}

type LatestResponse struct {
	Sites     []CheckResult `json:"sites"`
	UpdatedAt string        `json:"updated_at"`
}

type HistoryResponse struct {
	History map[string][]CheckResult `json:"history"`
}

func Hello() string {
	return "hello from models"
}

func NewHealthResponse(status string) HealthResponse {
	return HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func (r CheckResult) IsSuccessful() bool {
	return r.IsUp && r.StatusCode >= 200 && r.StatusCode < 400 && r.Error == ""
}
