package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
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

type documentStore struct {
	mu      sync.Mutex
	nextID  uint64
	entries map[string]document
}

func newDocumentStore() *documentStore {
	return &documentStore{entries: make(map[string]document)}
}

func (store *documentStore) create(c *gin.Context) {
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

	// The ID and map entry change together so concurrent requests cannot reuse an ID.
	store.mu.Lock()
	store.nextID++
	result := document{
		ID:      strconv.FormatUint(store.nextID, 10),
		Title:   title,
		Content: *request.Content,
	}
	store.entries[result.ID] = result
	store.mu.Unlock()

	c.JSON(http.StatusCreated, result)
}

func (store *documentStore) get(c *gin.Context) {
	store.mu.Lock()
	result, ok := store.entries[c.Param("id")]
	store.mu.Unlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}
