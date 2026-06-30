package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ovenpickled/hop/handler"
)

// SetupRouter registers all routes and returns a configured gin engine
func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the URL Shortener API",
		})
	})

	r. POST("/create-short-url", handler.CreateShortUrl)
	r.GET("/:shortUrl", handler.HandleShortUrlRedirect)

	return r
}
