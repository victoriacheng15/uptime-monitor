package step_definitions

import (
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"

	"uptime-monitor/e2e/support"
)

func RegisterHTTPClientSteps(ctx *godog.ScenarioContext, state *support.TestState) {
	ctx.Step(`^the environment variable "([^"]*)" is configured with target HTTP endpoints$`, func(name string) error {
		var targets []string
		for path := range state.ResponseStatuses {
			targets = append(targets, state.MockServer.URL+path)
		}
		if len(targets) == 0 {
			targets = []string{
				state.MockServer.URL + "/target-a",
				state.MockServer.URL + "/target-b",
			}
		}
		os.Setenv(name, strings.Join(targets, ","))
		os.Setenv("HISTORY_BUCKET", "dummy-bucket")
		return nil
	})

	ctx.Step(`^the target endpoints are mocked to return:$`, func(tbl *godog.Table) error {
		for _, row := range tbl.Rows[1:] {
			path := row.Cells[0].Value
			var status int
			_, err := fmt.Sscanf(row.Cells[1].Value, "%d", &status)
			if err != nil {
				return err
			}
			state.ResponseStatuses[path] = status
		}
		return nil
	})

	ctx.Step(`^the target HTTP endpoints "([^"]*)" must be requested by the HTTP client$`, func(paths string) error {
		for _, part := range strings.Split(paths, ",") {
			path := strings.TrimSpace(part)
			if !state.RequestedPaths[path] {
				return fmt.Errorf("expected request to %s, but got none", path)
			}
		}
		return nil
	})
}
