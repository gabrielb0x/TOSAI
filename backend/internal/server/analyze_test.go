package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type openAIJSONFixture struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

func TestCallOpenAIJSONResponseCombinesSplitOutputText(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     "resp_split",
			"status": "completed",
			"output": []map[string]any{
				{
					"type":   "message",
					"status": "completed",
					"content": []map[string]any{
						{"type": "output_text", "text": "{\"name\":\"Google\","},
						{"type": "output_text", "text": "\"items\":[\"privacy\",\"ads\"]}"},
					},
				},
			},
		})
	}))
	defer server.Close()

	service := &analysisService{
		client:   server.Client(),
		apiKey:   "test-key",
		model:    "gpt-test",
		baseURL:  server.URL,
		appEnv:   "test",
		maxChars: 12000,
	}
	service.client.Timeout = 5 * time.Second

	var output openAIJSONFixture
	debugStep, err := service.callOpenAIJSONResponse(
		context.Background(),
		"gpt-test",
		"unit",
		"system",
		"user",
		nil,
		"fixture",
		simpleFixtureSchema(),
		200,
		&output,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output.Name != "Google" {
		t.Fatalf("expected parsed name Google, got %q", output.Name)
	}
	if len(output.Items) != 2 || output.Items[0] != "privacy" || output.Items[1] != "ads" {
		t.Fatalf("unexpected parsed items: %#v", output.Items)
	}

	if debugStep["final_attempt"] != 1 {
		t.Fatalf("expected final_attempt=1, got %#v", debugStep["final_attempt"])
	}
}

func TestCallOpenAIJSONResponseRetriesOnIncompleteJSON(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callNumber := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")

		if callNumber == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "resp_incomplete",
				"status": "incomplete",
				"incomplete_details": map[string]any{
					"reason": "max_output_tokens",
				},
				"output": []map[string]any{
					{
						"type":   "message",
						"status": "incomplete",
						"content": []map[string]any{
							{"type": "output_text", "text": "{\"name\":\"Google\""},
						},
					},
				},
			})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     "resp_complete",
			"status": "completed",
			"output": []map[string]any{
				{
					"type":   "message",
					"status": "completed",
					"content": []map[string]any{
						{"type": "output_text", "text": "{\"name\":\"Google\",\"items\":[\"privacy\"]}"},
					},
				},
			},
		})
	}))
	defer server.Close()

	service := &analysisService{
		client:   server.Client(),
		apiKey:   "test-key",
		model:    "gpt-test",
		baseURL:  server.URL,
		appEnv:   "test",
		maxChars: 12000,
	}
	service.client.Timeout = 5 * time.Second

	var output openAIJSONFixture
	debugStep, err := service.callOpenAIJSONResponse(
		context.Background(),
		"gpt-test",
		"unit",
		"system",
		"user",
		nil,
		"fixture",
		simpleFixtureSchema(),
		100,
		&output,
	)
	if err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}

	if got := calls.Load(); got != 2 {
		t.Fatalf("expected 2 upstream calls, got %d", got)
	}
	if output.Name != "Google" || len(output.Items) != 1 || output.Items[0] != "privacy" {
		t.Fatalf("unexpected parsed output: %#v", output)
	}
	if debugStep["final_attempt"] != 2 {
		t.Fatalf("expected final_attempt=2, got %#v", debugStep["final_attempt"])
	}
}

func simpleFixtureSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
			"items": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"name", "items"},
	}
}
