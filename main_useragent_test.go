package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

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
