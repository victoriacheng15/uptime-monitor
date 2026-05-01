package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"uptime-monitor/internal/models"
)

const (
	LatestKey       = "latest.json"
	HistoryKey      = "history.json"
	historyLimit    = 5
	jsonContentType = "application/json"
)

type S3API interface {
	GetObject(ctx context.Context, input *s3.GetObjectInput, options ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, input *s3.PutObjectInput, options ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type Store struct {
	bucket string
	client S3API
	now    func() time.Time
}

func NewS3(ctx context.Context, bucket string) (Store, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return Store{}, err
	}

	return New(bucket, s3.NewFromConfig(cfg)), nil
}

func New(bucket string, client S3API) Store {
	return Store{
		bucket: bucket,
		client: client,
		now:    time.Now,
	}
}

func (s Store) Save(ctx context.Context, results []models.CheckResult) error {
	history, err := s.History(ctx)
	if err != nil {
		return err
	}

	for _, result := range results {
		entries := append(history.History[result.URL], result)
		if len(entries) > historyLimit {
			entries = entries[len(entries)-historyLimit:]
		}
		history.History[result.URL] = entries
	}

	latest := models.LatestResponse{
		Sites:     results,
		UpdatedAt: s.now().UTC().Format(time.RFC3339),
	}

	if err := s.putJSON(ctx, LatestKey, latest); err != nil {
		return err
	}

	return s.putJSON(ctx, HistoryKey, history)
}

func (s Store) Latest(ctx context.Context) (models.LatestResponse, error) {
	var latest models.LatestResponse
	if err := s.getJSON(ctx, LatestKey, &latest); err != nil {
		if isNoSuchKey(err) {
			return models.LatestResponse{Sites: []models.CheckResult{}}, nil
		}
		return models.LatestResponse{}, err
	}

	return latest, nil
}

func (s Store) History(ctx context.Context) (models.HistoryResponse, error) {
	history := models.HistoryResponse{History: map[string][]models.CheckResult{}}
	if err := s.getJSON(ctx, HistoryKey, &history); err != nil {
		if isNoSuchKey(err) {
			return history, nil
		}
		return models.HistoryResponse{}, err
	}
	if history.History == nil {
		history.History = map[string][]models.CheckResult{}
	}

	return history, nil
}

func (s Store) putJSON(ctx context.Context, key string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(jsonContentType),
	})
	return err
}

func (s Store) getJSON(ctx context.Context, key string, payload any) error {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, payload)
}

func isNoSuchKey(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchKey"
}
