package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds configuration values loaded from environment variables.
type Config struct {
	AppEnv                  string
	Port                    string
	DatabaseURL             string
	OpenAIAPIKey            string
	OpenAIModel             string
	OpenAIResearchModel     string
	OpenAIBaseURL           string
	AdminAPIToken           string
	CORSOrigins             []string
	AllowAllOrigins         bool
	TrustedProxies          []string
	HTTPTimeoutSec          int
	AnalysisMaxChar         int
	AnalysisCacheMaxAgeDays int
	AnalysisRateLimitPerMin int
	APIDebugMode            bool
}

// Load reads environment variables into a Config struct.
func Load() Config {
	loadDotenv()

	corsRaw := getenvDefault("CORS_ORIGINS", getenvDefault("ALLOWED_ORIGINS", "*"))
	corsOrigins := parseOrigins(corsRaw)
	trustedProxies := parseCSV(getenvDefault("TRUSTED_PROXIES", "127.0.0.1,::1"))

	cfg := Config{
		AppEnv:                  getenvDefault("APP_ENV", "local"),
		Port:                    getenvDefault("APP_PORT", getenvDefault("PORT", "9000")),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:            os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:             getenvDefault("OPENAI_MODEL", "gpt-5.4-mini"),
		OpenAIResearchModel:     getenvDefault("OPENAI_RESEARCH_MODEL", getenvDefault("OPENAI_MODEL", "gpt-5.4-mini")),
		OpenAIBaseURL:           strings.TrimRight(getenvDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"), "/"),
		AdminAPIToken:           os.Getenv("ADMIN_API_TOKEN"),
		CORSOrigins:             corsOrigins,
		AllowAllOrigins:         len(corsOrigins) == 1 && corsOrigins[0] == "*",
		TrustedProxies:          trustedProxies,
		HTTPTimeoutSec:          getenvIntDefault("HTTP_TIMEOUT_SECONDS", 45),
		AnalysisMaxChar:         getenvIntDefault("ANALYSIS_INPUT_MAX_CHARS", 12000),
		AnalysisCacheMaxAgeDays: getenvIntDefault("ANALYSIS_CACHE_MAX_AGE_DAYS", 90),
		AnalysisRateLimitPerMin: getenvIntDefault("ANALYSIS_RATE_LIMIT_PER_MINUTE", 1),
		APIDebugMode:            getenvBoolDefault("API_DEBUG_MODE", false),
	}

	if cfg.DatabaseURL == "" {
		log.Println("avertissement: DATABASE_URL n'est pas défini ; backend lancé en mode API-only (sans persistance PostgreSQL)")
	}

	if cfg.OpenAIAPIKey == "" {
		log.Println("avertissement: OPENAI_API_KEY n'est pas défini ; l'endpoint /v1/summary renverra une erreur")
	}

	return cfg
}

func loadDotenv() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("avertissement: impossible de déterminer le répertoire courant: %v", err)
	}

	backendDir := cwd
	if filepath.Base(cwd) != "backend" {
		backendDir = filepath.Join(cwd, "backend")
	}

	candidates := []string{
		filepath.Join(backendDir, ".env"),
		filepath.Join(backendDir, "config", ".env"),
	}

	// Si l'application est lancée depuis backend/, on tente aussi les chemins relatifs déjà présents.
	if filepath.Base(cwd) == "backend" {
		candidates = append(candidates,
			filepath.Join(cwd, "config", ".env"),
		)
	} else {
		candidates = append(candidates,
			filepath.Join(cwd, ".env"),
			filepath.Join(cwd, "config", ".env"),
		)
	}

	seen := map[string]struct{}{}
	loaded := false
	for _, path := range candidates {
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}

		if _, err := os.Stat(path); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Printf("avertissement: impossible de lire %s: %v", path, err)
			}
			continue
		}

		if err := godotenv.Load(path); err != nil {
			log.Printf("avertissement: échec du chargement de %s: %v", path, err)
			continue
		}
		log.Printf("fichier d'environnement chargé: %s", path)
		loaded = true
		break
	}

	if !loaded {
		log.Printf("avertissement: aucun fichier .env trouvé (backend/.env prioritaire, puis backend/config/.env). On suppose que les variables sont déjà présentes dans l'environnement.")
	}
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvIntDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		log.Printf("avertissement: %s invalide (%q), fallback=%d", key, value, fallback)
		return fallback
	}
	return parsed
}

func getenvBoolDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		log.Printf("avertissement: %s invalide (%q), fallback=%t", key, value, fallback)
		return fallback
	}
}

func parseOrigins(raw string) []string {
	parts := parseCSV(raw)
	if len(parts) == 0 {
		return []string{"*"}
	}
	return parts
}

func parseCSV(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	parts := strings.Split(trimmed, ",")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		cleaned = append(cleaned, value)
	}
	return cleaned
}
