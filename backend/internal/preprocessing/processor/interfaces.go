package processor

import (
	"context"

	"ai-knowledge-app/internal/preprocessing/core"
)

// MinerUProcessor MinerU文档转换处理器
type MinerUProcessor interface {
	ConvertToMarkdown(ctx context.Context, filePath string, options *core.ConversionOptions) (*core.MarkdownResult, error)
	SupportedFormats() []string
}

// TextChunker 文本分块器接口
type TextChunker interface {
	ChunkText(ctx context.Context, text string, options *core.ChunkingOptions) ([]core.DocumentChunk, error)
}

// VectorizationProcessor 向量化处理器接口（预留扩展）
type VectorizationProcessor interface {
	// GenerateEmbeddings 为文档块生成向量嵌入
	GenerateEmbeddings(ctx context.Context, chunks []core.DocumentChunk, options *core.EmbeddingOptions) ([]core.DocumentEmbedding, error)

	// BatchGenerateEmbeddings 批量生成向量嵌入
	BatchGenerateEmbeddings(ctx context.Context, chunkBatches [][]core.DocumentChunk, options *core.EmbeddingOptions) ([][]core.DocumentEmbedding, error)

	// GetEmbeddingDimensions 获取嵌入向量的维度
	GetEmbeddingDimensions() int

	// SupportedModels 获取支持的嵌入模型列表
	SupportedModels() []string
}
