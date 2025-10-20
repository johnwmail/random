package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomPrintable(t *testing.T) {
	// Test with a specific length
	length := 15
	result := GenerateRandomPrintable(length)
	assert.Equal(t, length, len(result), "Generated printable string should have the correct length")

	// Test with zero length
	length = 0
	result = GenerateRandomPrintable(length)
	assert.Equal(t, "", result, "Generated printable string should be empty for zero length")
}

func TestGenerateRandomAlphanumeric(t *testing.T) {
	// Test with a specific length
	length := 20
	result := GenerateRandomAlphanumeric(length)
	assert.Equal(t, length, len(result), "Generated alphanumeric string should have the correct length")

	// Test with zero length
	length = 0
	result = GenerateRandomAlphanumeric(length)
	assert.Equal(t, "", result, "Generated alphanumeric string should be empty for zero length")
}

func TestGenerateStringsJSON(t *testing.T) {
	// Set up the router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/json", generateStrings)

	// Create a request to the /json endpoint
	req, _ := http.NewRequest(http.MethodGet, "/json?p=15&a=20", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code, "HTTP status code should be 200")

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Should be able to unmarshal JSON response")

	assert.Equal(t, 15, response.Printable.Length, "Printable length should be correct")
	assert.Equal(t, 20, response.AlphaNumeric.Length, "Alphanumeric length should be correct")
}

func TestGenerateStringsHTML(t *testing.T) {
	// Set up the router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/", generateStrings)

	// Create a request to the / endpoint
	req, _ := http.NewRequest(http.MethodGet, "/?p=18&a=22", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code, "HTTP status code should be 200")

	body := w.Body.String()
	assert.True(t, strings.Contains(body, "Random String Generator"), "Response should include the page headline text")
	assert.True(t, strings.Contains(body, "<svg"), "Response should include inline SVG icons")
	assert.True(t, strings.Contains(body, "Printable String"), "Response should contain Printable string")
	assert.True(t, strings.Contains(body, "Alphanumeric String"), "Response should contain Alphanumeric string")
}

// TestAPIGatewayV1Event verifies the app handles API Gateway v1 events correctly
func TestAPIGatewayV1Event(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/json", generateStrings)
	ginLambda := ginadapter.New(r)

	// Create a mock API Gateway v1 event
	v1Event := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/json",
		QueryStringParameters: map[string]string{
			"p": "15",
			"a": "20",
		},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	handler := &universalHandler{v1: ginLambda}
	payload, err := json.Marshal(v1Event)
	assert.NoError(t, err, "Should marshal v1 event successfully")

	result, err := handler.tryAPIGatewayV1(context.Background(), payload)
	assert.NoError(t, err, "Should handle v1 event without error")
	assert.NotNil(t, result, "Should return a result")

	var resp events.APIGatewayProxyResponse
	err = json.Unmarshal(result, &resp)
	assert.NoError(t, err, "Should unmarshal response successfully")
	assert.Equal(t, 200, resp.StatusCode, "Should return 200 status code")
}

// TestAPIGatewayV2Event verifies the app handles API Gateway v2 events correctly
func TestAPIGatewayV2Event(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/json", generateStrings)
	ginLambdaV2 := ginadapter.NewV2(r)

	// Create a mock API Gateway v2 event
	v2Event := events.APIGatewayV2HTTPRequest{
		Version:        "2.0",
		RawPath:        "/json",
		RawQueryString: "p=15&a=20",
		Headers: map[string]string{
			"content-type": "application/json",
		},
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "GET",
				Path:   "/json",
			},
		},
	}

	handler := &universalHandler{v2: ginLambdaV2}
	payload, err := json.Marshal(v2Event)
	assert.NoError(t, err, "Should marshal v2 event successfully")

	result, err := handler.tryAPIGatewayV2(context.Background(), payload)
	assert.NoError(t, err, "Should handle v2 event without error")
	assert.NotNil(t, result, "Should return a result")

	var resp events.APIGatewayV2HTTPResponse
	err = json.Unmarshal(result, &resp)
	assert.NoError(t, err, "Should unmarshal response successfully")
	assert.Equal(t, 200, resp.StatusCode, "Should return 200 status code")
}

// --- User-Agent behavior tests (previously in main_useragent_test.go) ---

// helper to create a test router with the relevant routes
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.Static("/static", "./static")
	r.GET("/json", generateStrings)
	r.GET("/", generateStrings)
	return r
}

func TestUserAgent_CLI_ReturnsJSON(t *testing.T) {
	r := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "curl/7.80.0")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("expected application/json content-type for CLI UA, got %s", ct)
	}

	// quick sanity check that the body is JSON-ish
	body := strings.TrimSpace(w.Body.String())
	if !strings.HasPrefix(body, "{") || !strings.HasSuffix(body, "}") {
		t.Fatalf("expected JSON object body for CLI UA, got %s", body)
	}
}

func TestUserAgent_Browser_ReturnsHTML(t *testing.T) {
	r := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html content-type for browser UA, got %s", ct)
	}

	// look for a DOCTYPE or <html> in the response body
	body := w.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") && !strings.Contains(body, "<html") {
		t.Fatalf("expected HTML body for browser UA, got %s", body[:len(body)])
	}
}
