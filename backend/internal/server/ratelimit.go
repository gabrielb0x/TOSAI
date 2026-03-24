package server

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]rateLimitEntry
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	if limit <= 0 || window <= 0 {
		return nil
	}

	return &rateLimiter{
		limit:   limit,
		window:  window,
		entries: map[string]rateLimitEntry{},
	}
}

func (l *rateLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	if l == nil {
		return true, 0
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for existingKey, entry := range l.entries {
		if now.After(entry.resetAt) {
			delete(l.entries, existingKey)
		}
	}

	entry, exists := l.entries[key]
	if !exists || now.After(entry.resetAt) {
		l.entries[key] = rateLimitEntry{
			count:   1,
			resetAt: now.Add(l.window),
		}
		return true, 0
	}

	if entry.count >= l.limit {
		return false, entry.resetAt.Sub(now)
	}

	entry.count++
	l.entries[key] = entry
	return true, 0
}

func summaryRateLimitMiddleware(limiter *rateLimiter, appEnv string, debugMode bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		clientIP := strings.TrimSpace(c.ClientIP())
		if clientIP == "" {
			clientIP = "unknown"
		}

		allowed, retryAfter := limiter.allow(clientIP, time.Now().UTC())
		if allowed {
			c.Next()
			return
		}

		retryAfterSeconds := int(retryAfter.Seconds())
		if retryAfterSeconds < 1 {
			retryAfterSeconds = 1
		}

		c.Header("Retry-After", strconv.Itoa(retryAfterSeconds))
		writeAPIError(c, (&apiError{
			Status:  http.StatusTooManyRequests,
			Code:    "rate_limited",
			Message: "une seule demande par minute est autorisee par client",
		}).WithDetail("retry_after_seconds", retryAfterSeconds).
			WithDetail("limit_per_minute", limiter.limit), appEnv, debugMode)
		c.Abort()
	}
}
