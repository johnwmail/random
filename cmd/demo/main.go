package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Serve simple HTML for browser-friendly responses at / and /hello
	path := req.Path
	if path == "/hello" || path == "/" {
		// Check query param and Accept header for JSON preference
		qs := req.QueryStringParameters["format"]
		accept := ""
		for k, v := range req.Headers {
			if strings.EqualFold(k, "Accept") {
				accept = v
				break
			}
		}

		wantJSON := false
		if strings.EqualFold(qs, "json") {
			wantJSON = true
		} else if accept != "" {
			if strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html") {
				wantJSON = true
			}
		}

		if wantJSON {
			b, _ := json.Marshal(map[string]string{"message": "hello"})
			return events.APIGatewayProxyResponse{
				StatusCode:        200,
				Headers:           map[string]string{"Content-Type": "application/json"},
				MultiValueHeaders: map[string][]string{},
				Body:              string(b),
				IsBase64Encoded:   false,
			}, nil
		}

		html := `<!doctype html>
<html>
  <head>
	<meta charset="utf-8" />
	<title>Hello</title>
	<style>body{font-family:system-ui,Segoe UI,Roboto,Helvetica,Arial;margin:40px}</style>
  </head>
  <body>
	<h1>Hello</h1>
	<p>This is the demo Lambda served through API Gateway (REST).</p>
  </body>
</html>`

		return events.APIGatewayProxyResponse{
			StatusCode:        200,
			Headers:           map[string]string{"Content-Type": "text/html; charset=utf-8"},
			MultiValueHeaders: map[string][]string{},
			Body:              html,
			IsBase64Encoded:   false,
		}, nil
	}

	// 404 for others (JSON)
	return events.APIGatewayProxyResponse{
		StatusCode:        404,
		Headers:           map[string]string{"Content-Type": "application/json"},
		MultiValueHeaders: map[string][]string{},
		Body:              `{"message":"not found"}`,
		IsBase64Encoded:   false,
	}, nil
}

func main() { lambda.Start(handler) }
