package server

import (
	"net/http"

	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gabrielb0x/TOSAI/backend/internal/config"
)

// New sets up the Gin engine with routes and middleware.
func New(cfg config.Config, pool *pgxpool.Pool) *gin.Engine {
	r := gin.Default()

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

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/summary", func(c *gin.Context) {
			c.JSON(http.StatusNotImplemented, gin.H{
				"message": "summary endpoint will be implemented in later steps",
			})
		})
	}

	registerStatic(r)

	return r
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
