package mcp

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiterMiddleware enforces rate limits per API key
type RateLimiterMiddleware struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	limit    rate.Limit
	burst    int
}

// NewRateLimiterMiddleware creates a new rate limiter
func NewRateLimiterMiddleware(requestsPerMinute int, burst int) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Limit(requestsPerMinute) / 60.0, // Convert to per-second
		burst:    burst,
	}
}

// Middleware wraps an http.Handler with rate limiting
func (m *RateLimiterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from context (set by auth middleware)
		apiKey := GetAPIKeyFromContext(r.Context())
		if apiKey == "" {
			// No API key, use default limiter (shouldn't happen if auth is enabled)
			apiKey = "anonymous"
		}

		// Get or create limiter for this API key
		limiter := m.getLimiter(apiKey)

		// Check if request is allowed
		if !limiter.Allow() {
			http.Error(w, `{"jsonrpc":"2.0","id":null,"error":{"code":-32002,"message":"rate limit exceeded"}}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getLimiter returns or creates a rate limiter for an API key
func (m *RateLimiterMiddleware) getLimiter(apiKey string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, exists := m.limiters[apiKey]
	if !exists {
		limiter = rate.NewLimiter(m.limit, m.burst)
		m.limiters[apiKey] = limiter
	}

	return limiter
}

// Cleanup removes inactive limiters to prevent memory leaks
func (m *RateLimiterMiddleware) Cleanup(inactiveDuration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// TODO: Track last access time per limiter
	// For now, this is a placeholder for future implementation
	_ = inactiveDuration
}
