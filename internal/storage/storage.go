package storage

import "uptime-monitor/internal/models"

func Hello() string {
	return "hello from storage"
}

func Latest(results []models.CheckResult) models.LatestResponse {
	return models.LatestResponse{
		Sites: results,
	}
}

func History(results []models.CheckResult) models.HistoryResponse {
	history := make(map[string][]models.CheckResult, len(results))
	for _, result := range results {
		history[result.URL] = append(history[result.URL], result)
	}

	return models.HistoryResponse{History: history}
}
