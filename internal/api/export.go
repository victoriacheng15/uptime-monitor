package api

import (
	"context"
	"uptime-monitor/internal/models"
)

// ExportedPersistence mirrors the internal persistence interface for testing purposes.
type ExportedPersistence interface {
	Save(ctx context.Context, results []models.CheckResult) error
	Latest(ctx context.Context) (models.LatestResponse, error)
	History(ctx context.Context) (models.HistoryResponse, error)
}

// SetNewPersistence allows external test suites to mock S3 persistence.
func SetNewPersistence(fn func(ctx context.Context) (ExportedPersistence, error)) {
	newPersistence = func(ctx context.Context) (persistence, error) {
		return fn(ctx)
	}
}

// SetPerformChecks allows external test suites to mock the target HTTP checkers.
func SetPerformChecks(fn func(ctx context.Context, urls []string) []models.CheckResult) {
	performChecks = fn
}
