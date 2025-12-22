package preprocessing

import (
	"ai-knowledge-app/internal/preprocessing/core"
	"ai-knowledge-app/internal/preprocessing/repository"
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Service 预处理服务主入口
type Service struct {
	*ServiceImpl
}

// ServiceImpl 文档预处理服务实现
type ServiceImpl struct {
	chunkRepo      core.DocumentChunkRepository
	statusRepo     core.ProcessingStatusRepository
	cascadeManager *repository.CascadeDeleteManager
	db             *gorm.DB
}

// NewService 创建新的文档预处理服务
func NewService(db *gorm.DB) *Service {
	service := &ServiceImpl{
		chunkRepo:      repository.NewDocumentChunkRepository(db),
		statusRepo:     repository.NewProcessingStatusRepository(db),
		cascadeManager: repository.NewCascadeDeleteManager(db),
		db:             db,
	}

	return &Service{ServiceImpl: service}
}

// GetSupportedFormats 获取支持的文档格式
func (s *ServiceImpl) GetSupportedFormats() []string {
	return []string{"pdf", "docx", "doc", "txt", "md"}
}

// ProcessDocument 处理文档
func (s *ServiceImpl) ProcessDocument(ctx context.Context, documentID string) error {
	// 1. 获取文档信息（这里需要从documents表查询）
	// 注意：这里需要添加document repository来查询文档信息
	// 暂时使用模拟数据进行演示

	// 2. 模拟文档内容（实际应该从文件系统或存储中读取）
	documentContent := `这是一个示例文档的内容。

这个文档包含多个段落，用于演示文档分块功能。

第一段：介绍了文档的基本信息和用途。

第二段：详细说明了文档处理的流程和步骤。

第三段：展示了如何将长文档分割成更小的、可管理的块。

第四段：每个块都包含相关的元数据信息，如索引、偏移量等。

第五段：这样的分块方式有助于后续的向量化和语义搜索。`

	// 3. 执行文档分块
	chunks := s.chunkDocument(documentContent, documentID)

	// 4. 保存chunks到数据库
	if err := s.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		return fmt.Errorf("failed to save chunks: %w", err)
	}

	// 5. 更新处理状态
	status := &core.ProcessingStatus{
		DocumentID:       documentID,
		PreprocessStatus: core.StatusCompleted,
		Progress:         100.0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		CompletedAt:      &[]time.Time{time.Now()}[0],
	}

	// 尝试创建状态记录，如果已存在则更新
	if err := s.statusRepo.Create(ctx, status); err != nil {
		// 如果创建失败，可能是因为记录已存在，尝试更新
		if updateErr := s.statusRepo.Update(ctx, status); updateErr != nil {
			return fmt.Errorf("failed to create or update status: create error: %w, update error: %v", err, updateErr)
		}
	}

	return nil
}

// chunkDocument 将文档内容分块
func (s *ServiceImpl) chunkDocument(content, documentID string) []core.DocumentChunk {
	// 简单的分块逻辑：按段落分割
	paragraphs := strings.Split(content, "\n\n")
	chunks := make([]core.DocumentChunk, 0, len(paragraphs))

	offset := 0
	for i, paragraph := range paragraphs {
		if strings.TrimSpace(paragraph) == "" {
			continue
		}

		chunk := core.DocumentChunk{
			ID:          core.GenerateID(),
			DocumentID:  documentID,
			Content:     strings.TrimSpace(paragraph),
			ChunkIndex:  i,
			StartOffset: offset,
			EndOffset:   offset + len(paragraph),
			Metadata: map[string]any{
				"type":       "paragraph",
				"length":     len(paragraph),
				"word_count": len(strings.Fields(paragraph)),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		chunks = append(chunks, chunk)
		offset += len(paragraph) + 2 // +2 for \n\n
	}

	return chunks
}

// GetProcessingStatus 获取处理状态
func (s *ServiceImpl) GetProcessingStatus(ctx context.Context, documentID string) (*core.ProcessingStatus, error) {
	// 从数据库查询真实的处理状态
	status, err := s.statusRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		// 如果没有找到状态记录，返回默认的未开始状态
		if err == gorm.ErrRecordNotFound {
			return &core.ProcessingStatus{
				ID:                    core.GenerateID(),
				DocumentID:            documentID,
				PreprocessStatus:      core.StatusPending,
				VectorizationStatus:   core.StatusNotStarted,
				Progress:              0.0,
				VectorizationProgress: 0.0,
				Error:                 "",
				VectorizationError:    "",
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
				ProcessingTime:        0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get processing status: %w", err)
	}

	return status, nil
}

// BatchProcessDocuments 批量处理文档
func (s *ServiceImpl) BatchProcessDocuments(ctx context.Context, documentIDs []string) error {
	// TODO: 实现批量处理逻辑
	return nil
}

// ProcessDocumentAsync 异步处理文档
func (s *ServiceImpl) ProcessDocumentAsync(documentID string, priority int) (*core.ProcessingTask, error) {
	// 创建一个模拟的处理任务
	task := &core.ProcessingTask{
		ID:         core.GenerateID(),
		DocumentID: documentID,
		Status:     core.StatusProcessing,
		Priority:   priority,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Error:      "",
	}

	return task, nil
}

// BatchProcessDocumentsAsync 异步批量处理文档
func (s *ServiceImpl) BatchProcessDocumentsAsync(documentIDs []string, priority int) ([]*core.ProcessingTask, error) {
	// TODO: 实现异步批量处理逻辑
	return nil, nil
}

// GetTaskStatus 获取任务状态
func (s *ServiceImpl) GetTaskStatus(taskID string) (*core.ProcessingTask, error) {
	// 创建一个模拟的任务状态
	task := &core.ProcessingTask{
		ID:         taskID,
		DocumentID: "1", // 模拟文档ID
		Status:     core.StatusCompleted,
		Priority:   1,
		CreatedAt:  time.Now().Add(-10 * time.Minute),
		UpdatedAt:  time.Now(),
		Error:      "",
	}

	return task, nil
}

// GetTaskByDocumentID 根据文档ID获取任务
func (s *ServiceImpl) GetTaskByDocumentID(documentID string) (*core.ProcessingTask, error) {
	// TODO: 实现根据文档ID获取任务逻辑
	return nil, nil
}

// CancelTask 取消任务
func (s *ServiceImpl) CancelTask(taskID string) error {
	// TODO: 实现取消任务逻辑
	return nil
}

// GetQueueStats 获取队列统计
func (s *ServiceImpl) GetQueueStats() map[string]any {
	// TODO: 实现获取队列统计逻辑
	return map[string]any{
		"pending_tasks":    0,
		"processing_tasks": 0,
		"completed_tasks":  0,
		"failed_tasks":     0,
	}
}

// ReprocessDocument 重新处理文档
func (s *ServiceImpl) ReprocessDocument(ctx context.Context, documentID string) error {
	// TODO: 实现重新处理文档逻辑
	return nil
}

// GetDocumentChunks 获取文档分块
func (s *ServiceImpl) GetDocumentChunks(ctx context.Context, documentID string) ([]core.DocumentChunk, error) {
	// 从数据库中查询真实的chunk数据
	return s.chunkRepo.GetByDocumentID(ctx, documentID)
}

// GetChunkCount 获取分块数量
func (s *ServiceImpl) GetChunkCount(ctx context.Context, documentID string) (int, error) {
	// TODO: 实现获取分块数量逻辑
	return 0, nil
}

// DeleteDocumentData 删除文档数据
func (s *ServiceImpl) DeleteDocumentData(ctx context.Context, documentID string) error {
	// TODO: 实现删除文档数据逻辑
	return nil
}

// GetProcessingStatistics 获取处理统计
func (s *ServiceImpl) GetProcessingStatistics(ctx context.Context) (map[string]any, error) {
	// TODO: 实现获取处理统计逻辑
	return map[string]any{
		"total_documents":     0,
		"processed_documents": 0,
		"failed_documents":    0,
		"processing_rate":     0.0,
	}, nil
}

// ValidateDocumentForProcessing 验证文档是否可以处理
func (s *ServiceImpl) ValidateDocumentForProcessing(ctx context.Context, documentID string) error {
	// TODO: 实现验证文档逻辑
	return nil
}
