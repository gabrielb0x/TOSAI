package server

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gabrielb0x/TOSAI/backend/internal/config"
)

// New sets up the Gin engine with routes and middleware.
func New(cfg config.Config, pool *pgxpool.Pool) *gin.Engine {
	r := gin.Default()
	analysisSvc := newAnalysisService(cfg)

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

	registerSummaryRoutes(r, analysisSvc)

	api := r.Group("/v1")
	registerSummaryRoutes(api, analysisSvc)

	return r
}

type summaryRouter interface {
	GET(string, ...gin.HandlerFunc) gin.IRoutes
	POST(string, ...gin.HandlerFunc) gin.IRoutes
}

func registerSummaryRoutes(router summaryRouter, analysisSvc *analysisService) {
	router.GET("/summary", analysisSvc.handleSummaryGET)
	router.POST("/summary", analysisSvc.handleSummaryPOST)
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
