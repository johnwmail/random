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
		// Create a modern, responsive HTML string response
		html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Random String Generator</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
            line-height: 1.6;
        }

        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 40px;
            max-width: 600px;
            width: 100%;
            animation: fadeIn 0.5s ease-in;
        }

        @keyframes fadeIn {
            from {
                opacity: 0;
                transform: translateY(20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        h1 {
            color: #333;
            font-size: 28px;
            margin-bottom: 10px;
            text-align: center;
        }

        .subtitle {
            color: #666;
            text-align: center;
            margin-bottom: 30px;
            font-size: 14px;
        }

        .string-card {
            background: #f8f9fa;
            border-radius: 12px;
            padding: 20px;
            margin-bottom: 20px;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }

        .string-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }

        .card-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 15px;
            flex-wrap: wrap;
            gap: 10px;
        }

        .card-title {
            font-size: 16px;
            font-weight: 600;
            color: #495057;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .length-control {
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .length-label {
            font-size: 14px;
            color: #6c757d;
        }

        input[type=number] {
            width: 60px;
            padding: 8px 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
            text-align: center;
            transition: border-color 0.2s ease, box-shadow 0.2s ease;
            background: white;
        }

        input[type=number]:focus {
            outline: none;
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }

        input[type=number]::-webkit-inner-spin-button,
        input[type=number]::-webkit-outer-spin-button {
            opacity: 1;
        }

        .string-display {
            background: white;
            padding: 15px;
            border-radius: 8px;
            border: 2px solid #e0e0e0;
            font-family: 'Courier New', monospace;
            font-size: 16px;
            word-break: break-all;
            color: #212529;
            position: relative;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 10px;
        }

        .string-text {
            flex: 1;
            min-width: 0;
        }

        .copy-btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 8px 12px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 12px;
            font-weight: 500;
            transition: all 0.2s ease;
            white-space: nowrap;
            flex-shrink: 0;
        }

        .copy-btn:hover {
            background: #5568d3;
            transform: translateY(-1px);
        }

        .copy-btn:active {
            transform: translateY(0);
        }

        .copy-btn.copied {
            background: #10b981;
        }

        .refresh-btn {
            width: 100%;
            padding: 14px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 10px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            margin-top: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 8px;
        }

        .refresh-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
        }

        .refresh-btn:active {
            transform: translateY(0);
        }

        .icon {
            display: inline-block;
            width: 18px;
            height: 18px;
        }

        @media (max-width: 600px) {
            .container {
                padding: 25px;
            }

            h1 {
                font-size: 24px;
            }

            .card-header {
                flex-direction: column;
                align-items: flex-start;
            }

            .string-display {
                font-size: 14px;
                flex-direction: column;
                align-items: stretch;
            }

            .copy-btn {
                width: 100%;
            }
        }

        @media (prefers-color-scheme: dark) {
            /* Optional: Add dark mode support in the future */
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

        function copyToClipboard(elementId, buttonId) {
            var text = document.getElementById(elementId).textContent;
            var button = document.getElementById(buttonId);
            
            navigator.clipboard.writeText(text).then(function() {
                var originalText = button.textContent;
                button.textContent = '✓ Copied';
                button.classList.add('copied');
                
                setTimeout(function() {
                    button.textContent = originalText;
                    button.classList.remove('copied');
                }, 2000);
            }).catch(function(err) {
                console.error('Failed to copy:', err);
            });
        }
    </script>
</head>
<body>
    <div class="container">
        <h1>🎲 Random String Generator</h1>
        <p class="subtitle">Generate secure random strings instantly</p>
        
        <div class="string-card">
            <div class="card-header">
                <div class="card-title">
                    <span>🔒</span>
                    <span>Printable String</span>
                </div>
                <div class="length-control">
                    <span class="length-label">Length:</span>
                    <input type="number" id="p" name="p" value="` + strconv.Itoa(printableLength) + `" oninput="refreshStrings()" min="1" max="99">
                </div>
            </div>
            <div class="string-display">
                <span class="string-text" id="printable-string">` + GenerateRandomPrintable(printableLength) + `</span>
                <button class="copy-btn" id="copy-p" onclick="copyToClipboard('printable-string', 'copy-p')">Copy</button>
            </div>
        </div>

        <div class="string-card">
            <div class="card-header">
                <div class="card-title">
                    <span>🔤</span>
                    <span>Alphanumeric String</span>
                </div>
                <div class="length-control">
                    <span class="length-label">Length:</span>
                    <input type="number" id="a" name="a" value="` + strconv.Itoa(alphanumericLength) + `" oninput="refreshStrings()" min="1" max="99">
                </div>
            </div>
            <div class="string-display">
                <span class="string-text" id="alphanumeric-string">` + GenerateRandomAlphanumeric(alphanumericLength) + `</span>
                <button class="copy-btn" id="copy-a" onclick="copyToClipboard('alphanumeric-string', 'copy-a')">Copy</button>
            </div>
        </div>

        <button class="refresh-btn" onclick="refreshStrings()">
            <span class="icon">🔄</span>
            <span>Generate New Strings</span>
        </button>
    </div>
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
