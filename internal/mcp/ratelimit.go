package mcp

import (
	"container/list"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// limiterEntry holds the rate limiter and metadata for the LRU cache
type limiterEntry struct {
	limiter  *rate.Limiter
	apiKey   string
	lastSeen time.Time
}

// RateLimiterMiddleware enforces rate limits per API key using an LRU strategy
// to prevent memory leaks from inactive API keys.
type RateLimiterMiddleware struct {
	mu       sync.Mutex
	limiters map[string]*list.Element
	lruList  *list.List
	limit    rate.Limit
	burst    int

	// Configuration for cleanup
	maxSize          int
	inactiveDuration time.Duration
}

// NewRateLimiterMiddleware creates a new rate limiter
func NewRateLimiterMiddleware(requestsPerMinute int, burst int) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiters:         make(map[string]*list.Element),
		lruList:          list.New(),
		limit:            rate.Limit(requestsPerMinute) / 60.0, // Convert to per-second
		burst:            burst,
		maxSize:          10000,           // Maximum number of API keys to track
		inactiveDuration: 3 * time.Minute, // Time before an inactive bucket is cleared
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

// getLimiter returns or creates a rate limiter for an API key, managing the LRU cache
func (m *RateLimiterMiddleware) getLimiter(apiKey string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// 1. Passive Tail Eviction: Check if the oldest entry is stale
	for m.lruList.Len() > 0 {
		oldest := m.lruList.Back()
		entry := oldest.Value.(*limiterEntry)

		if now.Sub(entry.lastSeen) > m.inactiveDuration {
			// The oldest bucket is stale, evict it
			m.lruList.Remove(oldest)
			delete(m.limiters, entry.apiKey)
		} else {
			// Since we process oldest to newest, if this one isn't stale,
			// the rest aren't either. Stop checking.
			break
		}
	}

	// 2. Fetch or Create
	if element, exists := m.limiters[apiKey]; exists {
		// Existing user: mark as recently used by moving to front
		m.lruList.MoveToFront(element)
		entry := element.Value.(*limiterEntry)
		entry.lastSeen = now
		return entry.limiter
	}

	// 3. New user mapping
	// Before adding, ensure we haven't hit the absolute maximum threshold
	if len(m.limiters) >= m.maxSize {
		// Force evict the oldest entry to make room
		oldest := m.lruList.Back()
		if oldest != nil {
			entry := oldest.Value.(*limiterEntry)
			m.lruList.Remove(oldest)
			delete(m.limiters, entry.apiKey)
		}
	}

	// Add the new user to the front
	limiter := rate.NewLimiter(m.limit, m.burst)
	entry := &limiterEntry{
		limiter:  limiter,
		apiKey:   apiKey,
		lastSeen: now,
	}
	element := m.lruList.PushFront(entry)
	m.limiters[apiKey] = element

	return limiter
}

// Cleanup removes inactive limiters to prevent memory leaks.
// Note: This is maintained for package compatibility, but is largely unused
// now as getLimiter performs constant time (O(1)) tail eviction synchronously.
func (m *RateLimiterMiddleware) Cleanup(inactiveDuration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Optional aggressive full-scan (usually unnecessary due to tail eviction)
	now := time.Now()
	for element := m.lruList.Front(); element != nil; {
		next := element.Next()
		entry := element.Value.(*limiterEntry)

		if now.Sub(entry.lastSeen) > inactiveDuration {
			m.lruList.Remove(element)
			delete(m.limiters, entry.apiKey)
		}

		element = next
	}
}
