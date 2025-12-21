package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/auto-tos/auto-tos/internal/config"
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
	if cfg.AllowedOrigins == "*" {
		corsCfg.AllowAllOrigins = true
	} else {
		corsCfg.AllowOrigins = []string{cfg.AllowedOrigins}
	}
	corsCfg.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
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
