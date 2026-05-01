package models

import (
	"testing"
	"time"
)

func TestNewHealthResponse(t *testing.T) {
	response := NewHealthResponse("healthy")

	if response.Status != "healthy" {
		t.Fatalf("expected status healthy, got %q", response.Status)
	}

	if _, err := time.Parse(time.RFC3339, response.Timestamp); err != nil {
		t.Fatalf("expected RFC3339 timestamp, got %q: %v", response.Timestamp, err)
	}
}

func TestCheckResultIsSuccessful(t *testing.T) {
	tests := []struct {
		name   string
		result CheckResult
		want   bool
	}{
		{
			name: "returns true for up 200 response",
			result: CheckResult{
				StatusCode: 200,
				IsUp:       true,
			},
			want: true,
		},
		{
			name: "returns true for up redirect response",
			result: CheckResult{
				StatusCode: 301,
				IsUp:       true,
			},
			want: true,
		},
		{
			name: "returns false when status is server error",
			result: CheckResult{
				StatusCode: 500,
				IsUp:       false,
			},
		},
		{
			name: "returns false when request has error",
			result: CheckResult{
				StatusCode: 200,
				IsUp:       true,
				Error:      "timeout",
			},
		},
		{
			name: "returns false when status is below success range",
			result: CheckResult{
				StatusCode: 199,
				IsUp:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsSuccessful(); got != tt.want {
				t.Fatalf("expected %t, got %t", tt.want, got)
			}
		})
	}
}
