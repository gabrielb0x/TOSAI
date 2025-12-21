-- Minimal PostgreSQL bootstrap for local installs (run with a superuser).
-- Usage:
--   psql -f backend/config/database_init.sql

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'autotos') THEN
        CREATE ROLE autotos LOGIN PASSWORD 'autotos';
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'autotos') THEN
        CREATE DATABASE autotos OWNER autotos;
    END IF;
END$$;

GRANT ALL PRIVILEGES ON DATABASE autotos TO autotos;
