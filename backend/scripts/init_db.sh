#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${BACKEND_DIR}/.env"
ENV_EXAMPLE_FILE="${BACKEND_DIR}/.env.example"
SCHEMA_FILE="${BACKEND_DIR}/config/database_init.sql"

if ! command -v psql >/dev/null 2>&1; then
  echo "psql est requis pour initialiser PostgreSQL." >&2
  exit 1
fi

if [ ! -f "${ENV_FILE}" ] && [ -f "${ENV_EXAMPLE_FILE}" ]; then
  cp "${ENV_EXAMPLE_FILE}" "${ENV_FILE}"
  echo "backend/.env cree depuis .env.example"
fi

if [ -f "${ENV_FILE}" ]; then
  set -a
  # shellcheck disable=SC1090
  . "${ENV_FILE}"
  set +a
fi

DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-tosai}"
DB_USER="${DB_USER:-tosai_app}"
DB_PASSWORD="${DB_PASSWORD:-change-me}"
DB_SSLMODE="${DB_SSLMODE:-disable}"

if [ -n "${DATABASE_URL:-}" ]; then
  APP_DATABASE_URL="${DATABASE_URL}"
else
  APP_DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
fi

escape_sed_replacement() {
  printf '%s' "$1" | sed 's/[&|]/\\&/g'
}

upsert_env_var() {
  local file="$1"
  local key="$2"
  local value="$3"
  local escaped_value

  escaped_value="$(escape_sed_replacement "${value}")"

  if grep -q "^${key}=" "${file}"; then
    sed -i "s|^${key}=.*$|${key}=${escaped_value}|" "${file}"
    return
  fi

  printf '\n%s=%s\n' "${key}" "${value}" >>"${file}"
}

run_admin_psql() {
  if [ -n "${POSTGRES_ADMIN_URL:-}" ]; then
    psql "${POSTGRES_ADMIN_URL}" -v ON_ERROR_STOP=1 "$@"
    return
  fi

  if command -v sudo >/dev/null 2>&1 && [ "$(id -un)" != "postgres" ]; then
    sudo -u postgres psql -d postgres -v ON_ERROR_STOP=1 "$@"
    return
  fi

  psql "postgresql://${POSTGRES_ADMIN_USER:-postgres}@${DB_HOST}:${DB_PORT}/postgres?sslmode=${DB_SSLMODE}" -v ON_ERROR_STOP=1 "$@"
}

echo "Initialisation PostgreSQL pour ${DB_NAME}..."

run_admin_psql -v db_user="${DB_USER}" -v db_password="${DB_PASSWORD}" -v db_name="${DB_NAME}" <<'SQL'
SELECT format('CREATE ROLE %I LOGIN PASSWORD %L', :'db_user', :'db_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'db_user')
\gexec

SELECT format('ALTER ROLE %I WITH LOGIN PASSWORD %L', :'db_user', :'db_password')
WHERE EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'db_user')
\gexec

SELECT format('CREATE DATABASE %I OWNER %I', :'db_name', :'db_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'db_name')
\gexec

SELECT format('ALTER DATABASE %I OWNER TO %I', :'db_name', :'db_user')
\gexec
SQL

psql "${APP_DATABASE_URL}" -v ON_ERROR_STOP=1 -f "${SCHEMA_FILE}"

if [ -f "${ENV_FILE}" ]; then
  upsert_env_var "${ENV_FILE}" "DATABASE_URL" "${APP_DATABASE_URL}"
  echo "DATABASE_URL mis a jour dans backend/.env"
fi

echo "Base PostgreSQL prete."
echo "Connexion applicative: ${APP_DATABASE_URL}"
