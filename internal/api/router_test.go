package api

import (
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
