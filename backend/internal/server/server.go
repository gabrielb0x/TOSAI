package server

import (
	"log"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gabrielb0x/TOSAI/backend/internal/config"
)

// New sets up the Gin engine with routes and middleware.
func New(cfg config.Config, pool *pgxpool.Pool) *gin.Engine {
	r := gin.Default()
	analysisSvc := newAnalysisService(cfg)
	rateLimiter := newRateLimiter(cfg.AnalysisRateLimitPerMin, time.Minute)

	r.Use(requestIDMiddleware())
	r.Use(func(c *gin.Context) {
		if pool != nil {
			c.Set("db", pool)
		}
		c.Next()
	})

	corsCfg := cors.DefaultConfig()
	if cfg.AllowAllOrigins {
		corsCfg.AllowAllOrigins = true
	} else {
		corsCfg.AllowOrigins = cfg.CORSOrigins
	}
	corsCfg.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsCfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsCfg.AllowPrivateNetwork = true
	corsCfg.ExposeHeaders = []string{"Content-Length"}
	corsCfg.AllowOrigins = trimEmpty(corsCfg.AllowOrigins)
	r.Use(cors.New(corsCfg))

	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Printf("avertissement: TRUSTED_PROXIES invalide (%v) ; trust proxies desactive", err)
		if disableErr := r.SetTrustedProxies(nil); disableErr != nil {
			log.Printf("avertissement: impossible de desactiver trust proxies: %v", disableErr)
		}
	}

	api := r.Group("/v1")
	api.GET("", handleAPIRoot)
	api.GET("/", handleAPIRoot)
	registerSummaryRoutes(api, analysisSvc, summaryRateLimitMiddleware(rateLimiter, analysisSvc.cacheMaxAge, cfg.AppEnv, cfg.APIDebugMode))

	registerSummaryRoutes(r, analysisSvc, summaryRateLimitMiddleware(rateLimiter, analysisSvc.cacheMaxAge, cfg.AppEnv, cfg.APIDebugMode))

	return r
}

type summaryRouter interface {
	GET(string, ...gin.HandlerFunc) gin.IRoutes
	POST(string, ...gin.HandlerFunc) gin.IRoutes
}

func registerSummaryRoutes(router summaryRouter, analysisSvc *analysisService, middleware ...gin.HandlerFunc) {
	getHandlers := append([]gin.HandlerFunc{}, middleware...)
	getHandlers = append(getHandlers, analysisSvc.handleSummaryGET)

	postHandlers := append([]gin.HandlerFunc{}, middleware...)
	postHandlers = append(postHandlers, analysisSvc.handleSummaryPOST)

	router.GET("/summary", getHandlers...)
	router.POST("/summary", postHandlers...)
}

func trimEmpty(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			continue
		}
		cleaned = append(cleaned, v)
	}
	if len(cleaned) == 0 {
		return []string{"*"}
	}
	return cleaned
}
