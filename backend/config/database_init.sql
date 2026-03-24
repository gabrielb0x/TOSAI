-- Schéma PostgreSQL TOSAI (PostgreSQL 15+)
-- Peut être appliqué en toute sécurité via :
--   psql -d tosai -f backend/config/database_init.sql
-- ⚠️ Ce script ne crée ni rôle ni base de données : utilisez le README pour ces étapes.

-- Extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 1) Domaines suivis
CREATE TABLE IF NOT EXISTS tosai_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2) Limites par domaine
CREATE TABLE IF NOT EXISTS tosai_domain_limits (
    domain_id UUID PRIMARY KEY REFERENCES tosai_domains(id) ON DELETE CASCADE,
    max_per_day INT NOT NULL DEFAULT 20,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3) Usage quotidien par domaine
CREATE TABLE IF NOT EXISTS tosai_domain_usage_daily (
    domain_id UUID REFERENCES tosai_domains(id) ON DELETE CASCADE,
    day DATE NOT NULL,
    count INT NOT NULL DEFAULT 0,
    PRIMARY KEY(domain_id, day)
);

-- 4) Documents récupérés
CREATE TABLE IF NOT EXISTS tosai_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain_id UUID NOT NULL REFERENCES tosai_domains(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    http_status INT,
    content_sha256 CHAR(64) NOT NULL,
    raw_text TEXT,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE(domain_id, content_sha256)
);

CREATE INDEX IF NOT EXISTS idx_tosai_documents_domain ON tosai_documents(domain_id);
CREATE INDEX IF NOT EXISTS idx_tosai_documents_domain_url ON tosai_documents(domain_id, url);

-- 5) Analyses IA
CREATE TABLE IF NOT EXISTS tosai_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES tosai_documents(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    model TEXT NOT NULL,
    prompt_version TEXT NOT NULL DEFAULT 'v1',
    rating CHAR(1) NOT NULL CHECK (rating IN ('A','B','C','D','E')),
    is_reportable BOOLEAN NOT NULL DEFAULT FALSE,
    summary_md TEXT,
    summary_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    score_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    risk_flags JSONB NOT NULL DEFAULT '{}'::jsonb,
    tokens_in INT,
    tokens_out INT,
    cost_usd NUMERIC(10,4)
);

CREATE INDEX IF NOT EXISTS idx_tosai_analyses_document ON tosai_analyses(document_id);
CREATE INDEX IF NOT EXISTS idx_tosai_analyses_created_at ON tosai_analyses(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tosai_analyses_rating ON tosai_analyses(rating);

ALTER TABLE tosai_analyses
    ADD COLUMN IF NOT EXISTS is_contestable BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE tosai_analyses
    ADD COLUMN IF NOT EXISTS research_json JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE tosai_analyses
    ADD COLUMN IF NOT EXISTS debug_json JSONB NOT NULL DEFAULT '{}'::jsonb;

-- 5b) Cache par URL normalisee pour eviter les appels OpenAI inutiles
CREATE TABLE IF NOT EXISTS tosai_cached_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    normalized_url TEXT NOT NULL UNIQUE,
    source_url TEXT NOT NULL,
    domain TEXT NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    analyzed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    http_status INT,
    content_type TEXT,
    character_count INT NOT NULL DEFAULT 0,
    raw_text TEXT,
    model TEXT NOT NULL,
    is_contestable BOOLEAN NOT NULL DEFAULT FALSE,
    research_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    analysis_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    debug_json JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_tosai_cached_analyses_analyzed_at ON tosai_cached_analyses(analyzed_at DESC);
CREATE INDEX IF NOT EXISTS idx_tosai_cached_analyses_domain ON tosai_cached_analyses(domain);

-- 6) Signalements utilisateurs
CREATE TABLE IF NOT EXISTS tosai_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    url TEXT NOT NULL,
    domain_id UUID REFERENCES tosai_domains(id) ON DELETE SET NULL,
    reason TEXT NOT NULL,
    details TEXT,
    status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new','triaged','dismissed','accepted')),
    admin_note TEXT
);

CREATE INDEX IF NOT EXISTS idx_tosai_reports_status ON tosai_reports(status);
CREATE INDEX IF NOT EXISTS idx_tosai_reports_created_at ON tosai_reports(created_at DESC);

-- 7) Audit admin (optionnel)
CREATE TABLE IF NOT EXISTS tosai_admin_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor TEXT NOT NULL,
    action TEXT NOT NULL,
    target_type TEXT,
    target_id UUID,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_tosai_admin_events_created_at ON tosai_admin_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tosai_admin_events_action ON tosai_admin_events(action);
