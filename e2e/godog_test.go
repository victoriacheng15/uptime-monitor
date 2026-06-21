package e2e

import (
	"testing"

	"github.com/cucumber/godog"
	"uptime-monitor/e2e/step_definitions"
	"uptime-monitor/e2e/support"
)

func TestFeatures(t *testing.T) {
	state := &support.TestState{}

	suite := godog.TestSuite{
		Name: "Uptime Monitor E2E Suite",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			support.InitializeScenario(ctx, state)
			step_definitions.RegisterCommonSteps(ctx, state)
			step_definitions.RegisterEventBridgeSteps(ctx, state)
			step_definitions.RegisterHTTPClientSteps(ctx, state)
			step_definitions.RegisterS3StorageSteps(ctx, state)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
