package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda
var (
	// Version information
	version    string
	buildTime  string
	commitHash string
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

	// First, generate a standard alphanumeric string.
	result := GenerateRandomAlphanumeric(length)
	runes := []rune(result)

	// Define the set of non-alphanumeric, printable characters.
	specialChars := []rune("!#$%*+-=?@^_")

	// Determine how many characters to replace (1 to 3, but not more than the string length).
	numReplacements := rand.Intn(3) + 1
	if numReplacements >= length {
		numReplacements = 1
	}

	// If there are no special characters, we can't do replacements.
	if len(specialChars) == 0 {
		return string(runes)
	}

	// Replace characters at random positions.
	for i := 0; i < numReplacements; i++ {
		pos := rand.Intn(length)
		runes[pos] = specialChars[rand.Intn(len(specialChars))]
	}

	return string(runes)
}

// Function to generate random alphanumeric string
func GenerateRandomAlphanumeric(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	result := make([]rune, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
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
            width: 30px;
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
            
            fetch(url)
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
		lambda.Start(Handler)
	} else {
		// Running locally
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("failed to run server: %v", err)
		}
	}
}

// Handler is the Lambda function handler
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}
