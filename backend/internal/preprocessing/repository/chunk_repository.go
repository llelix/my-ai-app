package repository

import (
	"context"

	"ai-knowledge-app/internal/preprocessing/core"

	"gorm.io/gorm"
)

// ChunkRepository 文档块存储库实现
type ChunkRepository struct {
	db *gorm.DB
}

// NewDocumentChunkRepository 创建文档块存储库
func NewDocumentChunkRepository(db *gorm.DB) core.DocumentChunkRepository {
	return &ChunkRepository{db: db}
}

// Create 创建文档块
func (r *ChunkRepository) Create(ctx context.Context, chunk *core.DocumentChunk) error {
	model := &DocumentChunkModel{}
	model.FromDocumentChunk(chunk)
	return r.db.WithContext(ctx).Create(model).Error
}

// CreateBatch 批量创建文档块
func (r *ChunkRepository) CreateBatch(ctx context.Context, chunks []core.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	models := make([]DocumentChunkModel, len(chunks))
	for i, chunk := range chunks {
		models[i].FromDocumentChunk(&chunk)
	}

	return r.db.WithContext(ctx).CreateInBatches(models, 100).Error
}

// GetByDocumentID 根据文档ID获取所有块
func (r *ChunkRepository) GetByDocumentID(ctx context.Context, documentID string) ([]core.DocumentChunk, error) {
	var models []DocumentChunkModel
	err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("chunk_index ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	chunks := make([]core.DocumentChunk, len(models))
	for i, model := range models {
		chunks[i] = *model.ToDocumentChunk()
	}

	return chunks, nil
}

// GetByID 根据ID获取单个块
func (r *ChunkRepository) GetByID(ctx context.Context, id string) (*core.DocumentChunk, error) {
	var model DocumentChunkModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		return nil, err
	}

	return model.ToDocumentChunk(), nil
}

// Update 更新文档块
func (r *ChunkRepository) Update(ctx context.Context, chunk *core.DocumentChunk) error {
	model := &DocumentChunkModel{}
	model.FromDocumentChunk(chunk)
	return r.db.WithContext(ctx).Save(model).Error
}

// DeleteByDocumentID 删除文档的所有块
func (r *ChunkRepository) DeleteByDocumentID(ctx context.Context, documentID string) error {
	return r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Delete(&DocumentChunkModel{}).Error
}

// DeleteByID 删除单个块
func (r *ChunkRepository) DeleteByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&DocumentChunkModel{}).Error
}

// GetChunkCount 获取文档的块数量
func (r *ChunkRepository) GetChunkCount(ctx context.Context, documentID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&DocumentChunkModel{}).
		Where("document_id = ?", documentID).
		Count(&count).Error
	return int(count), err
}
