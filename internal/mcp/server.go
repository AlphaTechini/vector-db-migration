package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Server represents an MCP server
type Server struct {
	addr     string
	registry *ToolRegistry
	server   *http.Server
	mu       sync.Mutex
	
	// Middleware components (optional)
	auth        *AuthMiddleware
	rateLimiter *RateLimiterMiddleware
	audit       *AuditMiddleware
}

// ServerOption configures a Server
type ServerOption func(*Server)

// WithAPIKey enables API key authentication
func WithAPIKey(apiKey string) ServerOption {
	return func(s *Server) {
		s.auth = NewAuthMiddleware(apiKey)
	}
}

// WithRateLimit enables rate limiting
func WithRateLimit(requestsPerMinute, burst int) ServerOption {
	return func(s *Server) {
		s.rateLimiter = NewRateLimiterMiddleware(requestsPerMinute, burst)
	}
}

// WithAuditLog enables audit logging
func WithAuditLog(logger *log.Logger) ServerOption {
	return func(s *Server) {
		s.audit = NewAuditMiddleware(logger)
	}
}

// NewServer creates a new MCP server with optional middleware
func NewServer(addr string, registry *ToolRegistry, opts ...ServerOption) *Server {
	s := &Server{
		addr:     addr,
		registry: registry,
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// GetHandler returns the HTTP handler with all middleware applied
func (s *Server) GetHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)

	// Build middleware chain.
	// Middleware applied LAST wraps OUTERMOST and therefore executes FIRST.
	// Execution order: audit → auth → rateLimiter → handler
	//
	// Audit must be outermost so it captures every request, including
	// unauthenticated ones that auth rejects before reaching the handler.
	var handler http.Handler = mux

	// Innermost: rate limiting (only reached after auth succeeds)
	if s.rateLimiter != nil {
		handler = s.rateLimiter.Middleware(handler)
	}

	// Middle: authentication (rejects invalid keys before rate limiting)
	if s.auth != nil {
		handler = s.auth.Middleware(handler)
	}

	// Outermost: audit logging (runs first, sees all requests including 401s)
	if s.audit != nil {
		handler = s.audit.Middleware(handler)
	}

	return handler
}

// Start begins serving HTTP requests
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()

	if s.server != nil {
		s.mu.Unlock()
		return fmt.Errorf("server already started")
	}

	handler := s.GetHandler()

	srv := &http.Server{
		Addr:    s.addr,
		Handler: handler,
	}
	s.server = srv

	log.Printf("🔌 MCP server listening on %s", s.addr)
	if s.auth != nil {
		log.Println("   🔒 Authentication enabled")
	}
	if s.rateLimiter != nil {
		log.Println("   ⚡ Rate limiting enabled")
	}
	if s.audit != nil {
		log.Println("   📝 Audit logging enabled")
	}

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	s.mu.Unlock()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server == nil {
		return nil
	}

	log.Println("🛑 MCP server shutting down...")

	if err := s.server.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.server = nil
	log.Println("✅ MCP server stopped")
	return nil
}

// handleRequest processes incoming JSON-RPC requests
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		s.writeError(w, nil, InvalidRequest, "method not allowed")
		return
	}

	// Parse JSON-RPC request
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, nil, ParseError, "invalid JSON: "+err.Error())
		return
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		s.writeError(w, req.ID, InvalidRequest, "invalid JSON-RPC version")
		return
	}

	// Execute tool
	result, err := s.registry.Execute(r.Context(), req.Method, s.parseParams(req.Params))
	if err != nil {
		s.writeError(w, req.ID, InternalError, err.Error())
		return
	}

	// Write success response
	s.writeResponse(w, req.ID, result)
}

// parseParams converts raw JSON to map[string]interface{}
func (s *Server) parseParams(raw json.RawMessage) map[string]interface{} {
	if raw == nil {
		return make(map[string]interface{})
	}

	var params map[string]interface{}
	if err := json.Unmarshal(raw, &params); err != nil {
		return make(map[string]interface{})
	}

	return params
}

// writeResponse writes a JSON-RPC success response
func (s *Server) writeResponse(w http.ResponseWriter, id interface{}, result interface{}) {
	response := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// writeError writes a JSON-RPC error response
func (s *Server) writeError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := ErrorResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: RPCError{
			Code:    code,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}
