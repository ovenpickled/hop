package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ovenpickled/hop/handler"
)

// SetupRouter registers all routes and returns a configured gin Engine.
func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to the URL Shortener API",
		})
	})

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	r.POST("/create-short-url", handler.CreateShortUrl)
	r.GET("/:shortUrl", handler.HandleShortUrlRedirect)

	return r
}
