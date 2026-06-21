package support

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/cucumber/godog"

	"uptime-monitor/internal/api"
	"uptime-monitor/internal/models"
)

type TestState struct {
	MockServer       *httptest.Server
	MockPersistence  *MockPersistence
	MockS3Client     *FakeS3
	RequestedPaths   map[string]bool
	ResponseStatuses map[string]int
	CheckResults     []models.CheckResult
	LastStatusCode   int
}

type MockPersistence struct {
	SavedResults []models.CheckResult
}

func (m *MockPersistence) Save(ctx context.Context, results []models.CheckResult) error {
	m.SavedResults = results
	return nil
}

func (m *MockPersistence) Latest(ctx context.Context) (models.LatestResponse, error) {
	return models.LatestResponse{}, nil
}

func (m *MockPersistence) History(ctx context.Context) (models.HistoryResponse, error) {
	return models.HistoryResponse{}, nil
}

type FakeS3 struct {
	Objects map[string][]byte
	GetErr  error
	PutErr  error
}

func (f *FakeS3) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if f.GetErr != nil {
		return nil, f.GetErr
	}
	body, ok := f.Objects[aws.ToString(input.Key)]
	if !ok {
		return nil, FakeNoSuchKey{}
	}

	return &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(body)),
	}, nil
}

func (f *FakeS3) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if f.PutErr != nil {
		return nil, f.PutErr
	}
	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}

	f.Objects[aws.ToString(input.Key)] = body
	return &s3.PutObjectOutput{}, nil
}

type FakeNoSuchKey struct{}

func (FakeNoSuchKey) Error() string { return "NoSuchKey" }
func (FakeNoSuchKey) ErrorCode() string { return "NoSuchKey" }
func (FakeNoSuchKey) ErrorMessage() string { return "missing key" }
func (FakeNoSuchKey) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

func InitializeScenario(ctx *godog.ScenarioContext, state *TestState) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		state.RequestedPaths = make(map[string]bool)
		state.ResponseStatuses = make(map[string]int)

		state.MockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			state.RequestedPaths[r.URL.Path] = true
			status, ok := state.ResponseStatuses[r.URL.Path]
			if !ok {
				status = 200
			}
			w.WriteHeader(status)
		}))

		state.MockPersistence = &MockPersistence{}
		api.SetNewPersistence(func(ctx context.Context) (api.ExportedPersistence, error) {
			return state.MockPersistence, nil
		})

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if state.MockServer != nil {
			state.MockServer.Close()
		}
		return ctx, nil
	})
}
