package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda
var ginLambdaV2 *ginadapter.GinLambdaV2

// map of user-agent signatures considered CLI/programmatic clients for O(1) lookup
var cliSignaturesMap = map[string]struct{}{
	"curl":           {},
	"wget":           {},
	"powershell":     {},
	"httpie":         {},
	"python-requests":{},
	"python-urllib":  {},
	"go-http-client": {},
	"fetch":          {},
	"aria2":          {},
	"http_client":    {},
	"winhttp":        {},
	"axios":          {},
	"node-fetch":     {},
}

// Version/build info (set via -ldflags at build time)
var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "none"
)

// local random source to avoid using the deprecated global seed
var rnd *rand.Rand

func init() {
	if rnd == nil {
		rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
}

// RandomString struct for individual random strings
type RandomString struct {
	Length int    `json:"length"`
	String string `json:"string"`
}

// Response struct for JSON response
type Response struct {
	Printable    RandomString `json:"printable"`
	AlphaNumeric RandomString `json:"alphanumeric"`
}

// Function to generate random printable string
func GenerateRandomPrintable(length int) string {
	if length <= 0 {
		return ""
	}

	// First, generate a standard alphanumeric string.
	result := GenerateRandomAlphanumeric(length)
	runes := []rune(result)

	// Define the set of non-alphanumeric, printable characters.
	specialChars := []rune("!#$%*+-=?@^_")

	// Determine how many characters to replace (1 to 3, but not more than the string length).
	numReplacements := rnd.Intn(3) + 1
	if numReplacements >= length {
		numReplacements = 1
	}

	// If there are no special characters, we can't do replacements.
	if len(specialChars) == 0 {
		return string(runes)
	}

	// Replace characters at random positions.
	for i := 0; i < numReplacements; i++ {
		pos := rnd.Intn(length)
		runes[pos] = specialChars[rnd.Intn(len(specialChars))]
	}

	return string(runes)
}

// Function to generate random alphanumeric string
func GenerateRandomAlphanumeric(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	result := make([]rune, length)
	for i := range result {
		result[i] = letters[rnd.Intn(len(letters))]
	}
	return string(result)
}

func generateStrings(c *gin.Context) {
	// Default lengths for strings
	printableLength := rand.Intn(19) + 12    // Random length between 12 and 30
	alphanumericLength := rand.Intn(19) + 12 // Random length between 12 and 30

	// Read query parameters
	if val, ok := c.GetQuery("p"); ok {
		if length, err := strconv.Atoi(val); err == nil {
			printableLength = length
		}
	}
	if val, ok := c.GetQuery("a"); ok {
		if length, err := strconv.Atoi(val); err == nil {
			alphanumericLength = length
		}
	}

	if printableLength > 99 {
		printableLength = 99
	}
	if printableLength < 1 {
		printableLength = 1
	}
	if alphanumericLength > 99 {
		alphanumericLength = 99
	}
	if alphanumericLength < 1 {
		alphanumericLength = 1
	}

	// Detect CLI-like User-Agent (curl, wget, powershell, httpie, python-requests, etc.)
	ua := strings.ToLower(c.GetHeader("User-Agent"))
	isCLI := false
	if ua != "" {
		for _, sig := range cliSignatures {
			if strings.Contains(ua, sig) {
				isCLI = true
				break
			}
		}
	}

	// If requested path is /json or the client appears to be a CLI, return JSON
	if c.Request.URL.Path == "/json" || isCLI {
		response := Response{
			Printable: RandomString{
				Length: printableLength,
				String: GenerateRandomPrintable(printableLength),
			},
			AlphaNumeric: RandomString{
				Length: alphanumericLength,
				String: GenerateRandomAlphanumeric(alphanumericLength),
			},
		}
		// avoid caching so UI updates always fetch fresh values
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.IndentedJSON(http.StatusOK, response)
	} else {
		// Render HTML template with dynamic data
		tmpl, err := template.ParseFiles("static/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error loading template")
			return
		}

		data := map[string]interface{}{
			"PrintableLength":    printableLength,
			"PrintableString":    GenerateRandomPrintable(printableLength),
			"AlphanumericLength": alphanumericLength,
			"AlphanumericString": GenerateRandomAlphanumeric(alphanumericLength),
			"Version":            Version,
			"BuildTime":          BuildTime,
			"CommitHash":         CommitHash,
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		err = tmpl.Execute(c.Writer, data)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error rendering template")
		}
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	// Initialize a local random source for non-deterministic output across cold starts
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	r := gin.Default()

	// Serve static files (CSS, JS)
	r.Static("/static", "./static")

	// Define the endpoints
	r.GET("/json", generateStrings) // JSON response
	r.GET("/", generateStrings)     // HTML response

	// print out the Version, BuildTime and Commit Hash
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Commit Hash: %s\n", CommitHash)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// Running on AWS Lambda
		ginLambda = ginadapter.New(r)
		ginLambdaV2 = ginadapter.NewV2(r)
		lambda.Start(&universalHandler{v1: ginLambda, v2: ginLambdaV2})
	} else {
		// Running locally
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("failed to run server: %v", err)
		}
	}
}

// universalHandler routes between API Gateway v1, API Gateway v2, and Lambda Function URL events.
type universalHandler struct {
	v1 *ginadapter.GinLambda
	v2 *ginadapter.GinLambdaV2
}

func (h *universalHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	// Try standard event types first
	if result, err := h.tryAPIGatewayV2(ctx, payload); err == nil {
		return result, nil
	}
	if result, err := h.tryLambdaFunctionURL(ctx, payload); err == nil {
		return result, nil
	}
	if result, err := h.tryAPIGatewayV1(ctx, payload); err == nil {
		return result, nil
	}

	// Permissive fallback for generic/non-standard payloads
	if result, err := h.tryGenericPayload(ctx, payload); err == nil {
		return result, nil
	}

	// Final fallback: route to /json as GET
	return h.tryV1Fallback(ctx, payload)
}

// tryAPIGatewayV2 attempts to parse and handle an API Gateway v2 event.
func (h *universalHandler) tryAPIGatewayV2(ctx context.Context, payload []byte) ([]byte, error) {
	var v2req events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal(payload, &v2req); err != nil {
		return nil, err
	}
	// Accept if version is "2.0" OR if RequestContext has HTTP method
	if v2req.Version == "2.0" || v2req.RequestContext.HTTP.Method != "" {
		resp, err := h.v2.ProxyWithContext(ctx, v2req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}
	return nil, fmt.Errorf("not v2")
}

// tryLambdaFunctionURL attempts to parse and handle a Lambda Function URL event.
func (h *universalHandler) tryLambdaFunctionURL(ctx context.Context, payload []byte) ([]byte, error) {
	var furl events.LambdaFunctionURLRequest
	if err := json.Unmarshal(payload, &furl); err != nil {
		return nil, err
	}
	// Accept if RawPath is present OR if RequestContext has HTTP method
	if furl.RawPath != "" || furl.RequestContext.HTTP.Method != "" {
		converted := convertFunctionURLToV2(furl)
		resp, err := h.v2.ProxyWithContext(ctx, converted)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}
	return nil, fmt.Errorf("not function url")
}

// tryAPIGatewayV1 attempts to parse and handle an API Gateway v1 event.
func (h *universalHandler) tryAPIGatewayV1(ctx context.Context, payload []byte) ([]byte, error) {
	var v1req events.APIGatewayProxyRequest
	if err := json.Unmarshal(payload, &v1req); err != nil {
		return nil, err
	}
	// Accept if HTTPMethod is present OR Path is present OR RequestID is present
	if v1req.HTTPMethod != "" || v1req.Path != "" || v1req.RequestContext.RequestID != "" {
		resp, err := h.v1.ProxyWithContext(ctx, v1req)
		if err != nil {
			return nil, err
		}
		resp = sanitizeV1Response(resp)
		return json.Marshal(resp)
	}
	return nil, fmt.Errorf("not v1")
}

// tryGenericPayload attempts to detect and handle non-standard generic payloads.
func (h *universalHandler) tryGenericPayload(ctx context.Context, payload []byte) ([]byte, error) {
	var generic map[string]interface{}
	if err := json.Unmarshal(payload, &generic); err != nil {
		return nil, err
	}

	// Try to coerce to v2
	if v, ok := generic["version"].(string); ok && v == "2.0" {
		if b, err := json.Marshal(generic); err == nil {
			var v2req events.APIGatewayV2HTTPRequest
			if err := json.Unmarshal(b, &v2req); err == nil {
				resp, err := h.v2.ProxyWithContext(ctx, v2req)
				if err != nil {
					return nil, err
				}
				return json.Marshal(resp)
			}
		}
	}

	// Try to coerce to v1
	if _, hasHTTPMethod := generic["httpMethod"]; hasHTTPMethod || generic["path"] != nil || generic["resource"] != nil {
		if b, err := json.Marshal(generic); err == nil {
			var v1req events.APIGatewayProxyRequest
			if err := json.Unmarshal(b, &v1req); err == nil {
				resp, err := h.v1.ProxyWithContext(ctx, v1req)
				if err != nil {
					return nil, err
				}
				resp = sanitizeV1Response(resp)
				return json.Marshal(resp)
			}
		}
	}

	return nil, fmt.Errorf("unable to coerce generic payload")
}

// tryV1Fallback handles arbitrary payloads by routing to /json GET.
func (h *universalHandler) tryV1Fallback(ctx context.Context, payload []byte) ([]byte, error) {
	v1fallback := events.APIGatewayProxyRequest{
		Path:              "/json",
		HTTPMethod:        "GET",
		Headers:           map[string]string{"Content-Type": "application/json"},
		MultiValueHeaders: map[string][]string{},
		Body:              string(payload),
		IsBase64Encoded:   false,
	}
	resp, err := h.v1.ProxyWithContext(ctx, v1fallback)
	if err != nil {
		return nil, fmt.Errorf("unable to coerce generic payload to v1 proxy: %w", err)
	}
	resp = sanitizeV1Response(resp)
	return json.Marshal(resp)
}

// sanitizeV1Response ensures a v1 response has non-nil maps and base64 flag set correctly.
func sanitizeV1Response(resp events.APIGatewayProxyResponse) events.APIGatewayProxyResponse {
	if resp.Headers == nil {
		resp.Headers = map[string]string{}
	}
	if resp.MultiValueHeaders == nil {
		resp.MultiValueHeaders = map[string][]string{}
	}
	resp.IsBase64Encoded = false
	return resp
}

// convertFunctionURLToV2 maps a Lambda Function URL event to an APIGateway v2 HTTP request for the adapter.
func convertFunctionURLToV2(f events.LambdaFunctionURLRequest) events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		Version:               "2.0",
		RouteKey:              "$default",
		RawPath:               f.RawPath,
		RawQueryString:        f.RawQueryString,
		Cookies:               f.Cookies,
		Headers:               f.Headers,
		QueryStringParameters: f.QueryStringParameters,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			AccountID:  f.RequestContext.AccountID,
			RequestID:  f.RequestContext.RequestID,
			DomainName: f.RequestContext.DomainName,
			TimeEpoch:  f.RequestContext.TimeEpoch,
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method:    f.RequestContext.HTTP.Method,
				Path:      f.RequestContext.HTTP.Path,
				Protocol:  f.RequestContext.HTTP.Protocol,
				SourceIP:  f.RequestContext.HTTP.SourceIP,
				UserAgent: f.RequestContext.HTTP.UserAgent,
			},
		},
		Body:            f.Body,
		IsBase64Encoded: f.IsBase64Encoded,
	}
}
