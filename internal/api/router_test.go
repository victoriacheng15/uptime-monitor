package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"uptime-monitor/internal/models"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantAllow  string
		wantJSON   bool
		wantHealth models.HealthResponse
		wantChecks bool
		envTargets string
	}{
		{
			name:       "health returns healthy response",
			method:     http.MethodGet,
			path:       "/health",
			wantStatus: http.StatusOK,
			wantJSON:   true,
			wantHealth: models.HealthResponse{Status: "healthy"},
		},
		{
			name:       "health rejects unsupported method",
			method:     http.MethodPost,
			path:       "/health",
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodGet,
		},
		{
			name:       "check accepts post",
			method:     http.MethodPost,
			path:       "/check",
			wantStatus: http.StatusOK,
			wantJSON:   true,
			wantChecks: true,
			envTargets: "https://site-a.example,https://site-b.example",
		},
		{
			name:       "check rejects unsupported method",
			method:     http.MethodGet,
			path:       "/check",
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodPost,
		},
		{
			name:       "check fails when targets are missing",
			method:     http.MethodPost,
			path:       "/check",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "missing route returns not found",
			method:     http.MethodGet,
			path:       "/missing",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			t.Setenv("MONITOR_TARGETS", tt.envTargets)

			if tt.wantChecks {
				previous := performChecks
				performChecks = func(_ context.Context, urls []string) []models.CheckResult {
					return []models.CheckResult{
						{
							URL:       urls[0],
							IsUp:      true,
							Timestamp: "2026-04-30T12:00:00Z",
						},
					}
				}
				t.Cleanup(func() {
					performChecks = previous
				})
			}

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantAllow != "" {
				if got := rec.Header().Get("Allow"); got != tt.wantAllow {
					t.Fatalf("expected Allow header %q, got %q", tt.wantAllow, got)
				}
			}

			if !tt.wantJSON {
				return
			}

			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("expected content type application/json, got %q", got)
			}

			if tt.wantChecks {
				var response []models.CheckResult
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("decode response: %v", err)
				}

				if len(response) == 0 {
					t.Fatal("expected at least one check result")
				}

				if response[0].URL != "https://site-a.example" {
					t.Fatalf("expected first checked url, got %q", response[0].URL)
				}
				return
			}

			var response models.HealthResponse
			if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if response.Status != tt.wantHealth.Status {
				t.Fatalf("expected status %q, got %q", tt.wantHealth.Status, response.Status)
			}

			if _, err := time.Parse(time.RFC3339, response.Timestamp); err != nil {
				t.Fatalf("expected RFC3339 timestamp, got %q: %v", response.Timestamp, err)
			}
		})
	}
}

func TestMonitorTargets(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want []string
	}{
		{
			name: "returns nil when unset",
		},
		{
			name: "splits comma-separated urls",
			env:  "https://site-a.example,https://site-b.example",
			want: []string{
				"https://site-a.example",
				"https://site-b.example",
			},
		},
		{
			name: "trims whitespace and skips empty entries",
			env:  " https://site-a.example, , https://site-b.example ",
			want: []string{
				"https://site-a.example",
				"https://site-b.example",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONITOR_TARGETS", tt.env)

			got := monitorTargets()
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d targets, got %d", len(tt.want), len(got))
			}

			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("target %d: expected %q, got %q", i, tt.want[i], got[i])
				}
			}
		})
	}
}
