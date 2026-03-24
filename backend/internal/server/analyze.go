package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/gabrielb0x/TOSAI/backend/internal/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

const (
	defaultHTTPTimeoutSec = 25
	maxFetchedBytes       = 650000
	maxListItems          = 5
)

type analysisService struct {
	client   *http.Client
	apiKey   string
	model    string
	baseURL  string
	maxChars int
	appEnv   string
}

type summaryRequest struct {
	URL string `json:"url"`
}

type summaryAnalysis struct {
	Rating         string   `json:"rating"`
	Summary        string   `json:"summary"`
	Highlights     []string `json:"highlights"`
	Risks          []string `json:"risks"`
	Recommendation string   `json:"recommendation"`
	Confidence     string   `json:"confidence"`
}

type fetchedDocument struct {
	SourceURL   string
	HTTPStatus  int
	ContentType string
	Text        string
	Characters  int
}

type apiError struct {
	Status  int
	Code    string
	Message string
	Details map[string]any
	Err     error
}

func (e *apiError) Error() string {
	return e.Message
}

func (e *apiError) WithDetail(key string, value any) *apiError {
	if e == nil || strings.TrimSpace(key) == "" || value == nil {
		return e
	}

	if text, ok := value.(string); ok && strings.TrimSpace(text) == "" {
		return e
	}

	if e.Details == nil {
		e.Details = map[string]any{}
	}
	e.Details[key] = value
	return e
}

func (e *apiError) WithCause(err error) *apiError {
	if e == nil || err == nil {
		return e
	}
	e.Err = err
	return e
}

type openAIResponsesResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func newAnalysisService(cfg config.Config) *analysisService {
	timeout := cfg.HTTPTimeoutSec
	if timeout <= 0 {
		timeout = defaultHTTPTimeoutSec
	}

	maxChars := cfg.AnalysisMaxChar
	if maxChars <= 0 {
		maxChars = 12000
	}

	model := strings.TrimSpace(cfg.OpenAIModel)
	if model == "" {
		model = "gpt-5-nano"
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.OpenAIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &analysisService{
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		apiKey:   strings.TrimSpace(cfg.OpenAIAPIKey),
		model:    model,
		baseURL:  baseURL,
		maxChars: maxChars,
		appEnv:   strings.TrimSpace(cfg.AppEnv),
	}
}

func (s *analysisService) handleSummaryGET(c *gin.Context) {
	s.handleSummary(c, c.Query("url"))
}

func (s *analysisService) handleSummaryPOST(c *gin.Context) {
	var req summaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeAPIError(c, (&apiError{
			Status:  http.StatusBadRequest,
			Code:    "invalid_json",
			Message: "payload JSON invalide (attendu: {\"url\":\"https://...\"})",
		}).WithDetail("reason", err.Error()).WithCause(err), s.appEnv)
		return
	}
	s.handleSummary(c, req.URL)
}

func (s *analysisService) handleSummary(c *gin.Context, rawURL string) {
	target := strings.TrimSpace(rawURL)
	if target == "" {
		writeAPIError(c, &apiError{
			Status:  http.StatusBadRequest,
			Code:    "missing_url",
			Message: "parametre url obligatoire",
		}, s.appEnv)
		return
	}
	c.Set(analysisTargetURLContextKey, target)

	doc, err := s.fetchDocument(c.Request.Context(), target)
	if err != nil {
		writeAPIError(c, err, s.appEnv)
		return
	}

	analysis, err := s.analyzeWithOpenAI(c.Request.Context(), doc)
	if err != nil {
		writeAPIError(c, err, s.appEnv)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"request_id":   getRequestID(c),
		"source_url":   doc.SourceURL,
		"model":        s.model,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"fetched": gin.H{
			"http_status":  doc.HTTPStatus,
			"content_type": doc.ContentType,
			"characters":   doc.Characters,
		},
		"analysis": analysis,
	})
}

func (s *analysisService) fetchDocument(ctx context.Context, rawURL string) (fetchedDocument, error) {
	normalized, err := normalizeURL(rawURL)
	if err != nil {
		return fetchedDocument{}, (&apiError{
			Status:  http.StatusBadRequest,
			Code:    "invalid_url",
			Message: err.Error(),
		}).WithDetail("raw_url", strings.TrimSpace(rawURL)).WithCause(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalized, nil)
	if err != nil {
		return fetchedDocument{}, (&apiError{
			Status:  http.StatusBadRequest,
			Code:    "invalid_url",
			Message: "impossible de construire la requete cible",
		}).WithDetail("url", normalized).WithDetail("reason", err.Error()).WithCause(err)
	}
	req.Header.Set("User-Agent", "TOSAI/0.1 (+https://tosai.local)")
	req.Header.Set("Accept", "text/html,text/plain;q=0.9,*/*;q=0.1")

	resp, err := s.client.Do(req)
	if err != nil {
		return fetchedDocument{}, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "fetch_failed",
			Message: "impossible de recuperer la page cible",
		}).WithDetail("url", normalized).WithDetail("reason", err.Error()).WithCause(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFetchedBytes))
	if err != nil {
		return fetchedDocument{}, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "fetch_read_failed",
			Message: "lecture du contenu cible impossible",
		}).WithDetail("url", normalized).WithDetail("reason", err.Error()).WithCause(err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fetchedDocument{}, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "upstream_http_error",
			Message: fmt.Sprintf("le site distant a repondu avec HTTP %d", resp.StatusCode),
		}).WithDetail("url", normalized).
			WithDetail("upstream_status", resp.StatusCode).
			WithDetail("response_excerpt", excerptText(string(body), 220))
	}

	text := extractPlainText(resp.Header.Get("Content-Type"), body)
	if text == "" {
		return fetchedDocument{}, (&apiError{
			Status:  http.StatusUnprocessableEntity,
			Code:    "empty_content",
			Message: "contenu inutilisable apres extraction de texte",
		}).WithDetail("url", normalized).WithDetail("content_type", resp.Header.Get("Content-Type"))
	}

	text = truncateRunes(text, s.maxChars)
	return fetchedDocument{
		SourceURL:   normalized,
		HTTPStatus:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Text:        text,
		Characters:  len([]rune(text)),
	}, nil
}

func (s *analysisService) analyzeWithOpenAI(ctx context.Context, doc fetchedDocument) (summaryAnalysis, error) {
	if s.apiKey == "" {
		return summaryAnalysis{}, &apiError{
			Status:  http.StatusServiceUnavailable,
			Code:    "missing_openai_api_key",
			Message: "OPENAI_API_KEY manquant cote backend",
		}
	}

	systemPrompt := "Tu es un auditeur juridique produit. Tu analyses des CGU/ToS et tu renvoies uniquement un JSON strict conforme au schema."
	userPrompt := fmt.Sprintf(
		"Analyse la page suivante en francais. URL: %s\n\nContenu extrait:\n%s\n\nFournis une note globale A-E, un resume clair, les points majeurs, les risques et une recommandation utilisateur.",
		doc.SourceURL,
		doc.Text,
	)

	payload := map[string]any{
		"model": s.model,
		"input": []map[string]any{
			{
				"role": "system",
				"content": []map[string]string{
					{"type": "input_text", "text": systemPrompt},
				},
			},
			{
				"role": "user",
				"content": []map[string]string{
					{"type": "input_text", "text": userPrompt},
				},
			},
		},
		"max_output_tokens": 900,
		"text": map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   "tosai_analysis",
				"strict": true,
				"schema": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"rating": map[string]any{
							"type": "string",
							"enum": []string{"A", "B", "C", "D", "E"},
						},
						"summary": map[string]any{
							"type":      "string",
							"minLength": 20,
						},
						"highlights": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"risks": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"recommendation": map[string]any{
							"type": "string",
						},
						"confidence": map[string]any{
							"type": "string",
							"enum": []string{"low", "medium", "high"},
						},
					},
					"required": []string{"rating", "summary", "highlights", "risks", "recommendation", "confidence"},
				},
			},
		},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return summaryAnalysis{}, (&apiError{
			Status:  http.StatusInternalServerError,
			Code:    "openai_payload_error",
			Message: "impossible de construire la requete OpenAI",
		}).WithDetail("reason", err.Error()).WithCause(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/responses", bytes.NewReader(requestBody))
	if err != nil {
		return summaryAnalysis{}, (&apiError{
			Status:  http.StatusInternalServerError,
			Code:    "openai_request_error",
			Message: "impossible de creer la requete OpenAI",
		}).WithDetail("reason", err.Error()).WithCause(err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return summaryAnalysis{}, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_unreachable",
			Message: "impossible de contacter OpenAI",
		}).WithDetail("base_url", s.baseURL).WithDetail("reason", err.Error()).WithCause(err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, maxFetchedBytes))
	if err != nil {
		return summaryAnalysis{}, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_read_error",
			Message: "lecture de la reponse OpenAI impossible",
		}).WithDetail("base_url", s.baseURL).WithDetail("reason", err.Error()).WithCause(err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return summaryAnalysis{}, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_http_error",
			Message: fmt.Sprintf("OpenAI a repondu HTTP %d: %s", resp.StatusCode, extractOpenAIErrorMessage(responseBody)),
		}).WithDetail("upstream_status", resp.StatusCode).
			WithDetail("response_excerpt", excerptText(string(responseBody), 220))
	}

	var parsed openAIResponsesResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return summaryAnalysis{}, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_parse_error",
			Message: "reponse OpenAI invalide",
		}).WithDetail("reason", err.Error()).
			WithDetail("response_excerpt", excerptText(string(responseBody), 220)).
			WithCause(err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return summaryAnalysis{}, &apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_error",
			Message: parsed.Error.Message,
		}
	}

	rawJSON := strings.TrimSpace(parsed.OutputText)
	if rawJSON == "" {
		rawJSON = strings.TrimSpace(extractTextFromOutput(parsed.Output))
	}
	if rawJSON == "" {
		return summaryAnalysis{}, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_empty_output",
			Message: "OpenAI n'a renvoye aucun contenu exploitable",
		}).WithDetail("response_excerpt", excerptText(string(responseBody), 220))
	}

	var analysis summaryAnalysis
	if err := json.Unmarshal([]byte(rawJSON), &analysis); err != nil {
		return summaryAnalysis{}, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_output_invalid",
			Message: "OpenAI n'a pas renvoye le JSON attendu",
		}).WithDetail("reason", err.Error()).
			WithDetail("response_excerpt", excerptText(rawJSON, 220)).
			WithCause(err)
	}

	analysis = sanitizeAnalysis(analysis)
	return analysis, nil
}

func writeAPIError(c *gin.Context, err error, appEnv string) {
	if err == nil {
		apiErr := &apiError{
			Status:  http.StatusInternalServerError,
			Code:    "unknown",
			Message: "erreur inconnue",
		}
		logAPIError(c, apiErr)
		c.JSON(apiErr.Status, gin.H{
			"status":     "error",
			"code":       apiErr.Code,
			"message":    apiErr.Message,
			"request_id": getRequestID(c),
		})
		return
	}

	apiErr := &apiError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: "erreur interne du serveur",
		Err:     err,
	}
	if typed, ok := err.(*apiError); ok {
		apiErr = typed
	}

	logAPIError(c, apiErr)

	payload := gin.H{
		"status":     "error",
		"code":       apiErr.Code,
		"message":    apiErr.Message,
		"request_id": getRequestID(c),
	}
	if shouldExposeErrorDetails(appEnv) && len(apiErr.Details) > 0 {
		payload["details"] = apiErr.Details
	}

	c.JSON(apiErr.Status, payload)
}

func sanitizeAnalysis(input summaryAnalysis) summaryAnalysis {
	input.Rating = strings.ToUpper(strings.TrimSpace(input.Rating))
	switch input.Rating {
	case "A", "B", "C", "D", "E":
	default:
		input.Rating = "C"
	}
	input.Summary = strings.TrimSpace(input.Summary)
	if input.Summary == "" {
		input.Summary = "Resume indisponible."
	}
	input.Recommendation = strings.TrimSpace(input.Recommendation)
	if input.Recommendation == "" {
		input.Recommendation = "Lire attentivement les clauses sensibles avant de continuer."
	}
	input.Confidence = strings.ToLower(strings.TrimSpace(input.Confidence))
	if input.Confidence == "" {
		input.Confidence = "medium"
	}
	input.Highlights = sanitizeList(input.Highlights)
	input.Risks = sanitizeList(input.Risks)
	return input
}

func sanitizeList(items []string) []string {
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		text := strings.TrimSpace(item)
		if text == "" {
			continue
		}
		cleaned = append(cleaned, text)
		if len(cleaned) >= maxListItems {
			break
		}
	}
	if len(cleaned) == 0 {
		return []string{"Aucun element pertinent detecte."}
	}
	return cleaned
}

func normalizeURL(raw string) (string, error) {
	parsed, err := neturl.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("URL invalide")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("URL invalide: seul http/https est accepte")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("URL invalide: domaine manquant")
	}
	return parsed.String(), nil
}

func extractPlainText(contentType string, body []byte) string {
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return ""
	}

	lowerType := strings.ToLower(contentType)
	if strings.Contains(lowerType, "html") {
		extracted := strings.TrimSpace(extractTextFromHTML(raw))
		if extracted != "" {
			return extracted
		}
	}
	return normalizeWhitespace(raw)
}

func extractTextFromHTML(content string) string {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return ""
	}

	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "script", "style", "noscript", "svg":
				return
			}
		}

		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)
			if text != "" {
				builder.WriteString(text)
				builder.WriteByte('\n')
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(doc)
	return normalizeWhitespace(builder.String())
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func truncateRunes(value string, max int) string {
	if max <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func extractTextFromOutput(output []struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}) string {
	for _, item := range output {
		for _, content := range item.Content {
			if content.Type == "output_text" || content.Type == "text" {
				if strings.TrimSpace(content.Text) != "" {
					return content.Text
				}
			}
		}
	}
	return ""
}

func extractOpenAIErrorMessage(body []byte) string {
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err == nil {
		if payload, ok := parsed["error"].(map[string]any); ok {
			if message, ok := payload["message"].(string); ok && strings.TrimSpace(message) != "" {
				return message
			}
		}
	}

	message := strings.TrimSpace(string(body))
	if message == "" {
		return "erreur inconnue"
	}
	return message
}

func excerptText(value string, max int) string {
	cleaned := normalizeWhitespace(strings.TrimSpace(value))
	if cleaned == "" {
		return ""
	}
	if max <= 0 {
		return cleaned
	}

	runes := []rune(cleaned)
	if len(runes) <= max {
		return cleaned
	}
	return string(runes[:max]) + "..."
}
