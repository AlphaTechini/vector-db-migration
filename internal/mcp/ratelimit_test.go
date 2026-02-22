package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterMiddleware_AllowWithinLimit(t *testing.T) {
	// 10 requests per minute, burst of 5
	middleware := NewRateLimiterMiddleware(10, 5)
	
	called := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
	})

	// First 5 requests should succeed (burst)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/", nil)
		ctx := context.WithValue(req.Context(), ContextKeyAPIKey, "test-key")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		middleware.Middleware(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, rr.Code)
		}
	}

	if called != 5 {
		t.Errorf("Expected handler to be called 5 times, got %d", called)
	}
}

func TestRateLimiterMiddleware_RejectOverLimit(t *testing.T) {
	// 10 requests per minute, burst of 3
	middleware := NewRateLimiterMiddleware(10, 3)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// First 3 requests should succeed (burst)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/", nil)
		ctx := context.WithValue(req.Context(), ContextKeyAPIKey, "test-key")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		middleware.Middleware(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, rr.Code)
		}
	}

	// 4th request should fail (over limit)
	req := httptest.NewRequest("POST", "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyAPIKey, "test-key")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware.Middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr.Code)
	}

	expectedBody := `{"jsonrpc":"2.0","id":null,"error":{"code":-32002,"message":"rate limit exceeded"}}`
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, rr.Body.String())
	}
}

func TestRateLimiterMiddleware_SeparateKeys(t *testing.T) {
	// 10 requests per minute, burst of 2
	middleware := NewRateLimiterMiddleware(10, 2)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Use up limit for key1
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/", nil)
		ctx := context.WithValue(req.Context(), ContextKeyAPIKey, "key1")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		middleware.Middleware(handler).ServeHTTP(rr, req)
	}

	// key2 should still have full burst
	req := httptest.NewRequest("POST", "/", nil)
	ctx := context.WithValue(req.Context(), ContextKeyAPIKey, "key2")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware.Middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected key2 to have separate limit, got status %d", rr.Code)
	}
}

func TestRateLimiterMiddleware_AnonymousFallback(t *testing.T) {
	middleware := NewRateLimiterMiddleware(10, 5)
	
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	// No API key in context (shouldn't happen with auth enabled)
	req := httptest.NewRequest("POST", "/", nil)
	rr := httptest.NewRecorder()

	middleware.Middleware(handler).ServeHTTP(rr, req)

	if !called {
		t.Error("Expected handler to be called with anonymous fallback")
	}
}

func TestRateLimiterMiddleware_ConcurrentAccess(t *testing.T) {
	middleware := NewRateLimiterMiddleware(60, 10) // High limits for concurrent test
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	done := make(chan bool, 20)

	// 20 concurrent requests with same API key
	for i := 0; i < 20; i++ {
		go func() {
			req := httptest.NewRequest("POST", "/", nil)
			ctx := context.WithValue(req.Context(), ContextKeyAPIKey, "concurrent-key")
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			middleware.Middleware(handler).ServeHTTP(rr, req)
			done <- rr.Code == http.StatusOK
		}()
	}

	// Count successful requests
	successes := 0
	for i := 0; i < 20; i++ {
		if <-done {
			successes++
		}
	}

	// Should allow burst (10) + some from refill
	if successes < 10 || successes > 15 {
		t.Errorf("Expected 10-15 successes, got %d", successes)
	}
}

func TestRateLimiterMiddleware_Cleanup(t *testing.T) {
	middleware := NewRateLimiterMiddleware(10, 5)
	
	// Create some limiters
	middleware.getLimiter("key1")
	middleware.getLimiter("key2")
	middleware.getLimiter("key3")

	// Cleanup (currently a no-op, but tests the method exists)
	middleware.Cleanup(1 * time.Hour)

	// Verify limiters still exist (cleanup not implemented yet)
	if len(middleware.limiters) != 3 {
		t.Errorf("Expected 3 limiters, got %d", len(middleware.limiters))
	}
}
