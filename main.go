package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda
var ginLambdaV2 *ginadapter.GinLambdaV2

// map of user-agent signatures considered CLI/programmatic clients for O(1) lookup
var cliSignaturesMap = map[string]struct{}{
	"curl":            {},
	"wget":            {},
	"powershell":      {},
	"httpie":          {},
	"python-requests": {},
	"python-urllib":   {},
	"go-http-client":  {},
	"fetch":           {},
	"aria2":           {},
	"http_client":     {},
	"winhttp":         {},
	"axios":           {},
	"node-fetch":      {},
}

const MaxAllowedLength = 100

// cryptoRandInt generates a cryptographically secure random integer in the range [0, max)
func cryptoRandInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		log.Fatalf("crypto/rand failed: %v", err)
	}
	return int(n.Int64())
}

// parseLengths extracts and clamps printable and alphanumeric lengths from the request
func parseLengths(c *gin.Context) (int, int) {
	printableLength := cryptoRandInt(19) + 12    // Random length between 12 and 30
	alphanumericLength := cryptoRandInt(19) + 12 // Random length between 12 and 30

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

	return printableLength, alphanumericLength
}

// buildResponse creates the Response payload for JSON responses
func buildResponse(printableLength, alphanumericLength int) Response {
	return Response{
		Printable: RandomString{
			Length: printableLength,
			String: GenerateRandomPrintable(printableLength),
		},
		AlphaNumeric: RandomString{
			Length: alphanumericLength,
			String: GenerateRandomAlphanumeric(alphanumericLength),
		},
	}
}

// isCLIUserAgent returns true when the provided user-agent string matches common CLI clients
func isCLIUserAgent(ua string) bool {
	ua = strings.ToLower(ua)
	if ua == "" {
		return false
	}
	for sig := range cliSignaturesMap {
		if strings.Contains(ua, sig) {
			return true
		}
	}
	return false
}

// Version/build info (set via -ldflags at build time)
var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "none"
)

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
	if length > MaxAllowedLength {
		length = MaxAllowedLength
	}

	// First, generate a standard alphanumeric string.
	result := GenerateRandomAlphanumeric(length)
	runes := []rune(result)

	// Define the set of non-alphanumeric, printable characters.
	specialChars := []rune("!#$%*+-=?@^_")

	// Determine how many characters to replace (1 to 3, but not more than the string length).
	numReplacements := cryptoRandInt(3) + 1
	if numReplacements >= length {
		numReplacements = 1
	}

	// If there are no special characters, we can't do replacements.
	if len(specialChars) == 0 {
		return string(runes)
	}

	// Replace characters at random positions.
	for i := 0; i < numReplacements; i++ {
		pos := cryptoRandInt(length)
		runes[pos] = specialChars[cryptoRandInt(len(specialChars))]
	}

	return string(runes)
}

// Function to generate random alphanumeric string
func GenerateRandomAlphanumeric(length int) string {
	if length <= 0 {
		return ""
	}
	if length > MaxAllowedLength {
		length = MaxAllowedLength
	}
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	result := make([]rune, length)
	for i := range result {
		result[i] = letters[cryptoRandInt(len(letters))]
	}
	return string(result)
}

func generateStrings(c *gin.Context) {
	printableLength, alphanumericLength := parseLengths(c)

	ua := c.GetHeader("User-Agent")
	if c.Request.URL.Path == "/json" || isCLIUserAgent(ua) {
		response := buildResponse(printableLength, alphanumericLength)
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.IndentedJSON(http.StatusOK, response)
		return
	}

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

func main() {
	gin.SetMode(gin.ReleaseMode)
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
	// Define event handlers in order of preference
	handlers := []func(context.Context, []byte) ([]byte, error){
		h.tryAPIGatewayV2,
		h.tryLambdaFunctionURL,
		h.tryAPIGatewayV1,
		h.tryGenericPayload,
	}

	// Try each handler in sequence
	for _, handler := range handlers {
		if result, err := handler(ctx, payload); err == nil {
			return result, nil
		}
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
		// Copy Content-Type from MultiValueHeaders to Headers for ALB compatibility
		if ctype, ok := resp.MultiValueHeaders["Content-Type"]; ok && len(ctype) > 0 {
			resp.Headers["Content-Type"] = ctype[0]
		}
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

	// Try v2 coercion first
	if result, err := h.coerceToV2(ctx, generic); err == nil {
		return result, nil
	}

	// Try v1 coercion
	if result, err := h.coerceToV1(ctx, generic); err == nil {
		return result, nil
	}

	return nil, fmt.Errorf("unable to coerce generic payload")
}

// coerceToV2 attempts to convert generic payload to API Gateway v2 format
func (h *universalHandler) coerceToV2(ctx context.Context, generic map[string]interface{}) ([]byte, error) {
	v, ok := generic["version"].(string)
	if !ok || v != "2.0" {
		return nil, fmt.Errorf("not v2 format")
	}

	b, err := json.Marshal(generic)
	if err != nil {
		return nil, err
	}

	var v2req events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal(b, &v2req); err != nil {
		return nil, err
	}

	resp, err := h.v2.ProxyWithContext(ctx, v2req)
	if err != nil {
		return nil, err
	}
	return json.Marshal(resp)
}

// coerceToV1 attempts to convert generic payload to API Gateway v1 format
func (h *universalHandler) coerceToV1(ctx context.Context, generic map[string]interface{}) ([]byte, error) {
	_, hasHTTPMethod := generic["httpMethod"]
	if !hasHTTPMethod && generic["path"] == nil && generic["resource"] == nil {
		return nil, fmt.Errorf("not v1 format")
	}

	b, err := json.Marshal(generic)
	if err != nil {
		return nil, err
	}

	var v1req events.APIGatewayProxyRequest
	if err := json.Unmarshal(b, &v1req); err != nil {
		return nil, err
	}

	resp, err := h.v1.ProxyWithContext(ctx, v1req)
	if err != nil {
		return nil, err
	}
	resp = sanitizeV1Response(resp)
	// Copy Content-Type from MultiValueHeaders to Headers for ALB compatibility
	if ctype, ok := resp.MultiValueHeaders["Content-Type"]; ok && len(ctype) > 0 {
		resp.Headers["Content-Type"] = ctype[0]
	}
	return json.Marshal(resp)
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
	// Copy Content-Type from MultiValueHeaders to Headers for ALB compatibility
	if ctype, ok := resp.MultiValueHeaders["Content-Type"]; ok && len(ctype) > 0 {
		resp.Headers["Content-Type"] = ctype[0]
	}
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
