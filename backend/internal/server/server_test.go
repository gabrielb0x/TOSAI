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
}

func TestSummaryRouteCompatibilityAliasStillWorks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := New(config.Config{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/summary", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSummaryPostRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := New(config.Config{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/summary", strings.NewReader("{"))
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
