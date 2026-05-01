package api

import (
	"context"
	"net/http"
	"os"
	"strings"

	"uptime-monitor/internal/models"
	"uptime-monitor/internal/monitor"
	"uptime-monitor/internal/storage"
)

var performChecks = func(ctx context.Context, urls []string) []models.CheckResult {
	checker := monitor.New(nil)
	return checker.PerformChecks(ctx, urls)
}

type persistence interface {
	Save(ctx context.Context, results []models.CheckResult) error
	Latest(ctx context.Context) (models.LatestResponse, error)
	History(ctx context.Context) (models.HistoryResponse, error)
}

var newPersistence = func(ctx context.Context) (persistence, error) {
	bucket := os.Getenv("HISTORY_BUCKET")
	if bucket == "" {
		return nil, errMissingHistoryBucket
	}

	return storage.NewS3(ctx, bucket)
}

func checkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		targets := monitorTargets()
		if len(targets) == 0 {
			http.Error(w, "MONITOR_TARGETS is not configured", http.StatusInternalServerError)
			return
		}

		results := performChecks(r.Context(), targets)
		store, err := newPersistence(r.Context())
		if err != nil {
			http.Error(w, "storage is not available", http.StatusInternalServerError)
			return
		}
		if err := store.Save(r.Context(), results); err != nil {
			http.Error(w, "failed to persist check results", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, results)
	}
}

func monitorTargets() []string {
	raw := os.Getenv("MONITOR_TARGETS")
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	targets := make([]string, 0, len(parts))
	for _, part := range parts {
		target := strings.TrimSpace(part)
		if target != "" {
			targets = append(targets, target)
		}
	}

	return targets
}
