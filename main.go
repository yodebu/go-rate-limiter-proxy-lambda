package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/ratelimit"
)

// InputRequest is the struct that will be used to unmarshal the http request.
type InputRequest struct {
	Method  string            `json:"method"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// OutputResponse is the
type OutputResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

var limiter ratelimit.Limiter

var once sync.Once

// initLimiter
func initLimiter(rate int) {
	once.Do(func() {
		limiter = ratelimit.New(rate)
	})
}

func Handler(ctx context.Context, input InputRequest) (OutputResponse, error) {
	initLimiter(1)
	limiter.Take()
	reqBody := bytes.NewBufferString(input.Body)
	req, err := http.NewRequest(input.Method, input.Url, reqBody)
	if err != nil {
		return OutputResponse{}, err
	}

	for key, value := range input.Headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return OutputResponse{}, err
	}

	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	respHeaders := make(map[string]string)
	for key, value := range resp.Header {
		respHeaders[key] = value[0]
	}
	return OutputResponse{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       string(respBody),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
