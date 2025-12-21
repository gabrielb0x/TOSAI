package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed web/*
var embeddedWeb embed.FS

func registerStatic(r *gin.Engine) {
	content, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		log.Printf("failed to mount embedded web assets: %v", err)
		return
	}

	r.StaticFS("/web", http.FS(content))
	r.GET("/", func(c *gin.Context) {
		c.FileFromFS("test.html", http.FS(content))
	})
}
