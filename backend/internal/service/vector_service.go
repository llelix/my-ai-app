package service

import (
	"context"
	"fmt"

	"ai-knowledge-app/internal/config"
	"github.com/pgvector/pgvector-go"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

// VectorService 向量服务接口
type VectorService interface {
	GenerateEmbedding(ctx context.Context, text string) (pgvector.Vector, error)
}

// OpenAIVectorService OpenAI向量服务
type OpenAIVectorService struct {
	config    *config.AIConfig
	embedder  embeddings.Embedder
}

// NewVectorService 创建向量服务
func NewVectorService(cfg *config.AIConfig) VectorService {
	// 创建OpenAI LLM客户端用于embeddings
	llm, err := openai.New(
		openai.WithModel("text-embedding-ada-002"),
		openai.WithBaseURL(cfg.OpenAI.BaseURL),
		openai.WithToken(cfg.OpenAI.APIKey),
	)
	if err != nil {
		// 如果创建失败，返回一个基本的实现
		return &OpenAIVectorService{
			config:   cfg,
			embedder: nil,
		}
	}

	// 创建embedder
	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return &OpenAIVectorService{
			config:   cfg,
			embedder: nil,
		}
	}

	return &OpenAIVectorService{
		config:   cfg,
		embedder: embedder,
	}
}

// GenerateEmbedding 生成文本的向量表示
func (s *OpenAIVectorService) GenerateEmbedding(ctx context.Context, text string) (pgvector.Vector, error) {
	if text == "" {
		return pgvector.NewVector(nil), fmt.Errorf("input text cannot be empty")
	}

	// 检查embedder是否已初始化
	if s.embedder == nil {
		// 尝试重新初始化embedder
		llm, err := openai.New(
			openai.WithModel("text-embedding-ada-002"),
			openai.WithBaseURL(s.config.OpenAI.BaseURL),
			openai.WithToken(s.config.OpenAI.APIKey),
		)
		if err != nil {
			return pgvector.NewVector(nil), fmt.Errorf("failed to initialize LLM: %w", err)
		}

		embedder, err := embeddings.NewEmbedder(llm)
		if err != nil {
			return pgvector.NewVector(nil), fmt.Errorf("failed to initialize embedder: %w", err)
		}
		s.embedder = embedder
	}

	// 使用LangChain-Go生成embedding
	vectors, err := s.embedder.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return pgvector.NewVector(nil), fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return pgvector.NewVector(nil), fmt.Errorf("no embedding data returned")
	}

	// pgvector.NewVector接受[]float32，所以直接使用
	return pgvector.NewVector(vectors[0]), nil
}
