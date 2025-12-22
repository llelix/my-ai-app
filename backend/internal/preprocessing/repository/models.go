package repository

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"ai-knowledge-app/internal/preprocessing/core"
)

// DocumentChunkModel 文档块数据库模型
type DocumentChunkModel struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	DocumentID  string    `gorm:"type:varchar(36);not null;index" json:"document_id"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	ChunkIndex  int       `gorm:"not null;index" json:"chunk_index"`
	StartOffset int       `gorm:"not null" json:"start_offset"`
	EndOffset   int       `gorm:"not null" json:"end_offset"`
	Metadata    string    `gorm:"type:text" json:"metadata"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (DocumentChunkModel) TableName() string {
	return "document_chunks"
}

// ToDocumentChunk 转换为业务模型
func (m *DocumentChunkModel) ToDocumentChunk() *core.DocumentChunk {
	// 解析metadata JSON字符串
	var metadata map[string]any
	if m.Metadata != "" {
		json.Unmarshal([]byte(m.Metadata), &metadata)
	}
	if metadata == nil {
		metadata = make(map[string]any)
	}

	return &core.DocumentChunk{
		ID:          m.ID,
		DocumentID:  m.DocumentID,
		Content:     m.Content,
		ChunkIndex:  m.ChunkIndex,
		StartOffset: m.StartOffset,
		EndOffset:   m.EndOffset,
		Metadata:    metadata,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// FromDocumentChunk 从业务模型创建
func (m *DocumentChunkModel) FromDocumentChunk(chunk *core.DocumentChunk) {
	m.ID = chunk.ID
	m.DocumentID = chunk.DocumentID
	m.Content = chunk.Content
	m.ChunkIndex = chunk.ChunkIndex
	m.StartOffset = chunk.StartOffset
	m.EndOffset = chunk.EndOffset

	// 序列化metadata为JSON字符串
	if chunk.Metadata != nil {
		if metadataBytes, err := json.Marshal(chunk.Metadata); err == nil {
			m.Metadata = string(metadataBytes)
		}
	}

	m.CreatedAt = chunk.CreatedAt
	m.UpdatedAt = chunk.UpdatedAt
}

// DocumentProcessingStatusModel 文档处理状态数据库模型
type DocumentProcessingStatusModel struct {
	ID                    string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	DocumentID            string     `gorm:"type:varchar(36);not null;uniqueIndex" json:"document_id"`
	PreprocessStatus      string     `gorm:"type:varchar(20);not null;index" json:"preprocess_status"`
	VectorizationStatus   string     `gorm:"type:varchar(20);not null;default:'not_started'" json:"vectorization_status"`
	Progress              float64    `gorm:"type:decimal(5,2);default:0.00" json:"progress"`
	VectorizationProgress float64    `gorm:"type:decimal(5,2);default:0.00" json:"vectorization_progress"`
	ErrorMessage          string     `gorm:"type:text" json:"error_message"`
	VectorizationError    string     `gorm:"type:text" json:"vectorization_error"`
	ProcessingOptions     string     `gorm:"type:text" json:"processing_options"`
	CreatedAt             time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	CompletedAt           *time.Time `json:"completed_at"`
}

// TableName 指定表名
func (DocumentProcessingStatusModel) TableName() string {
	return "document_processing_status"
}

// ToProcessingStatus 转换为业务模型
func (m *DocumentProcessingStatusModel) ToProcessingStatus() *core.ProcessingStatus {
	var processingTime time.Duration
	if m.CompletedAt != nil {
		processingTime = m.CompletedAt.Sub(m.CreatedAt)
	}

	return &core.ProcessingStatus{
		ID:                    m.ID,
		DocumentID:            m.DocumentID,
		PreprocessStatus:      core.ProcessingStatusType(m.PreprocessStatus),
		VectorizationStatus:   core.ProcessingStatusType(m.VectorizationStatus),
		Progress:              m.Progress,
		VectorizationProgress: m.VectorizationProgress,
		Error:                 m.ErrorMessage,
		VectorizationError:    m.VectorizationError,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
		CompletedAt:           m.CompletedAt,
		ProcessingTime:        processingTime,
	}
}

// FromProcessingStatus 从业务模型创建
func (m *DocumentProcessingStatusModel) FromProcessingStatus(status *core.ProcessingStatus) {
	if status.ID != "" {
		m.ID = status.ID
	} else if m.ID == "" {
		m.ID = core.GenerateID()
	}
	m.DocumentID = status.DocumentID
	m.PreprocessStatus = string(status.PreprocessStatus)
	m.VectorizationStatus = string(status.VectorizationStatus)
	m.Progress = status.Progress
	m.VectorizationProgress = status.VectorizationProgress
	m.ErrorMessage = status.Error
	m.VectorizationError = status.VectorizationError
	m.CreatedAt = status.CreatedAt
	m.UpdatedAt = status.UpdatedAt
	m.CompletedAt = status.CompletedAt
}

// DocumentEmbeddingModel 文档嵌入数据库模型（预留）
type DocumentEmbeddingModel struct {
	ID         string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ChunkID    string    `gorm:"type:varchar(36);not null;index" json:"chunk_id"`
	VectorData string    `gorm:"type:text" json:"vector_data"` // 简化为text类型
	ModelName  string    `gorm:"type:varchar(100);not null;index" json:"model_name"`
	Dimensions int       `gorm:"not null" json:"dimensions"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (DocumentEmbeddingModel) TableName() string {
	return "document_embeddings"
}

// BeforeCreate GORM钩子，创建前生成ID
func (m *DocumentChunkModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = core.GenerateID()
	}
	return nil
}

// BeforeCreate GORM钩子，创建前生成ID
func (m *DocumentProcessingStatusModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = core.GenerateID()
	}
	return nil
}

// BeforeCreate GORM钩子，创建前生成ID
func (m *DocumentEmbeddingModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = core.GenerateID()
	}
	return nil
}
