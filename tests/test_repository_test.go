package tests

import (
	"context"
	"sync"

	"codex-hubcoding-test/internal/model"
	"gorm.io/gorm"
)

// testRepository keeps the HTTP tests independent of an external MySQL instance.
type testRepository struct {
	mu      sync.Mutex
	nextID  uint64
	entries map[uint64]model.Document
}

func newTestRepository() *testRepository {
	return &testRepository{entries: make(map[uint64]model.Document)}
}

func (r *testRepository) Create(_ context.Context, doc *model.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	doc.ID = r.nextID
	r.entries[doc.ID] = *doc
	return nil
}

func (r *testRepository) GetByID(_ context.Context, id uint64) (model.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	doc, ok := r.entries[id]
	if !ok {
		return model.Document{}, gorm.ErrRecordNotFound
	}
	return doc, nil
}
