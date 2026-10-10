package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(repository documentRepository) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	documents := &documentHandler{repository: repository}
	router.POST("/documents", documents.create)
	router.GET("/documents/:id", documents.get)
	return router
}
