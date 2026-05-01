package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"uptime-monitor/internal/models"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		wantStatus  int
		wantAllow   string
		wantJSON    bool
		wantHealth  models.HealthResponse
		wantChecks  bool
		wantLatest  bool
		wantHistory bool
		envTargets  string
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
			name:       "latest returns stored snapshot",
			method:     http.MethodGet,
			path:       "/latest",
			wantStatus: http.StatusOK,
			wantJSON:   true,
			wantLatest: true,
		},
		{
			name:       "latest rejects unsupported method",
			method:     http.MethodPost,
			path:       "/latest",
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodGet,
		},
		{
			name:        "history returns stored entries",
			method:      http.MethodGet,
			path:        "/history",
			wantStatus:  http.StatusOK,
			wantJSON:    true,
			wantHistory: true,
		},
		{
			name:       "history rejects unsupported method",
			method:     http.MethodPost,
			path:       "/history",
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodGet,
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

			previousPersistence := newPersistence
			newPersistence = func(context.Context) (persistence, error) {
				return fakePersistence{
					latest: models.LatestResponse{
						Sites:     []models.CheckResult{{URL: "https://site-a.example", IsUp: true}},
						UpdatedAt: "2026-04-30T12:00:00Z",
					},
					history: models.HistoryResponse{
						History: map[string][]models.CheckResult{
							"https://site-a.example": {{URL: "https://site-a.example", IsUp: true}},
						},
					},
				}, nil
			}
			t.Cleanup(func() {
				newPersistence = previousPersistence
			})

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

			if tt.wantLatest {
				var response models.LatestResponse
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if len(response.Sites) != 1 || response.Sites[0].URL != "https://site-a.example" {
					t.Fatalf("expected latest site response, got %#v", response)
				}
				if response.UpdatedAt != "2026-04-30T12:00:00Z" {
					t.Fatalf("expected updated_at timestamp, got %q", response.UpdatedAt)
				}
				return
			}

			if tt.wantHistory {
				var response models.HistoryResponse
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if len(response.History["https://site-a.example"]) != 1 {
					t.Fatalf("expected one history entry, got %#v", response)
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

func TestRouterStorageError(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		method     string
		envTargets string
		wantStatus int
	}{
		{
			name:       "check returns error when persistence fails",
			path:       "/check",
			method:     http.MethodPost,
			envTargets: "https://site-a.example",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "latest returns error when persistence fails",
			path:       "/latest",
			method:     http.MethodGet,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "history returns error when persistence fails",
			path:       "/history",
			method:     http.MethodGet,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MONITOR_TARGETS", tt.envTargets)
			previousPersistence := newPersistence
			newPersistence = func(context.Context) (persistence, error) {
				return nil, errors.New("storage failed")
			}
			t.Cleanup(func() {
				newPersistence = previousPersistence
			})

			previousChecks := performChecks
			performChecks = func(_ context.Context, urls []string) []models.CheckResult {
				return []models.CheckResult{{URL: urls[0], IsUp: true}}
			}
			t.Cleanup(func() {
				performChecks = previousChecks
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			NewRouter().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
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

type fakePersistence struct {
	latest  models.LatestResponse
	history models.HistoryResponse
}

func (f fakePersistence) Save(context.Context, []models.CheckResult) error {
	return nil
}

func (f fakePersistence) Latest(context.Context) (models.LatestResponse, error) {
	return f.latest, nil
}

func (f fakePersistence) History(context.Context) (models.HistoryResponse, error) {
	return f.history, nil
}
