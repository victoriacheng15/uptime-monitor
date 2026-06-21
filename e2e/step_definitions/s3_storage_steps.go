package step_definitions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cucumber/godog"

	"uptime-monitor/e2e/support"
	"uptime-monitor/internal/models"
	"uptime-monitor/internal/storage"
)

func RegisterS3StorageSteps(ctx *godog.ScenarioContext, state *support.TestState) {
	ctx.Step(`^a set of check results:$`, func(tbl *godog.Table) error {
		state.CheckResults = nil
		for _, row := range tbl.Rows[1:] {
			url := row.Cells[0].Value
			var status int
			_, err := fmt.Sscanf(row.Cells[1].Value, "%d", &status)
			if err != nil {
				return err
			}
			isUp := row.Cells[2].Value == "true"
			state.CheckResults = append(state.CheckResults, models.CheckResult{
				URL:        url,
				StatusCode: status,
				IsUp:       isUp,
				Timestamp:  time.Now().Format(time.RFC3339),
			})
		}
		return nil
	})

	ctx.Step(`^the S3 storage is mocked and empty$`, func() error {
		state.MockS3Client = &support.FakeS3{
			Objects: make(map[string][]byte),
		}
		return nil
	})

	ctx.Step(`^the check results are persisted to storage$`, func() error {
		store := storage.New("test-bucket", state.MockS3Client)
		return store.Save(context.Background(), state.CheckResults)
	})

	ctx.Step(`^the mock S3 storage must contain "latest.json" and "history.json" matching the check results$`, func() error {
		latestBody, ok := state.MockS3Client.Objects["latest.json"]
		if !ok {
			return fmt.Errorf("latest.json is missing in mock S3 bucket")
		}
		var latest models.LatestResponse
		if err := json.Unmarshal(latestBody, &latest); err != nil {
			return fmt.Errorf("failed to unmarshal latest.json: %w", err)
		}
		if len(latest.Sites) != len(state.CheckResults) {
			return fmt.Errorf("expected %d sites in latest.json, got %d", len(state.CheckResults), len(latest.Sites))
		}

		historyBody, ok := state.MockS3Client.Objects["history.json"]
		if !ok {
			return fmt.Errorf("history.json is missing in mock S3 bucket")
		}
		var history models.HistoryResponse
		if err := json.Unmarshal(historyBody, &history); err != nil {
			return fmt.Errorf("failed to unmarshal history.json: %w", err)
		}

		for _, check := range state.CheckResults {
			entries, found := history.History[check.URL]
			if !found || len(entries) == 0 {
				return fmt.Errorf("history for %s is missing in history.json", check.URL)
			}
			if entries[len(entries)-1].StatusCode != check.StatusCode {
				return fmt.Errorf("expected last history status %d for %s, got %d", check.StatusCode, check.URL, entries[len(entries)-1].StatusCode)
			}
		}
		return nil
	})

	ctx.Step(`^the mock S3 storage contains an existing history for "([^"]*)" with (\d+) entries$`, func(url string, count int) error {
		state.MockS3Client = &support.FakeS3{
			Objects: make(map[string][]byte),
		}

		existing := models.HistoryResponse{
			History: map[string][]models.CheckResult{
				url: make([]models.CheckResult, count),
			},
		}
		for i := 0; i < count; i++ {
			existing.History[url][i] = models.CheckResult{
				URL:        url,
				StatusCode: 200,
				IsUp:       true,
				Timestamp:  time.Now().Add(-time.Duration(count-i) * time.Hour).Format(time.RFC3339),
			}
		}

		body, err := json.Marshal(existing)
		if err != nil {
			return err
		}
		state.MockS3Client.Objects["history.json"] = body
		return nil
	})

	ctx.Step(`^the history entries for "([^"]*)" must be trimmed to keep only (\d+) entries$`, func(url string, maxCount int) error {
		historyBody, ok := state.MockS3Client.Objects["history.json"]
		if !ok {
			return fmt.Errorf("history.json missing in S3")
		}
		var history models.HistoryResponse
		if err := json.Unmarshal(historyBody, &history); err != nil {
			return err
		}

		entries := history.History[url]
		if len(entries) != maxCount {
			return fmt.Errorf("expected history entries to be trimmed to %d, but got %d", maxCount, len(entries))
		}
		return nil
	})
}
