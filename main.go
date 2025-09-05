package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda
var ginLambdaV2 *ginadapter.GinLambdaV2
var (
	// Version information
	version    string
	buildTime  string
	commitHash string
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

	// Check the endpoint
	if c.Request.URL.Path == "/json" {
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
		// Create a plain HTML string response
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Generated Random Strings</title>
	<style>
		input[type=number] {
			width: 50px;
			text-align: right;
		}
	</style>
    <script>
        function refreshStrings() {
            var printableLength = document.getElementById("p").value;
            if (printableLength > 99) {
                printableLength = 99;
                document.getElementById("p").value = 99;
            }
            if (printableLength < 1) {
                printableLength = 1;
                document.getElementById("p").value = 1;
            }
            var alphanumericLength = document.getElementById("a").value;
            if (alphanumericLength > 99) {
                alphanumericLength = 99;
                document.getElementById("a").value = 99;
            }
            if (alphanumericLength < 1) {
                alphanumericLength = 1;
                document.getElementById("a").value = 1;
            }
            var url = "/json?p=" + printableLength + "&a=" + alphanumericLength;
            
			fetch(url, {cache: 'no-store'})
				.then(response => response.json())
                .then(data => {
                    document.getElementById("printable-string").textContent = data.printable.string;
                    document.getElementById("alphanumeric-string").textContent = data.alphanumeric.string;
                });
        }
    </script>
</head>
<body>
    <h1>Generated Random Strings</h1>
    <p>Printable: <input type="number" id="p" name="p" value="` + strconv.Itoa(printableLength) + `" oninput="refreshStrings()" min="1" max="99"> <span id="printable-string">` + GenerateRandomPrintable(printableLength) + `</span></p>
    <p>Alphanumeric: <input type="number" id="a" name="a" value="` + strconv.Itoa(alphanumericLength) + `" oninput="refreshStrings()" min="1" max="99"> <span id="alphanumeric-string">` + GenerateRandomAlphanumeric(alphanumericLength) + `</span></p>
</body>
</html>`

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	// Initialize a local random source for non-deterministic output across cold starts
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	r := gin.Default()

	// Define the endpoints
	r.GET("/json", generateStrings) // JSON response
	r.GET("/", generateStrings)     // HTML response

	// print out the version, buildtime and commit hash
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Build Time: %s\n", buildTime)
	fmt.Printf("Commit Hash: %s\n", commitHash)

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
	// Try API Gateway v2 (also covers Function URL once converted below)
	var v2req events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal(payload, &v2req); err == nil && (v2req.Version == "2.0" || v2req.RequestContext.HTTP.Method != "") {
		resp, err := h.v2.ProxyWithContext(ctx, v2req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}

	// Try Lambda Function URL event -> convert to APIGW v2 request
	var furl events.LambdaFunctionURLRequest
	if err := json.Unmarshal(payload, &furl); err == nil && (furl.RawPath != "" || furl.RequestContext.HTTP.Method != "") {
		converted := convertFunctionURLToV2(furl)
		resp, err := h.v2.ProxyWithContext(ctx, converted)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}

	// Fallback to API Gateway v1
	var v1req events.APIGatewayProxyRequest
	if err := json.Unmarshal(payload, &v1req); err == nil && (v1req.HTTPMethod != "" || v1req.Path != "" || v1req.RequestContext.RequestID != "") {
		resp, err := h.v1.ProxyWithContext(ctx, v1req)
		if err != nil {
			return nil, err
		}
		// Sanitize for REST API proxy expectations: avoid null maps and set base64 flag explicitly.
		if resp.Headers == nil {
			resp.Headers = map[string]string{}
		}
		if resp.MultiValueHeaders == nil {
			resp.MultiValueHeaders = map[string][]string{}
		}
		resp.IsBase64Encoded = false
		return json.Marshal(resp)
	}

	// Permissive fallback: some console or custom test events use non-standard shapes.
	// Try to coerce generic JSON payloads into v2 or v1 shapes by inspecting keys.
	var generic map[string]interface{}
	if err := json.Unmarshal(payload, &generic); err == nil {
		// Detect v2-like (version = "2.0" or requestContext.http present)
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

		if _, hasHTTPMethod := generic["httpMethod"]; hasHTTPMethod || generic["path"] != nil || generic["resource"] != nil {
			if b, err := json.Marshal(generic); err == nil {
				var v1 events.APIGatewayProxyRequest
				if err := json.Unmarshal(b, &v1); err == nil {
					resp, err := h.v1.ProxyWithContext(ctx, v1)
					if err != nil {
						return nil, err
					}
					if resp.Headers == nil {
						resp.Headers = map[string]string{}
					}
					if resp.MultiValueHeaders == nil {
						resp.MultiValueHeaders = map[string][]string{}
					}
					resp.IsBase64Encoded = false
					return json.Marshal(resp)
				}
			}
		}
	}

	// Final permissive fallback: forward raw payload as v1 POST / body so console tests succeed.
	// This keeps the function usable from the Lambda console with arbitrary test events.
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
	if resp.Headers == nil {
		resp.Headers = map[string]string{}
	}
	if resp.MultiValueHeaders == nil {
		resp.MultiValueHeaders = map[string][]string{}
	}
	resp.IsBase64Encoded = false
	return json.Marshal(resp)
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
