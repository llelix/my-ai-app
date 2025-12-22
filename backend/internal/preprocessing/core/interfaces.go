package core

import (
	"context"
)

// DocumentPreprocessingService 文档预处理服务接口
type DocumentPreprocessingService interface {
	ProcessDocument(ctx context.Context, documentID string) error
	GetProcessingStatus(ctx context.Context, documentID string) (*ProcessingStatus, error)
	BatchProcessDocuments(ctx context.Context, documentIDs []string) error

	// 异步处理方法
	ProcessDocumentAsync(documentID string, priority int) (*ProcessingTask, error)
	BatchProcessDocumentsAsync(documentIDs []string, priority int) ([]*ProcessingTask, error)
	GetTaskStatus(taskID string) (*ProcessingTask, error)
	GetTaskByDocumentID(documentID string) (*ProcessingTask, error)
	CancelTask(taskID string) error
	GetQueueStats() map[string]any

	// 扩展方法
	ReprocessDocument(ctx context.Context, documentID string) error
	GetDocumentChunks(ctx context.Context, documentID string) ([]DocumentChunk, error)
	GetChunkCount(ctx context.Context, documentID string) (int, error)
	DeleteDocumentData(ctx context.Context, documentID string) error
	GetProcessingStatistics(ctx context.Context) (map[string]any, error)
	ValidateDocumentForProcessing(ctx context.Context, documentID string) error
	GetSupportedFormats() []string
}

// DocumentChunkRepository 文档块存储库接口
type DocumentChunkRepository interface {
	// Create 创建文档块
	Create(ctx context.Context, chunk *DocumentChunk) error

	// CreateBatch 批量创建文档块
	CreateBatch(ctx context.Context, chunks []DocumentChunk) error

	// GetByDocumentID 根据文档ID获取所有块
	GetByDocumentID(ctx context.Context, documentID string) ([]DocumentChunk, error)

	// GetByID 根据ID获取单个块
	GetByID(ctx context.Context, id string) (*DocumentChunk, error)

	// Update 更新文档块
	Update(ctx context.Context, chunk *DocumentChunk) error

	// DeleteByDocumentID 删除文档的所有块
	DeleteByDocumentID(ctx context.Context, documentID string) error

	// DeleteByID 删除单个块
	DeleteByID(ctx context.Context, id string) error

	// GetChunkCount 获取文档的块数量
	GetChunkCount(ctx context.Context, documentID string) (int, error)
}

// ProcessingStatusRepository 处理状态存储库接口
type ProcessingStatusRepository interface {
	// Create 创建处理状态
	Create(ctx context.Context, status *ProcessingStatus) error

	// GetByDocumentID 根据文档ID获取处理状态
	GetByDocumentID(ctx context.Context, documentID string) (*ProcessingStatus, error)

	// Update 更新处理状态
	Update(ctx context.Context, status *ProcessingStatus) error

	// Delete 删除处理状态
	Delete(ctx context.Context, documentID string) error

	// GetPendingDocuments 获取待处理的文档列表
	GetPendingDocuments(ctx context.Context, limit int) ([]string, error)
}

// QualityValidator 质量验证器接口
type QualityValidator interface {
	// ValidateMarkdown 验证Markdown格式
	ValidateMarkdown(ctx context.Context, content string) (*ValidationResult, error)

	// ValidateChunks 验证文档块
	ValidateChunks(ctx context.Context, chunks []DocumentChunk) (*ValidationResult, error)
}
