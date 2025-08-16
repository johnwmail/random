package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type helloResp struct {
	Message string `json:"message"`
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Minimal REST v1 proxy handler
	path := req.Path
	// Serve at /hello and /
	if path == "/hello" || path == "/" {
		b, _ := json.Marshal(helloResp{Message: "hello"})
		return events.APIGatewayProxyResponse{
			StatusCode:        200,
			Headers:           map[string]string{"Content-Type": "application/json"},
			MultiValueHeaders: map[string][]string{},
			Body:              string(b),
			IsBase64Encoded:   false,
		}, nil
	}
	// 404 for others
	return events.APIGatewayProxyResponse{
		StatusCode:        404,
		Headers:           map[string]string{"Content-Type": "application/json"},
		MultiValueHeaders: map[string][]string{},
		Body:              `{"message":"not found"}`,
		IsBase64Encoded:   false,
	}, nil
}

func main() { lambda.Start(handler) }
