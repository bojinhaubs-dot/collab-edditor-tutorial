package repository

import (
	"context"

	"codex-hubcoding-test/internal/model"
	"gorm.io/gorm"
)

type Documents struct {
	db *gorm.DB
}

func NewDocuments(db *gorm.DB) *Documents {
	return &Documents{db: db}
}

func (r *Documents) Create(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *Documents) GetByID(ctx context.Context, id uint64) (model.Document, error) {
	var doc model.Document
	err := r.db.WithContext(ctx).First(&doc, id).Error
	return doc, err
}
