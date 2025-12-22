package repository

import (
	"context"

	"gorm.io/gorm"
)

// CascadeDeleteManager 级联删除管理器
type CascadeDeleteManager struct {
	db *gorm.DB
}

// NewCascadeDeleteManager 创建级联删除管理器
func NewCascadeDeleteManager(db *gorm.DB) *CascadeDeleteManager {
	return &CascadeDeleteManager{db: db}
}

// DeleteDocumentData 删除文档相关的所有数据
func (m *CascadeDeleteManager) DeleteDocumentData(ctx context.Context, documentID string) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除文档块
		if err := tx.Where("document_id = ?", documentID).Delete(&DocumentChunkModel{}).Error; err != nil {
			return err
		}

		// 删除处理状态
		if err := tx.Where("document_id = ?", documentID).Delete(&DocumentProcessingStatusModel{}).Error; err != nil {
			return err
		}

		// 删除嵌入向量（如果存在）
		if err := tx.Where("chunk_id IN (SELECT id FROM document_chunks WHERE document_id = ?)", documentID).Delete(&DocumentEmbeddingModel{}).Error; err != nil {
			return err
		}

		return nil
	})
}
