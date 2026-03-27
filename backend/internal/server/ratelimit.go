package server

import (
	"bytes"
	"encoding/json"
	"io"
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

func summaryRateLimitMiddleware(limiter *rateLimiter, cacheMaxAge time.Duration, appEnv string, debugMode bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		if shouldBypassRateLimitForCachedSummary(c, cacheMaxAge) {
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

func shouldBypassRateLimitForCachedSummary(c *gin.Context, cacheMaxAge time.Duration) bool {
	if c == nil || c.Request == nil {
		return false
	}

	targetURL, forceRefresh, ok := extractSummaryRateLimitRequest(c)
	if !ok {
		return false
	}

	normalizedURL, err := normalizeURL(targetURL)
	if err != nil {
		return false
	}

	cached, err := loadCachedAnalysis(c.Request.Context(), getDBPool(c), normalizedURL)
	if err != nil || cached == nil {
		return false
	}

	return shouldServeCachedSummary(forceRefresh, cached, cacheMaxAge, time.Now().UTC())
}

func shouldServeCachedSummary(forceRefresh bool, cached *cachedAnalysisRecord, cacheMaxAge time.Duration, now time.Time) bool {
	if cached == nil {
		return false
	}

	isStale := isCachedAnalysisStale(cached.AnalyzedAt, cacheMaxAge, now)
	canRefresh := cached.IsContestable || isStale
	return !forceRefresh || !canRefresh
}

func isCachedAnalysisStale(analyzedAt time.Time, cacheMaxAge time.Duration, now time.Time) bool {
	if analyzedAt.IsZero() {
		return true
	}
	if cacheMaxAge <= 0 {
		return true
	}
	return now.Sub(analyzedAt) > cacheMaxAge
}

func extractSummaryRateLimitRequest(c *gin.Context) (string, bool, bool) {
	if c == nil || c.Request == nil {
		return "", false, false
	}

	switch c.Request.Method {
	case http.MethodGet:
		targetURL := strings.TrimSpace(c.Query("url"))
		if targetURL == "" {
			return "", false, false
		}
		return targetURL, isTruthy(c.Query("force_refresh")), true
	case http.MethodPost:
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return "", false, false
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		var req summaryRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return "", false, false
		}

		targetURL := strings.TrimSpace(req.URL)
		if targetURL == "" {
			return "", false, false
		}
		return targetURL, req.ForceRefresh, true
	default:
		return "", false, false
	}
}
