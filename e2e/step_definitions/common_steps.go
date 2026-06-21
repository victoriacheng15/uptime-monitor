package step_definitions

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/cucumber/godog"

	"uptime-monitor/e2e/support"
	"uptime-monitor/internal/api"
)

func RegisterCommonSteps(ctx *godog.ScenarioContext, state *support.TestState) {
	ctx.Step(`^the status check handler is executed$`, func() error {
		router := api.NewRouter()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/check", nil)
		router.ServeHTTP(rec, req)
		state.LastStatusCode = rec.Code
		if rec.Code != http.StatusOK {
			return fmt.Errorf("check handler failed with status %d: %s", rec.Code, rec.Body.String())
		}
		return nil
	})

	ctx.Step(`^the status check handler runs and returns HTTP status code (\d+)$`, func(expectedCode int) error {
		if state.LastStatusCode != expectedCode {
			return fmt.Errorf("expected HTTP status %d, got %d", expectedCode, state.LastStatusCode)
		}
		return nil
	})
}
