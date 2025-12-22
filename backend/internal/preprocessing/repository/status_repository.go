package repository

import (
	"context"

	"ai-knowledge-app/internal/preprocessing/core"

	"gorm.io/gorm"
)

// StatusRepository 处理状态存储库实现
type StatusRepository struct {
	db *gorm.DB
}

// NewProcessingStatusRepository 创建处理状态存储库
func NewProcessingStatusRepository(db *gorm.DB) core.ProcessingStatusRepository {
	return &StatusRepository{db: db}
}

// Create 创建处理状态
func (r *StatusRepository) Create(ctx context.Context, status *core.ProcessingStatus) error {
	model := &DocumentProcessingStatusModel{}
	model.FromProcessingStatus(status)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByDocumentID 根据文档ID获取处理状态
func (r *StatusRepository) GetByDocumentID(ctx context.Context, documentID string) (*core.ProcessingStatus, error) {
	var model DocumentProcessingStatusModel
	err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		First(&model).Error
	if err != nil {
		return nil, err
	}

	return model.ToProcessingStatus(), nil
}

// Update 更新处理状态
func (r *StatusRepository) Update(ctx context.Context, status *core.ProcessingStatus) error {
	model := &DocumentProcessingStatusModel{}
	model.FromProcessingStatus(status)
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete 删除处理状态
func (r *StatusRepository) Delete(ctx context.Context, documentID string) error {
	return r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Delete(&DocumentProcessingStatusModel{}).Error
}

// GetPendingDocuments 获取待处理的文档列表
func (r *StatusRepository) GetPendingDocuments(ctx context.Context, limit int) ([]string, error) {
	var documentIDs []string
	err := r.db.WithContext(ctx).
		Model(&DocumentProcessingStatusModel{}).
		Where("preprocess_status = ?", core.StatusPending).
		Order("created_at ASC").
		Limit(limit).
		Pluck("document_id", &documentIDs).Error
	return documentIDs, err
}
