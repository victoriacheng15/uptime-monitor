package step_definitions

import (
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/cucumber/godog"

	"uptime-monitor/e2e/support"
	"uptime-monitor/internal/api"
)

func RegisterEventBridgeSteps(ctx *godog.ScenarioContext, state *support.TestState) {
	ctx.Step(`^an EventBridge scheduled trigger event is received$`, func() error {
		// Set dummy values if not already unset by a Given step
		if os.Getenv("MONITOR_TARGETS") == "" && !state.RequestedPaths["unset_targets"] {
			os.Setenv("MONITOR_TARGETS", state.MockServer.URL+"/target-a")
		}
		os.Setenv("HISTORY_BUCKET", "dummy-bucket")

		router := api.NewRouter()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/check", nil)
		router.ServeHTTP(rec, req)

		state.LastStatusCode = rec.Code
		return nil
	})

	ctx.Step(`^the environment variable "([^"]*)" is unset$`, func(name string) error {
		os.Unsetenv(name)
		state.RequestedPaths["unset_targets"] = true
		return nil
	})
}
