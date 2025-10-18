package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	assert.True(t, strings.Contains(w.Body.String(), "<h1>🎲 Random String Generator</h1>"), "Response should contain the correct HTML title")
	assert.True(t, strings.Contains(w.Body.String(), "Printable String"), "Response should contain Printable string")
	assert.True(t, strings.Contains(w.Body.String(), "Alphanumeric String"), "Response should contain Alphanumeric string")
}
