package core

import (
	"time"
)

// ProcessingStatusType 处理状态类型
type ProcessingStatusType string

const (
	StatusPending    ProcessingStatusType = "pending"
	StatusProcessing ProcessingStatusType = "processing"
	StatusCompleted  ProcessingStatusType = "completed"
	StatusFailed     ProcessingStatusType = "failed"
	StatusNotStarted ProcessingStatusType = "not_started" // 用于向量化预留状态
)

// ProcessingStatus 处理状态
type ProcessingStatus struct {
	ID                    string               `json:"id"`
	DocumentID            string               `json:"document_id"`
	PreprocessStatus      ProcessingStatusType `json:"preprocess_status"`
	VectorizationStatus   ProcessingStatusType `json:"vectorization_status"` // 预留字段
	Progress              float64              `json:"progress"`
	VectorizationProgress float64              `json:"vectorization_progress"` // 预留字段
	Error                 string               `json:"error,omitempty"`
	VectorizationError    string               `json:"vectorization_error,omitempty"` // 预留字段
	CreatedAt             time.Time            `json:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at"`
	CompletedAt           *time.Time           `json:"completed_at,omitempty"`
	ProcessingTime        time.Duration        `json:"processing_time"`
}

// DocumentChunk 文档块
type DocumentChunk struct {
	ID          string         `json:"id" db:"id"`
	DocumentID  string         `json:"document_id" db:"document_id"`
	Content     string         `json:"content" db:"content"`
	ChunkIndex  int            `json:"chunk_index" db:"chunk_index"`
	StartOffset int            `json:"start_offset" db:"start_offset"`
	EndOffset   int            `json:"end_offset" db:"end_offset"`
	Metadata    map[string]any `json:"metadata" db:"metadata"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// ConversionOptions 转换选项
type ConversionOptions struct {
	Language      string `json:"language"`
	Backend       string `json:"backend"`      // pipeline, vlm-transformers, vlm-vllm-engine
	ParseMethod   string `json:"parse_method"` // auto, txt, ocr
	FormulaEnable bool   `json:"formula_enable"`
	TableEnable   bool   `json:"table_enable"`
	ExtractImages bool   `json:"extract_images"`
}

// MarkdownResult 转换结果
type MarkdownResult struct {
	Content     string         `json:"content"`
	Images      []ImageInfo    `json:"images"`
	Metadata    map[string]any `json:"metadata"`
	ProcessTime time.Duration  `json:"process_time"`
}

// ImageInfo 图像信息
type ImageInfo struct {
	Path        string    `json:"path"`
	Caption     string    `json:"caption"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Format      string    `json:"format"`
	Size        int64     `json:"size"`
	Reference   string    `json:"reference"`
	ExtractedAt time.Time `json:"extracted_at"`
}

// ChunkingOptions 分块选项
type ChunkingOptions struct {
	ChunkSize    int      `json:"chunk_size"`
	ChunkOverlap int      `json:"chunk_overlap"`
	Separators   []string `json:"separators"`
}

// EmbeddingOptions 向量化选项（预留）
type EmbeddingOptions struct {
	Model      string `json:"model"`      // 嵌入模型名称
	BatchSize  int    `json:"batch_size"` // 批处理大小
	Normalize  bool   `json:"normalize"`  // 是否标准化向量
	Dimensions int    `json:"dimensions"` // 向量维度（可选）
}

// DocumentEmbedding 文档嵌入（预留）
type DocumentEmbedding struct {
	ChunkID    string    `json:"chunk_id"`
	Vector     []float32 `json:"vector"`
	Model      string    `json:"model"`
	Dimensions int       `json:"dimensions"`
	CreatedAt  time.Time `json:"created_at"`
}

// VectorizationConfig 向量化配置（预留）
type VectorizationConfig struct {
	Enabled    bool             `json:"enabled"`
	Model      string           `json:"model"`
	BatchSize  int              `json:"batch_size"`
	Dimensions int              `json:"dimensions"`
	Options    EmbeddingOptions `json:"options"`
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Score    float64  `json:"score"`
	Issues   []string `json:"issues"`
	Warnings []string `json:"warnings"`
}

// ProcessingTask 处理任务
type ProcessingTask struct {
	ID         string               `json:"id"`
	DocumentID string               `json:"document_id"`
	Status     ProcessingStatusType `json:"status"`
	Priority   int                  `json:"priority"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
	Error      string               `json:"error,omitempty"`
}
