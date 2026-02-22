package mcp

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// AuditMiddleware logs all MCP requests for security auditing
type AuditMiddleware struct {
	logger *log.Logger
}

// NewAuditMiddleware creates a new audit logging middleware
func NewAuditMiddleware(logger *log.Logger) *AuditMiddleware {
	return &AuditMiddleware{
		logger: logger,
	}
}

// Middleware wraps an http.Handler with audit logging
func (m *AuditMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Extract request details
		apiKey := GetAPIKeyFromContext(r.Context())
		method := r.Method
		path := r.URL.Path
		clientIP := r.RemoteAddr

		// Log request
		m.logger.Printf("[AUDIT] %s %s from %s (key: %s)",
			method, path, clientIP, maskString(apiKey, 4))

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Log response
		duration := time.Since(start)
		m.logger.Printf("[AUDIT] %s %s completed in %v with status %d",
			method, path, duration, wrapped.statusCode)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// maskString masks all but the last N characters
func maskString(s string, keepLast int) string {
	if len(s) <= keepLast {
		return "****"
	}
	return "****" + s[len(s)-keepLast:]
}

// AuditEntry represents a structured audit log entry
type AuditEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	EventType  string    `json:"event_type"`
	APIKey     string    `json:"api_key_masked"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	ClientIP   string    `json:"client_ip"`
	StatusCode int       `json:"status_code"`
	DurationMs int64     `json:"duration_ms"`
	ToolName   string    `json:"tool_name,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// LogAuditEntry writes a structured audit log entry
func LogAuditEntry(logger *log.Logger, entry AuditEntry) {
	entry.APIKey = maskString(entry.APIKey, 4)
	entry.DurationMs = time.Duration(entry.DurationMs).Milliseconds()

	bytes, _ := json.Marshal(entry)
	logger.Printf("[AUDIT] %s", string(bytes))
}
