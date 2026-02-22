package mcp

import (
	"context"
	"crypto/subtle"
	"net/http"
)

// AuthMiddleware validates API keys for MCP requests
type AuthMiddleware struct {
	apiKey []byte
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(apiKey string) *AuthMiddleware {
	return &AuthMiddleware{
		apiKey: []byte(apiKey),
	}
}

// Middleware wraps an http.Handler with API key validation
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health checks
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract API key from Authorization header
		apiKey := extractAPIKey(r)
		if apiKey == "" {
			http.Error(w, `{"jsonrpc":"2.0","id":null,"error":{"code":-32000,"message":"missing authorization"}}`, http.StatusUnauthorized)
			return
		}

		// Constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(apiKey), m.apiKey) != 1 {
			http.Error(w, `{"jsonrpc":"2.0","id":null,"error":{"code":-32001,"message":"invalid api key"}}`, http.StatusForbidden)
			return
		}

		// Add API key to context for audit logging
		ctx := context.WithValue(r.Context(), ContextKeyAPIKey, apiKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractAPIKey extracts the API key from the Authorization header
func extractAPIKey(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	// Support "Bearer <key>" format
	const bearerPrefix = "Bearer "
	if len(auth) > len(bearerPrefix) && auth[:len(bearerPrefix)] == bearerPrefix {
		return auth[len(bearerPrefix):]
	}

	// Also support raw key (for backwards compatibility)
	return auth
}

// ContextKeyAPIKey is the context key for storing API key
type ContextKeyAPIKey struct{}

// GetAPIKeyFromContext retrieves the API key from request context
func GetAPIKeyFromContext(ctx context.Context) string {
	key, ok := ctx.Value(ContextKeyAPIKey).(string)
	if !ok {
		return ""
	}
	return key
}
