package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/marcossnikel/payroll-project/internal/httpapi"
)

var appHandler = httpapi.NewApp(httpapi.Config{StepDelay: 100 * time.Millisecond}).Handler()

func handle(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := event.RawPath
	if path == "" {
		path = event.RequestContext.HTTP.Path
	}
	target := path
	if event.RawQueryString != "" {
		target += "?" + event.RawQueryString
	}

	request, err := http.NewRequestWithContext(ctx, event.RequestContext.HTTP.Method, target, strings.NewReader(event.Body))
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusBadRequest}, nil
	}
	for key, value := range event.Headers {
		request.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	appHandler.ServeHTTP(recorder, request)
	response := recorder.Result()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}

	headers := make(map[string]string, len(response.Header))
	for key, values := range response.Header {
		headers[key] = strings.Join(values, ",")
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: response.StatusCode,
		Headers:    headers,
		Body:       string(bytes.Clone(body)),
	}, nil
}

func main() {
	lambda.Start(handle)
}
