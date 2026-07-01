package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ovenpickled/hop/config"
	"github.com/ovenpickled/hop/shortener"
	"github.com/ovenpickled/hop/store"
)

// Request model definition
type UrlCreationRequest struct {
	LongUrl string `json:"long_url" binding:"required"`
	UserId  string `json:"user_id" binding:"required"`
}

// cfg is set once at startup via Init so handlers can read config (like BaseURL) without needing it threaded through every function signature.
var cfg config.Config

func Init(c config.Config) {
	cfg = c
}

func CreateShortUrl(c *gin.Context) {
	var creationRequest UrlCreationRequest
	if err := c.ShouldBindJSON(&creationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortUrl := shortener.GenerateShortLink(creationRequest.LongUrl, creationRequest.UserId)

	if err := store.SaveUrlMapping(shortUrl, creationRequest.LongUrl, creationRequest.UserId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save short url"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "short url created successfully",
		"short_url": cfg.BaseURL + shortUrl,
	})
}

func HandleShortUrlRedirect(c *gin.Context) {
	shortUrl := c.Param("shortUrl")

	initialUrl, err := store.RetrieveInitialUrl(shortUrl)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "short url not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve short url"})
		return
	}

	c.Redirect(http.StatusFound, initialUrl)
}
