package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gabrielb0x/TOSAI/backend/internal/config"
	"github.com/gin-gonic/gin"
)

func TestSummaryRouteAvailableAtRoot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := New(config.Config{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/summary", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON body: %v", err)
	}

	if payload["code"] != "missing_url" {
		t.Fatalf("expected code missing_url, got %#v", payload["code"])
	}

	requestID, ok := payload["request_id"].(string)
	if !ok || requestID == "" {
		t.Fatalf("expected request_id in response, got %#v", payload["request_id"])
	}
}

func TestSummaryRouteAvailableAtVersionedPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := New(config.Config{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/summary", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestAPIVersionRootAvailableWithTrailingSlash(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := New(config.Config{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON body: %v", err)
	}

	if payload["version"] != "v1" {
		t.Fatalf("expected version v1, got %#v", payload["version"])
	}
}

func TestSummaryPostRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := New(config.Config{APIDebugMode: true}, nil)
	req := httptest.NewRequest(http.MethodPost, "/v1/summary", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON body: %v", err)
	}

	if payload["code"] != "invalid_json" {
		t.Fatalf("expected code invalid_json, got %#v", payload["code"])
	}

	if _, ok := payload["details"].(map[string]any); !ok {
		t.Fatalf("expected details object, got %#v", payload["details"])
	}
}

func TestSummaryRouteRateLimitedToOneRequestPerMinute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := New(config.Config{
		AnalysisRateLimitPerMin: 1,
		TrustedProxies:          []string{"127.0.0.1"},
	}, nil)

	firstReq := httptest.NewRequest(http.MethodGet, "/v1/summary", nil)
	firstReq.RemoteAddr = "127.0.0.1:1234"
	firstRec := httptest.NewRecorder()
	router.ServeHTTP(firstRec, firstReq)

	if firstRec.Code != http.StatusBadRequest {
		t.Fatalf("expected first status %d, got %d", http.StatusBadRequest, firstRec.Code)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/v1/summary", nil)
	secondReq.RemoteAddr = "127.0.0.1:1234"
	secondRec := httptest.NewRecorder()
	router.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second status %d, got %d", http.StatusTooManyRequests, secondRec.Code)
	}
}

func TestHealthzRouteRemoved(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := New(config.Config{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestRootRouteRemovedInAPIOnlyMode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := New(config.Config{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
