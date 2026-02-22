package mcp

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAuditMiddleware_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	
	middleware := NewAuditMiddleware(logger)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()

	middleware.Middleware(handler).ServeHTTP(rr, req)

	logOutput := buf.String()
	
	if !strings.Contains(logOutput, "[AUDIT]") {
		t.Error("Expected audit log entry")
	}
	
	if !strings.Contains(logOutput, "POST") {
		t.Error("Expected method in log")
	}
	
	if !strings.Contains(logOutput, "127.0.0.1") {
		t.Error("Expected client IP in log")
	}
}

func TestAuditMiddleware_LogsResponse(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	
	middleware := NewAuditMiddleware(logger)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest("POST", "/", nil)
	rr := httptest.NewRecorder()

	middleware.Middleware(handler).ServeHTTP(rr, req)

	logOutput := buf.String()
	
	// Should have both request and response logs
	if strings.Count(logOutput, "[AUDIT]") < 2 {
		t.Errorf("Expected at least 2 audit entries, got %d", strings.Count(logOutput, "[AUDIT]"))
	}
	
	if !strings.Contains(logOutput, "completed") {
		t.Error("Expected completion log")
	}
	
	if !strings.Contains(logOutput, "201") {
		t.Error("Expected status code in log")
	}
}

func TestAuditMiddleware_MasksAPIKey(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	
	middleware := NewAuthMiddleware("secret-key-1234")
	audit := NewAuditMiddleware(logger)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "Bearer secret-key-1234")
	rr := httptest.NewRecorder()

	// Chain: audit → auth → handler
	chain := audit.Middleware(middleware.Middleware(handler))
	chain.ServeHTTP(rr, req)

	logOutput := buf.String()
	
	// API key should be masked in logs
	if strings.Contains(logOutput, "secret-key-1234") {
		t.Error("Expected API key to be masked in logs")
	}
	
	// Should show masked version
	if !strings.Contains(logOutput, "****") {
		t.Error("Expected masked key indicator")
	}
}

func TestMaskString_ShortString(t *testing.T) {
	result := maskString("abc", 4)
	expected := "****"
	
	if result != expected {
		t.Errorf("Expected '%s' for short string, got '%s'", expected, result)
	}
}

func TestMaskString_LongString(t *testing.T) {
	result := maskString("secret-key-1234", 4)
	expected := "****1234"
	
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestMaskString_EmptyString(t *testing.T) {
	result := maskString("", 4)
	expected := "****"
	
	if result != expected {
		t.Errorf("Expected '%s' for empty string, got '%s'", expected, result)
	}
}

func TestAuditEntry_JSONSerialization(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	
	entry := AuditEntry{
		Timestamp:  time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC),
		EventType:  "request",
		APIKey:     "secret-key",
		Method:     "POST",
		Path:       "/",
		ClientIP:   "127.0.0.1",
		StatusCode: 200,
		DurationMs: 15,
		ToolName:   "migration_status",
	}

	LogAuditEntry(logger, entry)

	logOutput := buf.String()
	
	// Should be valid JSON
	if !strings.Contains(logOutput, "{") {
		t.Error("Expected JSON output")
	}
	
	// API key should be masked
	if strings.Contains(logOutput, "secret-key") {
		t.Error("Expected API key to be masked in structured log")
	}
}

func TestResponseWriter_WrapsCorrectly(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	
	middleware := NewAuditMiddleware(logger)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("response body"))
	})

	req := httptest.NewRequest("POST", "/", nil)
	rr := httptest.NewRecorder()

	middleware.Middleware(handler).ServeHTTP(rr, req)

	logOutput := buf.String()
	
	// Should capture the actual status code (202)
	if !strings.Contains(logOutput, "202") {
		t.Error("Expected wrapped status code to be logged")
	}
}

func TestAuditMiddleware_DurationTracking(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	
	middleware := NewAuditMiddleware(logger)
	
	// Handler with known delay
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
	})

	req := httptest.NewRequest("POST", "/", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	middleware.Middleware(handler).ServeHTTP(rr, req)
	duration := time.Since(start)

	logOutput := buf.String()
	
	// Log should mention duration
	if !strings.Contains(logOutput, "completed in") {
		t.Error("Expected duration in log")
	}
	
	// Duration should be reasonable (>50ms due to sleep)
	if duration < 50*time.Millisecond {
		t.Error("Expected handler to take at least 50ms")
	}
}
