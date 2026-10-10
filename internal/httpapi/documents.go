package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"codex-hubcoding-test/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type document struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type createDocumentRequest struct {
	Title   string  `json:"title" binding:"required"`
	Content *string `json:"content" binding:"required"`
}

type documentRepository interface {
	Create(context.Context, *model.Document) error
	GetByID(context.Context, uint64) (model.Document, error)
}

type documentHandler struct {
	repository documentRepository
}

func (handler *documentHandler) create(c *gin.Context) {
	var request createDocumentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	title := strings.TrimSpace(request.Title)
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must not be blank"})
		return
	}

	stored := model.Document{
		Title:   title,
		Content: *request.Content,
	}
	if err := handler.repository.Create(c.Request.Context(), &stored); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save document"})
		return
	}

	c.JSON(http.StatusCreated, toDocumentResponse(stored))
}

func (handler *documentHandler) get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}
	stored, err := handler.repository.GetByID(c.Request.Context(), id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load document"})
		return
	}
	c.JSON(http.StatusOK, toDocumentResponse(stored))
}

func toDocumentResponse(stored model.Document) document {
	return document{
		ID:      strconv.FormatUint(stored.ID, 10),
		Title:   stored.Title,
		Content: stored.Content,
	}
}
