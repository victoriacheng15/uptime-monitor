package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"uptime-monitor/internal/models"
)

func TestStoreSave(t *testing.T) {
	existing := models.HistoryResponse{
		History: map[string][]models.CheckResult{
			"https://site-a.example": {
				{URL: "https://site-a.example", Timestamp: "2026-04-30T00:00:00Z"},
				{URL: "https://site-a.example", Timestamp: "2026-04-30T01:00:00Z"},
				{URL: "https://site-a.example", Timestamp: "2026-04-30T02:00:00Z"},
				{URL: "https://site-a.example", Timestamp: "2026-04-30T03:00:00Z"},
				{URL: "https://site-a.example", Timestamp: "2026-04-30T04:00:00Z"},
			},
		},
	}
	historyBody, err := json.Marshal(existing)
	if err != nil {
		t.Fatalf("marshal history: %v", err)
	}

	client := &fakeS3{
		objects: map[string][]byte{
			HistoryKey: historyBody,
		},
	}
	store := New("bucket", client)
	store.now = func() time.Time {
		return time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC)
	}

	results := []models.CheckResult{
		{URL: "https://site-a.example", StatusCode: 200, IsUp: true, Timestamp: "2026-04-30T05:00:00Z"},
		{URL: "https://site-b.example", StatusCode: 500, IsUp: false, Timestamp: "2026-04-30T05:00:00Z"},
	}

	if err := store.Save(context.Background(), results); err != nil {
		t.Fatalf("save results: %v", err)
	}

	var latest models.LatestResponse
	if err := json.Unmarshal(client.objects[LatestKey], &latest); err != nil {
		t.Fatalf("unmarshal latest: %v", err)
	}
	if latest.UpdatedAt != "2026-04-30T12:00:00Z" {
		t.Fatalf("expected latest timestamp, got %q", latest.UpdatedAt)
	}
	if len(latest.Sites) != 2 {
		t.Fatalf("expected 2 latest sites, got %d", len(latest.Sites))
	}

	var history models.HistoryResponse
	if err := json.Unmarshal(client.objects[HistoryKey], &history); err != nil {
		t.Fatalf("unmarshal history: %v", err)
	}
	if got := len(history.History["https://site-a.example"]); got != historyLimit {
		t.Fatalf("expected site-a history to stay at %d entries, got %d", historyLimit, got)
	}
	if got := history.History["https://site-a.example"][0].Timestamp; got != "2026-04-30T01:00:00Z" {
		t.Fatalf("expected oldest retained timestamp to be trimmed, got %q", got)
	}
	if got := len(history.History["https://site-b.example"]); got != 1 {
		t.Fatalf("expected site-b history to have one entry, got %d", got)
	}
}

func TestStoreLatest(t *testing.T) {
	latestBody, err := json.Marshal(models.LatestResponse{
		Sites:     []models.CheckResult{{URL: "https://site-a.example", IsUp: true}},
		UpdatedAt: "2026-04-30T12:00:00Z",
	})
	if err != nil {
		t.Fatalf("marshal latest: %v", err)
	}

	tests := []struct {
		name          string
		objects       map[string][]byte
		wantSites     int
		wantUpdatedAt string
		wantErr       bool
	}{
		{
			name:      "returns empty response when latest object is missing",
			objects:   map[string][]byte{},
			wantSites: 0,
		},
		{
			name: "returns stored latest response",
			objects: map[string][]byte{
				LatestKey: latestBody,
			},
			wantSites:     1,
			wantUpdatedAt: "2026-04-30T12:00:00Z",
		},
		{
			name: "returns error for invalid latest JSON",
			objects: map[string][]byte{
				LatestKey: []byte("{"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := New("bucket", &fakeS3{objects: tt.objects})

			latest, err := store.Latest(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("latest: %v", err)
			}
			if len(latest.Sites) != tt.wantSites {
				t.Fatalf("expected %d latest sites, got %d", tt.wantSites, len(latest.Sites))
			}
			if latest.UpdatedAt != tt.wantUpdatedAt {
				t.Fatalf("expected updated_at %q, got %q", tt.wantUpdatedAt, latest.UpdatedAt)
			}
		})
	}
}

func TestStoreHistory(t *testing.T) {
	historyBody, err := json.Marshal(models.HistoryResponse{
		History: map[string][]models.CheckResult{
			"https://site-a.example": {{URL: "https://site-a.example", IsUp: true}},
		},
	})
	if err != nil {
		t.Fatalf("marshal history: %v", err)
	}

	nilMapBody, err := json.Marshal(models.HistoryResponse{})
	if err != nil {
		t.Fatalf("marshal nil history: %v", err)
	}

	tests := []struct {
		name        string
		objects     map[string][]byte
		wantEntries int
		wantErr     bool
	}{
		{
			name:        "returns empty response when history object is missing",
			objects:     map[string][]byte{},
			wantEntries: 0,
		},
		{
			name: "returns stored history response",
			objects: map[string][]byte{
				HistoryKey: historyBody,
			},
			wantEntries: 1,
		},
		{
			name: "initializes nil history map",
			objects: map[string][]byte{
				HistoryKey: nilMapBody,
			},
			wantEntries: 0,
		},
		{
			name: "returns error for invalid history JSON",
			objects: map[string][]byte{
				HistoryKey: []byte("{"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := New("bucket", &fakeS3{objects: tt.objects})

			history, err := store.History(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("history: %v", err)
			}
			if history.History == nil {
				t.Fatal("expected history map to be initialized")
			}
			if got := len(history.History["https://site-a.example"]); got != tt.wantEntries {
				t.Fatalf("expected %d history entries, got %d", tt.wantEntries, got)
			}
		})
	}
}

func TestStorePutJSONMarshalError(t *testing.T) {
	store := New("bucket", &fakeS3{objects: map[string][]byte{}})

	if err := store.putJSON(context.Background(), LatestKey, make(chan struct{})); err == nil {
		t.Fatal("expected marshal error")
	}
}

type fakeS3 struct {
	objects map[string][]byte
}

func (f *fakeS3) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	body, ok := f.objects[aws.ToString(input.Key)]
	if !ok {
		return nil, fakeNoSuchKey{}
	}

	return &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(body)),
	}, nil
}

func (f *fakeS3) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}

	f.objects[aws.ToString(input.Key)] = body
	return &s3.PutObjectOutput{}, nil
}

type fakeNoSuchKey struct{}

func (fakeNoSuchKey) Error() string {
	return "NoSuchKey"
}

func (fakeNoSuchKey) ErrorCode() string {
	return "NoSuchKey"
}

func (fakeNoSuchKey) ErrorMessage() string {
	return "missing key"
}

func (fakeNoSuchKey) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}
