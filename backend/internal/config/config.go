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
	AppEnv          string
	Port            string
	DatabaseURL     string
	OpenAIAPIKey    string
	OpenAIModel     string
	OpenAIBaseURL   string
	AdminAPIToken   string
	CORSOrigins     []string
	AllowAllOrigins bool
	HTTPTimeoutSec  int
	AnalysisMaxChar int
}

// Load reads environment variables into a Config struct.
func Load() Config {
	loadDotenv()

	corsRaw := getenvDefault("CORS_ORIGINS", getenvDefault("ALLOWED_ORIGINS", "*"))
	corsOrigins := parseOrigins(corsRaw)

	cfg := Config{
		AppEnv:          getenvDefault("APP_ENV", "local"),
		Port:            getenvDefault("APP_PORT", getenvDefault("PORT", "9000")),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:     getenvDefault("OPENAI_MODEL", "gpt-4.1-mini"),
		OpenAIBaseURL:   strings.TrimRight(getenvDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"), "/"),
		AdminAPIToken:   os.Getenv("ADMIN_API_TOKEN"),
		CORSOrigins:     corsOrigins,
		AllowAllOrigins: len(corsOrigins) == 1 && corsOrigins[0] == "*",
		HTTPTimeoutSec:  getenvIntDefault("HTTP_TIMEOUT_SECONDS", 25),
		AnalysisMaxChar: getenvIntDefault("ANALYSIS_INPUT_MAX_CHARS", 12000),
	}

	if cfg.DatabaseURL == "" {
		log.Println("avertissement: DATABASE_URL n'est pas défini ; backend lancé en mode API-only (sans persistance PostgreSQL)")
	}

	if cfg.OpenAIAPIKey == "" {
		log.Println("avertissement: OPENAI_API_KEY n'est pas défini ; l'endpoint /api/v1/summary renverra une erreur")
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

func parseOrigins(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{"*"}
	}
	parts := strings.Split(trimmed, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
