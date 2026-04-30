package api

import (
	"context"
	"net/http"
	"os"
	"strings"

	"uptime-monitor/internal/models"
	"uptime-monitor/internal/monitor"
)

var performChecks = func(ctx context.Context, urls []string) []models.CheckResult {
	checker := monitor.New(nil)
	return checker.PerformChecks(ctx, urls)
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
