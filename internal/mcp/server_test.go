package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	registry := NewToolRegistry()
	server := NewServer(":8080", registry,
		WithAPIKey("test-key"),
		WithRateLimit(100, 10),
		WithAuditLog(log.New(io.Discard, "", 0)),
	)

	if server.addr != ":8080" {
		t.Errorf("Expected addr :8080, got %s", server.addr)
	}
	if server.auth == nil {
		t.Error("Expected auth middleware to be initialized")
	}
	if server.rateLimiter == nil {
		t.Error("Expected rateLimiter middleware to be initialized")
	}
	if server.audit == nil {
		t.Error("Expected audit middleware to be initialized")
	}
}

func TestServer_StartStop(t *testing.T) {
	registry := NewToolRegistry()
	// Use port 0 for random available port
	server := NewServer("127.0.0.1:0", registry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start(ctx)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Try starting again (should fail)
	err := server.Start(ctx)
	if err == nil || !strings.Contains(err.Error(), "already started") {
		t.Errorf("Expected 'already started' error, got %v", err)
	}

	// Stop the server
	err = server.Stop()
	if err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	// Wait for Start to return
	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Start returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for server to stop")
	}
}

func TestServer_HandleRequest_Methods(t *testing.T) {
	registry := NewToolRegistry()
	server := NewServer(":8080", registry)

	// Test GET (should be rejected)
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	server.handleRequest(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for GET, got %d", rr.Code)
	}

	var errResp ErrorResponse
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp.Error.Code != InvalidRequest {
		t.Errorf("Expected error code %d, got %d", InvalidRequest, errResp.Error.Code)
	}
}

func TestServer_HandleRequest_InvalidJSON(t *testing.T) {
	registry := NewToolRegistry()
	server := NewServer(":8080", registry)

	req := httptest.NewRequest("POST", "/", strings.NewReader("invalid json"))
	rr := httptest.NewRecorder()
	server.handleRequest(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}

	var errResp ErrorResponse
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp.Error.Code != ParseError {
		t.Errorf("Expected error code %d, got %d", ParseError, errResp.Error.Code)
	}
}

func TestServer_HandleRequest_InvalidVersion(t *testing.T) {
	registry := NewToolRegistry()
	server := NewServer(":8080", registry)

	payload := Request{
		JSONRPC: "1.0", // Invalid version
		ID:      1,
		Method:  "test",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	server.handleRequest(rr, req)

	var errResp ErrorResponse
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp.Error.Code != InvalidRequest {
		t.Errorf("Expected error code %d, got %d", InvalidRequest, errResp.Error.Code)
	}
}

func TestServer_HandleRequest_Success(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(&Tool{
		Name: "echo",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{} , error) {
			return params["msg"], nil
		},
	})

	server := NewServer(":8080", registry)

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "req-1",
		"method":  "echo",
		"params": map[string]interface{}{
			"msg": "hello world",
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	server.handleRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var resp Response
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.ID != "req-1" {
		t.Errorf("Expected ID req-1, got %v", resp.ID)
	}
	if resp.Result != "hello world" {
		t.Errorf("Expected result 'hello world', got %v", resp.Result)
	}
}

func TestServer_HandleRequest_ToolError(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(&Tool{
		Name: "fail",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, fmt.Errorf("tool failed")
		},
	})

	server := NewServer(":8080", registry)

	payload := Request{
		JSONRPC: "2.0",
		ID:      123,
		Method:  "fail",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	server.handleRequest(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}

	var errResp ErrorResponse
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp.Error.Code != InternalError {
		t.Errorf("Expected error code %d, got %d", InternalError, errResp.Error.Code)
	}
	if !strings.Contains(errResp.Error.Message, "tool failed") {
		t.Errorf("Expected error message to contain 'tool failed', got '%s'", errResp.Error.Message)
	}
}

func TestServer_MiddlewareChain(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(&Tool{
		Name: "test",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	})

	var auditBuf bytes.Buffer
	logger := log.New(&auditBuf, "", 0)

	server := NewServer(":8080", registry,
		WithAPIKey("secret"),
		WithRateLimit(60, 1),
		WithAuditLog(logger),
	)

	// Create a test server with the middleware chain
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleRequest)

	var handler http.Handler = mux
	if server.rateLimiter != nil {
		handler = server.rateLimiter.Middleware(handler)
	}
	if server.auth != nil {
		handler = server.auth.Middleware(handler)
	}
	if server.audit != nil {
		handler = server.audit.Middleware(handler)
	}

	// 1. Unauthenticated request
	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"test"}`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing auth, got %d", rr.Code)
	}

	// Audit log should still have captured the request
	if !strings.Contains(auditBuf.String(), "[AUDIT] POST /") {
		t.Error("Audit log should contain the request even if unauthorized")
	}
	if !strings.Contains(auditBuf.String(), "completed in") || !strings.Contains(auditBuf.String(), "with status 401") {
		t.Error("Audit log should contain completion and status 401")
	}

	auditBuf.Reset()

	// 2. Authenticated request
	req = httptest.NewRequest("POST", "/", bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"test"}`)))
	req.Header.Set("Authorization", "Bearer secret")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(auditBuf.String(), "with status 200") {
		t.Error("Audit log should contain status 200")
	}

	auditBuf.Reset()

	// 3. Rate limited request (burst was 1)
	req = httptest.NewRequest("POST", "/", bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"test"}`)))
	req.Header.Set("Authorization", "Bearer secret")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429 for rate limit, got %d", rr.Code)
	}
	if !strings.Contains(auditBuf.String(), "with status 429") {
		t.Error("Audit log should contain status 429")
	}
}
