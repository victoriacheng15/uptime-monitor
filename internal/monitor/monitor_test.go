package monitor

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripClient struct {
	response *http.Response
	err      error
	request  *http.Request
}

func (c *roundTripClient) Do(req *http.Request) (*http.Response, error) {
	c.request = req
	return c.response, c.err
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		response       *http.Response
		err            error
		wantStatusCode int
		wantIsUp       bool
		wantError      string
	}{
		{
			name: "marks 200 response up",
			url:  "https://example.com",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
			},
			wantStatusCode: http.StatusOK,
			wantIsUp:       true,
		},
		{
			name: "marks redirect up",
			url:  "https://example.com",
			response: &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Body:       io.NopCloser(strings.NewReader("redirect")),
			},
			wantStatusCode: http.StatusMovedPermanently,
			wantIsUp:       true,
		},
		{
			name: "marks 500 response down",
			url:  "https://example.com",
			response: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("error")),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:      "records request error",
			url:       "https://example.com",
			err:       errors.New("dial failed"),
			wantError: "dial failed",
		},
		{
			name:      "records invalid url error",
			url:       "://bad-url",
			wantError: "missing protocol scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &roundTripClient{
				response: tt.response,
				err:      tt.err,
			}
			monitor := New(client)
			now := time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC)
			monitor.now = func() time.Time {
				now = now.Add(25 * time.Millisecond)
				return now
			}

			result := monitor.Check(context.Background(), tt.url)

			if result.URL != tt.url {
				t.Fatalf("expected url %q, got %q", tt.url, result.URL)
			}

			if result.StatusCode != tt.wantStatusCode {
				t.Fatalf("expected status code %d, got %d", tt.wantStatusCode, result.StatusCode)
			}

			if result.IsUp != tt.wantIsUp {
				t.Fatalf("expected is_up %t, got %t", tt.wantIsUp, result.IsUp)
			}

			if tt.wantError != "" && !strings.Contains(result.Error, tt.wantError) {
				t.Fatalf("expected error containing %q, got %q", tt.wantError, result.Error)
			}

			if result.Timestamp == "" {
				t.Fatal("expected timestamp")
			}
		})
	}
}

func TestPerformChecks(t *testing.T) {
	client := &roundTripClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
		},
	}
	monitor := New(client)

	results := monitor.PerformChecks(context.Background(), []string{
		"https://site-a.example",
		"https://site-b.example",
	})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].URL != "https://site-a.example" {
		t.Fatalf("expected first url, got %q", results[0].URL)
	}

	if results[1].URL != "https://site-b.example" {
		t.Fatalf("expected second url, got %q", results[1].URL)
	}
}

func TestHello(t *testing.T) {
	got := Hello()
	want := "hello from monitor"
	if got != want {
		t.Errorf("Hello() = %q; want %q", got, want)
	}
}

func TestNewDefaultClient(t *testing.T) {
	m := New(nil)
	if m.client == nil {
		t.Error("expected default HTTP client to be assigned when nil is passed")
	}
}

func TestPackagePerformChecks(t *testing.T) {
	results := PerformChecks([]string{})
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
