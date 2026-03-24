package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	requestIDHeader             = "X-Request-Id"
	requestIDContextKey         = "request_id"
	analysisTargetURLContextKey = "analysis_target_url"
)

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(requestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Set(requestIDContextKey, requestID)
		c.Header(requestIDHeader, requestID)
		c.Next()
	}
}

func newRequestID() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return "req_" + hex.EncodeToString(buf)
}

func getRequestID(c *gin.Context) string {
	if c == nil {
		return newRequestID()
	}

	if value, exists := c.Get(requestIDContextKey); exists {
		if requestID, ok := value.(string); ok && strings.TrimSpace(requestID) != "" {
			return requestID
		}
	}

	requestID := newRequestID()
	c.Set(requestIDContextKey, requestID)
	c.Header(requestIDHeader, requestID)
	return requestID
}

func handleAPIRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"service":      "tosai-backend",
		"version":      "v1",
		"base_path":    "/v1",
		"request_id":   getRequestID(c),
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"endpoints": gin.H{
			"summary_get":  "/v1/summary?url=https://example.com/terms",
			"summary_post": "/v1/summary",
		},
	})
}

func shouldExposeErrorDetails(appEnv string) bool {
	return !strings.EqualFold(strings.TrimSpace(appEnv), "prod")
}

func logAPIError(c *gin.Context, apiErr *apiError) {
	if c == nil || apiErr == nil {
		return
	}

	path := c.FullPath()
	if path == "" && c.Request != nil && c.Request.URL != nil {
		path = c.Request.URL.Path
	}

	targetURL := ""
	if value, exists := c.Get(analysisTargetURLContextKey); exists {
		targetURL = fmt.Sprintf("%v", value)
	}

	details := "{}"
	if len(apiErr.Details) > 0 {
		if encoded, err := json.Marshal(apiErr.Details); err == nil {
			details = string(encoded)
		} else {
			details = fmt.Sprintf("%v", apiErr.Details)
		}
	}

	cause := ""
	if apiErr.Err != nil {
		cause = apiErr.Err.Error()
	}

	method := ""
	clientIP := ""
	if c.Request != nil {
		method = c.Request.Method
		clientIP = c.ClientIP()
	}

	log.Printf(
		"api_error request_id=%s method=%s path=%s client_ip=%s status=%d code=%s target_url=%q message=%q details=%s cause=%q",
		getRequestID(c),
		method,
		path,
		clientIP,
		apiErr.Status,
		apiErr.Code,
		targetURL,
		apiErr.Message,
		details,
		cause,
	)
}
