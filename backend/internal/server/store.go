package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	neturl "net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type cachedAnalysisRecord struct {
	NormalizedURL  string
	SourceURL      string
	Domain         string
	FetchedAt      time.Time
	AnalyzedAt     time.Time
	HTTPStatus     int
	ContentType    string
	CharacterCount int
	RawText        string
	Model          string
	IsContestable  bool
	Research       researchResult
	Analysis       summaryAnalysis
	Debug          map[string]any
}

func getDBPool(c *gin.Context) *pgxpool.Pool {
	if c == nil {
		return nil
	}

	value, exists := c.Get("db")
	if !exists {
		return nil
	}

	pool, ok := value.(*pgxpool.Pool)
	if !ok {
		return nil
	}
	return pool
}

func loadCachedAnalysis(ctx context.Context, pool *pgxpool.Pool, normalizedURL string) (*cachedAnalysisRecord, error) {
	if pool == nil {
		return nil, nil
	}

	record := &cachedAnalysisRecord{}
	var (
		researchJSON []byte
		analysisJSON []byte
		debugJSON    []byte
	)

	err := pool.QueryRow(ctx, `
		SELECT
			normalized_url,
			source_url,
			domain,
			fetched_at,
			analyzed_at,
			COALESCE(http_status, 0),
			COALESCE(content_type, ''),
			COALESCE(character_count, 0),
			COALESCE(raw_text, ''),
			model,
			is_contestable,
			research_json,
			analysis_json,
			debug_json
		FROM tosai_cached_analyses
		WHERE normalized_url = $1
		LIMIT 1
	`, normalizedURL).Scan(
		&record.NormalizedURL,
		&record.SourceURL,
		&record.Domain,
		&record.FetchedAt,
		&record.AnalyzedAt,
		&record.HTTPStatus,
		&record.ContentType,
		&record.CharacterCount,
		&record.RawText,
		&record.Model,
		&record.IsContestable,
		&researchJSON,
		&analysisJSON,
		&debugJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("load cached analysis: %w", err)
	}

	if len(researchJSON) > 0 {
		if err := json.Unmarshal(researchJSON, &record.Research); err != nil {
			return nil, fmt.Errorf("decode cached research: %w", err)
		}
	}
	if len(analysisJSON) > 0 {
		if err := json.Unmarshal(analysisJSON, &record.Analysis); err != nil {
			return nil, fmt.Errorf("decode cached analysis: %w", err)
		}
	}
	if len(debugJSON) > 0 {
		if err := json.Unmarshal(debugJSON, &record.Debug); err != nil {
			return nil, fmt.Errorf("decode cached debug: %w", err)
		}
	}

	return record, nil
}

func saveCachedAnalysis(ctx context.Context, pool *pgxpool.Pool, normalizedURL string, doc fetchedDocument, research researchResult, analysis summaryAnalysis, model string, debug map[string]any) error {
	if pool == nil {
		return nil
	}

	parsedURL, err := neturl.Parse(normalizedURL)
	if err != nil {
		return fmt.Errorf("parse normalized url for cache: %w", err)
	}

	researchJSON, err := marshalJSONB(research)
	if err != nil {
		return fmt.Errorf("marshal research json: %w", err)
	}
	analysisJSON, err := marshalJSONB(analysis)
	if err != nil {
		return fmt.Errorf("marshal analysis json: %w", err)
	}
	debugJSON, err := marshalJSONB(debug)
	if err != nil {
		return fmt.Errorf("marshal debug json: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO tosai_cached_analyses (
			normalized_url,
			source_url,
			domain,
			fetched_at,
			analyzed_at,
			http_status,
			content_type,
			character_count,
			raw_text,
			model,
			is_contestable,
			research_json,
			analysis_json,
			debug_json
		)
		VALUES ($1, $2, $3, NOW(), NOW(), $4, $5, $6, $7, $8, $9, $10::jsonb, $11::jsonb, $12::jsonb)
		ON CONFLICT (normalized_url) DO UPDATE SET
			source_url = EXCLUDED.source_url,
			domain = EXCLUDED.domain,
			fetched_at = NOW(),
			analyzed_at = NOW(),
			http_status = EXCLUDED.http_status,
			content_type = EXCLUDED.content_type,
			character_count = EXCLUDED.character_count,
			raw_text = EXCLUDED.raw_text,
			model = EXCLUDED.model,
			is_contestable = EXCLUDED.is_contestable,
			research_json = EXCLUDED.research_json,
			analysis_json = EXCLUDED.analysis_json,
			debug_json = EXCLUDED.debug_json
	`,
		normalizedURL,
		doc.SourceURL,
		parsedURL.Hostname(),
		doc.HTTPStatus,
		doc.ContentType,
		doc.Characters,
		doc.Text,
		model,
		analysis.IsContestable,
		string(researchJSON),
		string(analysisJSON),
		string(debugJSON),
	)
	if err != nil {
		return fmt.Errorf("upsert cached analysis: %w", err)
	}

	return nil
}

func marshalJSONB(value any) ([]byte, error) {
	if value == nil {
		return []byte(`{}`), nil
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return []byte(`{}`), nil
	}
	return raw, nil
}
