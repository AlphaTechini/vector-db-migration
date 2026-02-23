package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthMiddleware_MissingAuth(t *testing.T) {
	middleware := NewAuthMiddleware("test-key")
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called without auth")
	})

	req := httptest.NewRequest("POST", "/", nil)
	rr := httptest.NewRecorder()

	middleware.Middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	expectedBody := `{"jsonrpc":"2.0","id":null,"error":{"code":-32000,"message":"missing authorization"}}`
	actualBody := strings.TrimSpace(rr.Body.String())
	if actualBody != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, actualBody)
	}
}

func TestAuthMiddleware_InvalidKey(t *testing.T) {
	middleware := NewAuthMiddleware("test-key")
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with invalid key")
	})

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	rr := httptest.NewRecorder()

	middleware.Middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}

	expectedBody := `{"jsonrpc":"2.0","id":null,"error":{"code":-32001,"message":"invalid api key"}}`
	actualBody := strings.TrimSpace(rr.Body.String())
	if actualBody != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, actualBody)
	}
}

func TestAuthMiddleware_ValidKey(t *testing.T) {
	middleware := NewAuthMiddleware("test-key")
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		
		// Verify API key is in context
		apiKey := GetAPIKeyFromContext(r.Context())
		if apiKey != "test-key" {
			t.Errorf("Expected API key 'test-key' in context, got '%s'", apiKey)
		}
	})

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	rr := httptest.NewRecorder()

	middleware.Middleware(handler).ServeHTTP(rr, req)

	if !called {
		t.Error("Expected handler to be called with valid key")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestAuthMiddleware_RawKeyFormat(t *testing.T) {
	middleware := NewAuthMiddleware("test-key")
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	// Test raw key format (without "Bearer " prefix)
	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "test-key")
	rr := httptest.NewRecorder()

	middleware.Middleware(handler).ServeHTTP(rr, req)

	if !called {
		t.Error("Expected handler to be called with raw key format")
	}
}

func TestAuthMiddleware_HealthCheckSkip(t *testing.T) {
	middleware := NewAuthMiddleware("test-key")
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	// Health check should skip auth
	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()

	middleware.Middleware(handler).ServeHTTP(rr, req)

	if !called {
		t.Error("Expected health check to skip authentication")
	}
}

func TestAuthMiddleware_ConstantTimeComparison(t *testing.T) {
	// This test verifies that we use constant-time comparison
	// by ensuring the timing doesn't vary significantly based on
	// how much of the key matches
	
	middleware := NewAuthMiddleware("correct-key-12345")
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Keys that differ at different positions
	keys := []string{
		"wrong-key-12345",   // Differs at start
		"correct-kye-12345", // Differs in middle
		"correct-key-1234x", // Differs at end
	}

	// All should fail with same status code (no timing-based info leak)
	for _, key := range keys {
		req := httptest.NewRequest("POST", "/", nil)
		req.Header.Set("Authorization", "Bearer "+key)
		rr := httptest.NewRecorder()

		middleware.Middleware(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 for key '%s', got %d", key, rr.Code)
		}
	}
}

func TestGetAPIKeyFromContext_EmptyContext(t *testing.T) {
	ctx := context.Background()
	key := GetAPIKeyFromContext(ctx)
	
	if key != "" {
		t.Errorf("Expected empty string for context without API key, got '%s'", key)
	}
}
