package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/auto-tos/auto-tos/internal/config"
)

// New sets up the Gin engine with routes and middleware.
func New(cfg config.Config) *gin.Engine {
	r := gin.Default()

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

	return r
}
