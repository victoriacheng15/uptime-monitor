package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"uptime-monitor/internal/api"
)

func main() {
	router := api.NewRouter()
	adapter := httpadapter.NewV2(router)

	lambda.Start(func(ctx context.Context, payload json.RawMessage) (events.APIGatewayV2HTTPResponse, error) {
		var request events.APIGatewayV2HTTPRequest
		if err := json.Unmarshal(payload, &request); err != nil {
			return events.APIGatewayV2HTTPResponse{}, err
		}
		if request.RawPath != "" || request.RequestContext.HTTP.Method != "" {
			return adapter.ProxyWithContext(ctx, request)
		}

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/check", nil).WithContext(ctx)
		router.ServeHTTP(rec, req)

		return events.APIGatewayV2HTTPResponse{
			StatusCode: rec.Code,
			Headers:    responseHeaders(rec.Header()),
			Body:       rec.Body.String(),
		}, nil
	})
}

func responseHeaders(headers http.Header) map[string]string {
	response := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) > 0 {
			response[key] = values[0]
		}
	}

	return response
}
