package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/gabrielb0x/TOSAI/backend/internal/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

const (
	defaultHTTPTimeoutSec   = 45
	maxFetchedBytes         = 650000
	maxListItems            = 5
	openAIMaxJSONAttempts   = 2
	openAIRetryTokenBump    = 600
	openAIMaxOutputTokenCap = 2400
)

type analysisService struct {
	client        *http.Client
	apiKey        string
	model         string
	researchModel string
	baseURL       string
	maxChars      int
	appEnv        string
	cacheMaxAge   time.Duration
	debugMode     bool
}

type summaryRequest struct {
	URL          string `json:"url"`
	ForceRefresh bool   `json:"force_refresh"`
}

type summaryAnalysis struct {
	ServiceName    string          `json:"service_name"`
	Rating         string          `json:"rating"`
	Summary        string          `json:"summary"`
	Points         []analysisPoint `json:"points"`
	Highlights     []string        `json:"highlights"`
	Risks          []string        `json:"risks"`
	Recommendation string          `json:"recommendation"`
	Confidence     string          `json:"confidence"`
	IsContestable  bool            `json:"is_contestable"`
}

type analysisPoint struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Details  string `json:"details"`
}

type researchResult struct {
	TargetURL     string           `json:"target_url"`
	TargetTitle   string           `json:"target_title"`
	TargetContent string           `json:"target_content"`
	Summary       string           `json:"summary"`
	KeyFindings   []string         `json:"key_findings"`
	Confidence    string           `json:"confidence"`
	Sources       []researchSource `json:"sources"`
}

type researchSource struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Domain string `json:"domain"`
}

type fetchedDocument struct {
	SourceURL   string
	HTTPStatus  int
	ContentType string
	Text        string
	Characters  int
	AcquiredVia string
}

type apiError struct {
	Status  int
	Code    string
	Message string
	Details map[string]any
	Err     error
	Debug   map[string]any
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

func (e *apiError) WithDebug(debug map[string]any) *apiError {
	if e == nil || len(debug) == 0 {
		return e
	}
	e.Debug = debug
	return e
}

type openAIResponsesResponse struct {
	ID                string                   `json:"id"`
	Status            string                   `json:"status"`
	OutputText        string                   `json:"output_text"`
	Output            []openAIResponsesOutput  `json:"output"`
	IncompleteDetails *openAIIncompleteDetails `json:"incomplete_details"`
	Error             *struct {
		Message string `json:"message"`
	} `json:"error"`
	Usage map[string]any `json:"usage"`
}

type openAIResponsesOutput struct {
	Type    string                `json:"type"`
	Status  string                `json:"status"`
	Content []openAIResponsesText `json:"content"`
}

type openAIResponsesText struct {
	Type    string `json:"type"`
	Text    string `json:"text"`
	Refusal string `json:"refusal"`
}

type openAIIncompleteDetails struct {
	Reason string `json:"reason"`
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
		model = "gpt-5.4-mini"
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.OpenAIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	cacheMaxAgeDays := cfg.AnalysisCacheMaxAgeDays
	if cacheMaxAgeDays <= 0 {
		cacheMaxAgeDays = 90
	}

	return &analysisService{
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		apiKey:        strings.TrimSpace(cfg.OpenAIAPIKey),
		model:         model,
		researchModel: strings.TrimSpace(cfg.OpenAIResearchModel),
		baseURL:       baseURL,
		maxChars:      maxChars,
		appEnv:        strings.TrimSpace(cfg.AppEnv),
		cacheMaxAge:   time.Duration(cacheMaxAgeDays) * 24 * time.Hour,
		debugMode:     cfg.APIDebugMode,
	}
}

func (s *analysisService) handleSummaryGET(c *gin.Context) {
	s.handleSummary(c, c.Query("url"), isTruthy(c.Query("force_refresh")))
}

func (s *analysisService) handleSummaryPOST(c *gin.Context) {
	var req summaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeAPIError(c, (&apiError{
			Status:  http.StatusBadRequest,
			Code:    "invalid_json",
			Message: "payload JSON invalide (attendu: {\"url\":\"example.com\"})",
		}).WithDetail("reason", err.Error()).WithCause(err), s.appEnv, s.debugMode)
		return
	}
	s.handleSummary(c, req.URL, req.ForceRefresh)
}

func (s *analysisService) handleSummary(c *gin.Context, rawURL string, forceRefresh bool) {
	target := strings.TrimSpace(rawURL)
	debugPayload := map[string]any{
		"request": map[string]any{
			"force_refresh":      forceRefresh,
			"api_debug_mode":     s.debugMode,
			"cache_max_age_days": int(s.cacheMaxAge.Hours() / 24),
		},
	}

	if target == "" {
		writeAPIError(c, (&apiError{
			Status:  http.StatusBadRequest,
			Code:    "missing_url",
			Message: "parametre url obligatoire",
		}).WithDebug(debugPayload), s.appEnv, s.debugMode)
		return
	}
	c.Set(analysisTargetURLContextKey, target)

	normalizedURL, err := normalizeURL(target)
	if err != nil {
		writeAPIError(c, (&apiError{
			Status:  http.StatusBadRequest,
			Code:    "invalid_url",
			Message: err.Error(),
		}).WithDetail("raw_url", target).WithCause(err).WithDebug(debugPayload), s.appEnv, s.debugMode)
		return
	}
	c.Set(analysisTargetURLContextKey, normalizedURL)

	pool := getDBPool(c)
	cached, err := loadCachedAnalysis(c.Request.Context(), pool, normalizedURL)
	if err != nil {
		log.Printf("avertissement: lecture cache %s impossible: %v", normalizedURL, err)
		debugPayload["cache"] = map[string]any{
			"lookup_error": err.Error(),
		}
	} else if cached != nil {
		isStale := s.isCacheStale(cached.AnalyzedAt)
		canRefresh := cached.IsContestable || isStale
		debugPayload["cache"] = map[string]any{
			"hit":               true,
			"stale":             isStale,
			"analysis_created":  cached.AnalyzedAt.UTC().Format(time.RFC3339),
			"force_refresh":     forceRefresh,
			"refresh_allowed":   canRefresh,
			"stored_model":      cached.Model,
			"stored_debug_size": len(cached.Debug),
		}

		if !forceRefresh || !canRefresh {
			cachedAnalysis := cached.Analysis
			cachedAnalysis.IsContestable = canRefresh

			response := s.buildSuccessResponse(
				c,
				cached.SourceURL,
				cached.Model,
				fetchedDocument{
					SourceURL:   cached.SourceURL,
					HTTPStatus:  cached.HTTPStatus,
					ContentType: cached.ContentType,
					Text:        cached.RawText,
					Characters:  cached.CharacterCount,
					AcquiredVia: "cache",
				},
				cachedAnalysis,
				map[string]any{
					"cached":                true,
					"cache_stale":           isStale,
					"analysis_created_at":   cached.AnalyzedAt.UTC().Format(time.RFC3339),
					"refresh_allowed":       canRefresh,
					"force_refresh_applied": false,
				},
				debugPayload,
			)
			c.JSON(http.StatusOK, response)
			return
		}
	}

	research, researchDebug, err := s.researchWithOpenAI(c.Request.Context(), normalizedURL)
	if len(researchDebug) > 0 {
		debugPayload["web_search"] = researchDebug
		debugPayload["research"] = researchDebug
	}
	if err != nil {
		writeAPIError(c, attachDebugPayload(err, debugPayload), s.appEnv, s.debugMode)
		return
	}

	doc, err := s.resolveTermsDocument(c.Request.Context(), normalizedURL, research)
	if err != nil {
		writeAPIError(c, attachDebugPayload(err, debugPayload), s.appEnv, s.debugMode)
		return
	}

	debugPayload["document"] = map[string]any{
		"source_url":          doc.SourceURL,
		"http_status":         doc.HTTPStatus,
		"content_type":        doc.ContentType,
		"characters":          doc.Characters,
		"acquired_via":        doc.AcquiredVia,
		"text_excerpt":        excerptText(doc.Text, 1800),
		"force_refresh":       forceRefresh,
		"cache_enabled":       pool != nil,
		"requested_domain":    normalizedURL,
		"search_target_url":   research.TargetURL,
		"search_confidence":   research.Confidence,
		"search_page_title":   research.TargetTitle,
		"search_page_excerpt": excerptText(research.TargetContent, 900),
	}

	analysis, analysisDebug, err := s.analyzeWithOpenAI(c.Request.Context(), normalizedURL, doc, research)
	if len(analysisDebug) > 0 {
		debugPayload["chatgpt"] = analysisDebug
		debugPayload["analysis"] = analysisDebug
	}
	if err != nil {
		writeAPIError(c, attachDebugPayload(err, debugPayload), s.appEnv, s.debugMode)
		return
	}

	if pool != nil {
		if err := saveCachedAnalysis(c.Request.Context(), pool, normalizedURL, doc, research, analysis, s.model, debugPayload); err != nil {
			log.Printf("avertissement: ecriture cache %s impossible: %v", doc.SourceURL, err)
			debugPayload["cache_write"] = map[string]any{
				"ok":    false,
				"error": err.Error(),
			}
		} else {
			debugPayload["cache_write"] = map[string]any{
				"ok": true,
			}
		}
	}

	c.JSON(http.StatusOK, s.buildSuccessResponse(
		c,
		doc.SourceURL,
		s.model,
		doc,
		analysis,
		map[string]any{
			"cached":                false,
			"cache_stale":           false,
			"analysis_created_at":   time.Now().UTC().Format(time.RFC3339),
			"refresh_allowed":       analysis.IsContestable,
			"force_refresh_applied": forceRefresh,
		},
		debugPayload,
	))
}

func (s *analysisService) fetchDocument(ctx context.Context, rawURL string) (fetchedDocument, error) {
	normalized, err := normalizeDocumentURL(rawURL)
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
		AcquiredVia: "http_fetch",
	}, nil
}

func (s *analysisService) resolveTermsDocument(ctx context.Context, normalizedURL string, research researchResult) (fetchedDocument, error) {
	targetURL := strings.TrimSpace(research.TargetURL)
	if targetURL != "" {
		doc, err := s.fetchDocument(ctx, targetURL)
		if err == nil {
			return doc, nil
		}
		if strings.TrimSpace(research.TargetContent) == "" {
			return fetchedDocument{}, attachDetail(err, "search_target_url", targetURL)
		}
		log.Printf("avertissement: fallback contenu web search pour %s apres echec fetch %s: %v", normalizedURL, targetURL, err)
	}

	content := strings.TrimSpace(research.TargetContent)
	if content == "" {
		return fetchedDocument{}, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "terms_page_not_found",
			Message: "impossible d'identifier une page de CGU exploitable",
		}).WithDetail("requested_domain", normalizedURL)
	}

	sourceURL := targetURL
	if sourceURL == "" {
		sourceURL = normalizedURL
	}

	content = truncateRunes(content, s.maxChars)
	return fetchedDocument{
		SourceURL:   sourceURL,
		HTTPStatus:  http.StatusOK,
		ContentType: "text/plain; source=openai_web_search",
		Text:        content,
		Characters:  len([]rune(content)),
		AcquiredVia: "openai_web_search_excerpt",
	}, nil
}

func (s *analysisService) researchWithOpenAI(ctx context.Context, normalizedURL string) (researchResult, map[string]any, error) {
	if s.apiKey == "" {
		return researchResult{}, nil, &apiError{
			Status:  http.StatusServiceUnavailable,
			Code:    "missing_openai_api_key",
			Message: "OPENAI_API_KEY manquant cote backend",
		}
	}

	systemPrompt := "Tu travailles en coulisses pour le projet TOSAI. Tu ne dois jamais mentionner ce nom, ni dans le JSON ni dans tes formulations. Tu utilises obligatoirement la recherche web pour identifier la page officielle de CGU/ToS/Terms la plus pertinente pour un domaine, puis en extraire un contenu concret et fiable."
	userPrompt := fmt.Sprintf(
		"Domaine a analyser: %s\n\nUtilise la recherche web pour trouver la meilleure page officielle parmi Terms of Service, Conditions d'utilisation, CGU, Terms, User Agreement, Legal Terms ou equivalent. Priorise la page officielle du service, sur le domaine principal ou un sous-domaine officiel. Retourne uniquement un JSON strict avec l'URL cible, le titre de page, un extrait (ou tout le contenu si possible) concret du contenu de la page retrouvee, un niveau de confiance et les sources. Ignore les elements de reputation generale qui n'aident pas a confirmer ou comprendre la page de CGU.",
		normalizedURL,
	)

	researchSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"target_url": map[string]any{
				"type":      "string",
				"minLength": 8,
			},
			"target_title": map[string]any{
				"type":      "string",
				"minLength": 3,
			},
			"target_content": map[string]any{
				"type":      "string",
				"minLength": 40,
			},
			"summary": map[string]any{
				"type":      "string",
				"minLength": 20,
			},
			"key_findings": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"confidence": map[string]any{
				"type": "string",
				"enum": []string{"low", "medium", "high"},
			},
			"sources": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"title":  map[string]any{"type": "string"},
						"url":    map[string]any{"type": "string"},
						"domain": map[string]any{"type": "string"},
					},
					"required": []string{"title", "url", "domain"},
				},
			},
		},
		"required": []string{"target_url", "target_title", "target_content", "summary", "key_findings", "confidence", "sources"},
	}

	var research researchResult
	debugStep, err := s.callOpenAIJSONResponse(
		ctx,
		s.researchModel,
		"research",
		systemPrompt,
		userPrompt,
		[]map[string]any{{"type": "web_search"}},
		"tosai_research",
		researchSchema,
		1400,
		&research,
	)
	if err != nil {
		return researchResult{}, debugStep, err
	}

	research = sanitizeResearch(research)
	debugStep["parsed_output"] = research
	return research, debugStep, nil
}

func (s *analysisService) analyzeWithOpenAI(ctx context.Context, normalizedURL string, doc fetchedDocument, research researchResult) (summaryAnalysis, map[string]any, error) {
	if s.apiKey == "" {
		return summaryAnalysis{}, nil, &apiError{
			Status:  http.StatusServiceUnavailable,
			Code:    "missing_openai_api_key",
			Message: "OPENAI_API_KEY manquant cote backend",
		}
	}

	systemPrompt := "Tu travailles en coulisses pour le projet TOSAI. Tu ne dois jamais mentionner ce nom. Tu es un auditeur juridique produit specialise dans les CGU, politiques de confidentialite et conditions de service. Tu renvoies uniquement un JSON strict conforme au schema, sans phrase meta. Ton style doit etre tres concis, direct et atomique, proche d'une fiche ToS;DR: un grade global et une liste de points courts et clairs."
	userPrompt := fmt.Sprintf(
		"Analyse en francais la page suivante pour un utilisateur final.\nDomaine demande: %s\nURL de la page retenue: %s\nMethode de collecte: %s\n\nContenu principal a analyser:\n%s\n\nContexte issu de la recherche web:\n%s\n\nFournis:\n- service_name: nom court du service\n- rating: note globale A-E\n- summary: 1 a 2 phrases maximum, tres courtes\n- points: 8 a 24 points atomiques si le texte le permet, sinon moins\n- recommendation: une phrase courte\n- confidence: low/medium/high\n- is_contestable\n\nRegles pour les points:\n- Chaque point doit etre une affirmation courte, autonome et concrete.\n- category doit etre l'un de blocker, bad, neutral, good.\n- Utilise blocker pour les clauses les plus problematiques ou intrusives.\n- Utilise bad pour les points clairement defavorables.\n- Utilise neutral pour les clauses standard ou contextuelles.\n- Utilise good pour les protections ou engagements favorables.\n- title doit rester bref, idealement une seule phrase courte.\n- details doit etre tres court aussi, et ne sert qu'a justifier le point en une phrase.\n- N'ajoute pas de remplissage ni de repetition.\n- Base-toi uniquement sur des elements raisonnablement soutenus par le texte collecte ou le contexte web.\n\nRegles de confiance:\n- Ne mentionne jamais TOSAI.\n- is_contestable doit etre false par defaut.\n- Mets is_contestable a true uniquement si tu n'es pas suffisamment certain de la veracite ou de la fiabilite de ton analyse: page possiblement non officielle, contenu trop incomplet, contradictions majeures, ou confidence=low.\n- Si tu es raisonnablement sur de toi, is_contestable doit etre false.",
		normalizedURL,
		doc.SourceURL,
		doc.AcquiredVia,
		doc.Text,
		researchToPromptText(research),
	)

	analysisSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"service_name": map[string]any{
				"type":      "string",
				"minLength": 2,
			},
			"rating": map[string]any{
				"type": "string",
				"enum": []string{"A", "B", "C", "D", "E"},
			},
			"summary": map[string]any{
				"type":      "string",
				"minLength": 8,
			},
			"points": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"category": map[string]any{
							"type": "string",
							"enum": []string{"blocker", "bad", "neutral", "good"},
						},
						"title": map[string]any{
							"type":      "string",
							"minLength": 5,
						},
						"details": map[string]any{
							"type":      "string",
							"minLength": 3,
						},
					},
					"required": []string{"category", "title", "details"},
				},
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
			"is_contestable": map[string]any{
				"type": "boolean",
			},
		},
		"required": []string{"service_name", "rating", "summary", "points", "recommendation", "confidence", "is_contestable"},
	}

	var analysis summaryAnalysis
	debugStep, err := s.callOpenAIJSONResponse(
		ctx,
		s.model,
		"analysis",
		systemPrompt,
		userPrompt,
		nil,
		"tosai_analysis",
		analysisSchema,
		1600,
		&analysis,
	)
	if err != nil {
		return summaryAnalysis{}, debugStep, err
	}

	analysis = sanitizeAnalysis(analysis)
	debugStep["parsed_output"] = analysis
	return analysis, debugStep, nil
}

func (s *analysisService) callOpenAIJSONResponse(ctx context.Context, model, purpose, systemPrompt, userPrompt string, tools []map[string]any, schemaName string, schema map[string]any, maxOutputTokens int, output any) (map[string]any, error) {
	if model == "" {
		model = s.model
	}

	debugStep := map[string]any{
		"purpose":             purpose,
		"model":               model,
		"system_prompt":       systemPrompt,
		"user_prompt_excerpt": excerptText(userPrompt, 4000),
	}

	attempts := make([]map[string]any, 0, openAIMaxJSONAttempts)
	currentMaxTokens := maxOutputTokens
	var lastErr error

	for attempt := 1; attempt <= openAIMaxJSONAttempts; attempt++ {
		attemptDebug, err := s.callOpenAIJSONResponseAttempt(
			ctx,
			model,
			systemPrompt,
			userPrompt,
			tools,
			schemaName,
			schema,
			currentMaxTokens,
			output,
		)
		attemptDebug["attempt"] = attempt
		attemptDebug["max_output_tokens"] = currentMaxTokens
		attempts = append(attempts, attemptDebug)

		if err == nil {
			debugStep["attempts"] = attempts
			debugStep["final_attempt"] = attempt
			return debugStep, nil
		}

		lastErr = err
		if !shouldRetryOpenAIJSONResponse(err) || attempt == openAIMaxJSONAttempts {
			debugStep["attempts"] = attempts
			return debugStep, err
		}

		currentMaxTokens = nextOpenAIMaxOutputTokens(currentMaxTokens)
	}

	debugStep["attempts"] = attempts
	return debugStep, lastErr
}

func (s *analysisService) callOpenAIJSONResponseAttempt(ctx context.Context, model, systemPrompt, userPrompt string, tools []map[string]any, schemaName string, schema map[string]any, maxOutputTokens int, output any) (map[string]any, error) {
	payload := map[string]any{
		"model": model,
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
		"max_output_tokens": maxOutputTokens,
		"text": map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   schemaName,
				"strict": true,
				"schema": schema,
			},
		},
	}
	if len(tools) > 0 {
		payload["tools"] = tools
		payload["include"] = []string{"web_search_call.action.sources"}
	}

	debugStep := map[string]any{}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return debugStep, (&apiError{
			Status:  http.StatusInternalServerError,
			Code:    "openai_payload_error",
			Message: "impossible de construire la requete OpenAI",
		}).WithDetail("reason", err.Error()).WithCause(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/responses", bytes.NewReader(requestBody))
	if err != nil {
		return debugStep, (&apiError{
			Status:  http.StatusInternalServerError,
			Code:    "openai_request_error",
			Message: "impossible de creer la requete OpenAI",
		}).WithDetail("reason", err.Error()).WithCause(err)
	}
	clientRequestID := newRequestID()
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Request-Id", clientRequestID)
	debugStep["client_request_id"] = clientRequestID

	resp, err := s.client.Do(req)
	if err != nil {
		return debugStep, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_unreachable",
			Message: "impossible de contacter OpenAI",
		}).WithDetail("base_url", s.baseURL).WithDetail("reason", err.Error()).WithCause(err)
	}
	defer resp.Body.Close()

	debugStep["openai_request_id"] = resp.Header.Get("X-Request-Id")
	debugStep["http_status"] = resp.StatusCode

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, maxFetchedBytes))
	if err != nil {
		return debugStep, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_read_error",
			Message: "lecture de la reponse OpenAI impossible",
		}).WithDetail("base_url", s.baseURL).WithDetail("reason", err.Error()).WithCause(err)
	}
	debugStep["response_excerpt"] = excerptText(string(responseBody), 4000)

	if resp.StatusCode >= http.StatusBadRequest {
		return debugStep, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_http_error",
			Message: fmt.Sprintf("OpenAI a repondu HTTP %d: %s", resp.StatusCode, extractOpenAIErrorMessage(responseBody)),
		}).WithDetail("upstream_status", resp.StatusCode).
			WithDetail("response_excerpt", excerptText(string(responseBody), 220))
	}

	var parsed openAIResponsesResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return debugStep, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_parse_error",
			Message: "reponse OpenAI invalide",
		}).WithDetail("reason", err.Error()).
			WithDetail("response_excerpt", excerptText(string(responseBody), 220)).
			WithCause(err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return debugStep, &apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_error",
			Message: parsed.Error.Message,
		}
	}

	debugStep["response_id"] = parsed.ID
	debugStep["usage"] = parsed.Usage
	debugStep["status"] = parsed.Status
	if parsed.IncompleteDetails != nil && strings.TrimSpace(parsed.IncompleteDetails.Reason) != "" {
		debugStep["incomplete_reason"] = parsed.IncompleteDetails.Reason
	}

	rawJSON := strings.TrimSpace(parsed.OutputText)
	aggregatedOutput := strings.TrimSpace(extractTextFromOutput(parsed.Output))
	if len([]rune(aggregatedOutput)) > len([]rune(rawJSON)) {
		rawJSON = aggregatedOutput
	}
	if rawJSON == "" {
		return debugStep, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_empty_output",
			Message: "OpenAI n'a renvoye aucun contenu exploitable",
		}).WithDetail("response_status", parsed.Status).
			WithDetail("incomplete_reason", incompleteReason(parsed)).
			WithDetail("response_excerpt", excerptText(string(responseBody), 220))
	}

	debugStep["raw_output"] = excerptText(rawJSON, 4000)

	decodedJSON, err := decodeOpenAIJSON(output, rawJSON, aggregatedOutput)
	if err != nil {
		return debugStep, (&apiError{
			Status:  http.StatusBadGateway,
			Code:    "openai_output_invalid",
			Message: "OpenAI n'a pas renvoye le JSON attendu",
		}).WithDetail("reason", err.Error()).
			WithDetail("response_status", parsed.Status).
			WithDetail("incomplete_reason", incompleteReason(parsed)).
			WithDetail("response_excerpt", excerptText(rawJSON, 220)).
			WithCause(err)
	}
	debugStep["decoded_output"] = excerptText(decodedJSON, 4000)

	return debugStep, nil
}

func shouldRetryOpenAIJSONResponse(err error) bool {
	apiErr, ok := err.(*apiError)
	if !ok {
		return false
	}

	switch apiErr.Code {
	case "openai_empty_output":
		return true
	case "openai_output_invalid":
		reason, _ := apiErr.Details["reason"].(string)
		responseStatus, _ := apiErr.Details["response_status"].(string)
		incompleteReason, _ := apiErr.Details["incomplete_reason"].(string)
		return strings.Contains(strings.ToLower(reason), "unexpected end of json input") ||
			strings.EqualFold(responseStatus, "incomplete") ||
			strings.EqualFold(incompleteReason, "max_output_tokens")
	default:
		return false
	}
}

func nextOpenAIMaxOutputTokens(current int) int {
	if current <= 0 {
		current = 1000
	}
	next := current + openAIRetryTokenBump
	if next > openAIMaxOutputTokenCap {
		return openAIMaxOutputTokenCap
	}
	return next
}

func writeAPIError(c *gin.Context, err error, appEnv string, debugMode bool) {
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
	if (debugMode || shouldExposeErrorDetails(appEnv)) && len(apiErr.Details) > 0 {
		payload["details"] = apiErr.Details
	}
	if debugMode && len(apiErr.Debug) > 0 {
		payload["debug"] = apiErr.Debug
	}

	c.JSON(apiErr.Status, payload)
}

func (s *analysisService) buildSuccessResponse(c *gin.Context, sourceURL, model string, doc fetchedDocument, analysis summaryAnalysis, meta map[string]any, debugPayload map[string]any) gin.H {
	response := gin.H{
		"status":       "ok",
		"request_id":   getRequestID(c),
		"source_url":   sourceURL,
		"model":        model,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"fetched": gin.H{
			"http_status":  doc.HTTPStatus,
			"content_type": doc.ContentType,
			"characters":   doc.Characters,
			"source":       doc.AcquiredVia,
		},
		"analysis": analysis,
		"meta":     meta,
	}

	if s.debugMode && len(debugPayload) > 0 {
		response["debug"] = debugPayload
	}

	return response
}

func (s *analysisService) isCacheStale(analyzedAt time.Time) bool {
	return isCachedAnalysisStale(analyzedAt, s.cacheMaxAge, time.Now().UTC())
}

func attachDebugPayload(err error, debugPayload map[string]any) error {
	if len(debugPayload) == 0 {
		return err
	}

	if apiErr, ok := err.(*apiError); ok {
		return apiErr.WithDebug(debugPayload)
	}

	return (&apiError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: "erreur interne du serveur",
		Err:     err,
	}).WithDebug(debugPayload)
}

func attachDetail(err error, key string, value any) error {
	if apiErr, ok := err.(*apiError); ok {
		return apiErr.WithDetail(key, value)
	}
	return err
}

func sanitizeAnalysis(input summaryAnalysis) summaryAnalysis {
	input.ServiceName = strings.TrimSpace(input.ServiceName)
	if input.ServiceName == "" {
		input.ServiceName = "Service analyse"
	}
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
	switch input.Confidence {
	case "low", "medium", "high":
	default:
		input.Confidence = "medium"
	}
	input.IsContestable = input.Confidence == "low"
	input.Points = sanitizePoints(input.Points)
	input.Highlights = deriveHighlightsFromPoints(input.Points, input.Highlights)
	input.Risks = deriveRisksFromPoints(input.Points, input.Risks)
	return input
}

func sanitizeResearch(input researchResult) researchResult {
	if normalizedTargetURL, err := normalizeDocumentURL(input.TargetURL); err == nil {
		input.TargetURL = normalizedTargetURL
	} else {
		input.TargetURL = strings.TrimSpace(input.TargetURL)
	}
	input.TargetTitle = strings.TrimSpace(input.TargetTitle)
	if input.TargetTitle == "" {
		input.TargetTitle = "Page de CGU identifiee"
	}
	input.TargetContent = strings.TrimSpace(input.TargetContent)
	input.Summary = strings.TrimSpace(input.Summary)
	if input.Summary == "" {
		input.Summary = "Aucune page de CGU clairement exploitable n'a ete retrouvee."
	}
	input.Confidence = strings.ToLower(strings.TrimSpace(input.Confidence))
	switch input.Confidence {
	case "low", "medium", "high":
	default:
		input.Confidence = "medium"
	}
	input.KeyFindings = sanitizeList(input.KeyFindings)
	input.Sources = sanitizeSources(input.Sources)
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

func sanitizeSources(items []researchSource) []researchSource {
	cleaned := make([]researchSource, 0, len(items))
	for _, item := range items {
		source := researchSource{
			Title:  strings.TrimSpace(item.Title),
			URL:    strings.TrimSpace(item.URL),
			Domain: strings.TrimSpace(item.Domain),
		}
		if source.Title == "" && source.URL == "" {
			continue
		}
		cleaned = append(cleaned, source)
		if len(cleaned) >= 6 {
			break
		}
	}
	return cleaned
}

func sanitizePoints(items []analysisPoint) []analysisPoint {
	cleaned := make([]analysisPoint, 0, len(items))
	for _, item := range items {
		point := analysisPoint{
			Category: strings.ToLower(strings.TrimSpace(item.Category)),
			Title:    strings.TrimSpace(item.Title),
			Details:  strings.TrimSpace(item.Details),
		}

		switch point.Category {
		case "blocker", "bad", "neutral", "good":
		default:
			point.Category = "neutral"
		}

		if point.Title == "" {
			continue
		}
		if point.Details == "" {
			point.Details = point.Title
		}

		cleaned = append(cleaned, point)
		if len(cleaned) >= 24 {
			break
		}
	}

	if len(cleaned) == 0 {
		return []analysisPoint{{
			Category: "neutral",
			Title:    "Aucun point exploitable detecte",
			Details:  "Le texte collecte ne permet pas d'identifier de clauses assez nettes.",
		}}
	}

	return cleaned
}

func deriveHighlightsFromPoints(points []analysisPoint, fallback []string) []string {
	derived := make([]string, 0, maxListItems)
	for _, point := range points {
		if point.Category != "good" {
			continue
		}
		derived = append(derived, point.Title)
		if len(derived) >= maxListItems {
			break
		}
	}
	if len(derived) > 0 {
		return derived
	}
	return sanitizeList(fallback)
}

func deriveRisksFromPoints(points []analysisPoint, fallback []string) []string {
	derived := make([]string, 0, maxListItems)
	for _, point := range points {
		if point.Category != "blocker" && point.Category != "bad" {
			continue
		}
		derived = append(derived, point.Title)
		if len(derived) >= maxListItems {
			break
		}
	}
	if len(derived) > 0 {
		return derived
	}
	return sanitizeList(fallback)
}

func researchToPromptText(research researchResult) string {
	var builder strings.Builder
	builder.WriteString("Page cible: ")
	builder.WriteString(strings.TrimSpace(research.TargetTitle))
	if research.TargetURL != "" {
		builder.WriteString(" ")
		builder.WriteString(strings.TrimSpace(research.TargetURL))
	}
	builder.WriteString("\nConfiance recherche: ")
	builder.WriteString(strings.TrimSpace(research.Confidence))
	builder.WriteString("\n\nExtrait retrouve:\n")
	builder.WriteString(strings.TrimSpace(research.TargetContent))
	builder.WriteString("\n\nResume: ")
	builder.WriteString(strings.TrimSpace(research.Summary))
	builder.WriteString("\n\nPoints cles:\n")
	for _, finding := range sanitizeList(research.KeyFindings) {
		builder.WriteString("- ")
		builder.WriteString(finding)
		builder.WriteByte('\n')
	}
	if len(research.Sources) > 0 {
		builder.WriteString("\nSources:\n")
		for _, source := range research.Sources {
			builder.WriteString("- ")
			builder.WriteString(source.Title)
			if source.Domain != "" {
				builder.WriteString(" [")
				builder.WriteString(source.Domain)
				builder.WriteString("]")
			}
			if source.URL != "" {
				builder.WriteString(" ")
				builder.WriteString(source.URL)
			}
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func normalizeURL(raw string) (string, error) {
	parsed, err := parseLooseURL(raw)
	if err != nil {
		return "", fmt.Errorf("URL invalide")
	}

	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" {
		return "", fmt.Errorf("URL invalide: domaine manquant")
	}
	return "https://" + host, nil
}

func normalizeDocumentURL(raw string) (string, error) {
	parsed, err := parseLooseURL(raw)
	if err != nil {
		return "", fmt.Errorf("URL invalide")
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("URL invalide: domaine manquant")
	}

	parsed.User = nil
	parsed.Fragment = ""
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	return parsed.String(), nil
}

func parseLooseURL(raw string) (*neturl.URL, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("URL invalide")
	}

	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}

	parsed, err := neturl.Parse(trimmed)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("URL invalide: seul http/https est accepte")
	}
	return parsed, nil
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

func extractTextFromOutput(output []openAIResponsesOutput) string {
	var combined strings.Builder

	for _, item := range output {
		var messageText strings.Builder
		for _, content := range item.Content {
			if content.Type == "output_text" || content.Type == "text" {
				text := strings.TrimSpace(content.Text)
				if text != "" {
					messageText.WriteString(text)
				}
			}
		}

		joined := strings.TrimSpace(messageText.String())
		if joined == "" {
			continue
		}
		combined.WriteString(joined)
	}

	return strings.TrimSpace(combined.String())
}

func decodeOpenAIJSON(output any, rawCandidates ...string) (string, error) {
	var firstErr error
	for _, raw := range rawCandidates {
		for _, candidate := range jsonDecodeCandidates(raw) {
			if err := json.Unmarshal([]byte(candidate), output); err == nil {
				return candidate, nil
			} else if firstErr == nil {
				firstErr = err
			}
		}
	}

	if firstErr == nil {
		firstErr = fmt.Errorf("aucun JSON exploitable detecte")
	}
	return "", firstErr
}

func jsonDecodeCandidates(raw string) []string {
	seen := map[string]struct{}{}
	candidates := make([]string, 0, 4)

	appendCandidate := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		candidates = append(candidates, value)
	}

	appendCandidate(raw)

	strippedFence := stripMarkdownCodeFence(raw)
	appendCandidate(strippedFence)

	if extracted := extractJSONObject(raw); extracted != "" {
		appendCandidate(extracted)
	}
	if extracted := extractJSONObject(strippedFence); extracted != "" {
		appendCandidate(extracted)
	}

	return candidates
}

func stripMarkdownCodeFence(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}

	lines := strings.Split(trimmed, "\n")
	if len(lines) < 2 {
		return trimmed
	}

	if strings.HasPrefix(lines[0], "```") {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
		lines = lines[:len(lines)-1]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func extractJSONObject(raw string) string {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return strings.TrimSpace(raw[start : end+1])
}

func incompleteReason(parsed openAIResponsesResponse) string {
	if parsed.IncompleteDetails == nil {
		return ""
	}
	return strings.TrimSpace(parsed.IncompleteDetails.Reason)
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
