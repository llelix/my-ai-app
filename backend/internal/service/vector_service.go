package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ai-knowledge-app/internal/config"
	"github.com/pgvector/pgvector-go"
)

// VectorService 向量服务接口
type VectorService interface {
	GenerateEmbedding(ctx context.Context, text string) (pgvector.Vector, error)
}

// OpenAIVectorService OpenAI向量服务
type OpenAIVectorService struct {
	config *config.AIConfig
	client *http.Client
}

// EmbeddingRequest OpenAI embedding请求
type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// EmbeddingResponse OpenAI embedding响应
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewVectorService 创建向量服务
func NewVectorService(cfg *config.AIConfig) VectorService {
	return &OpenAIVectorService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateEmbedding 生成文本的向量表示
func (s *OpenAIVectorService) GenerateEmbedding(ctx context.Context, text string) (pgvector.Vector, error) {
	if text == "" {
		return pgvector.NewVector(nil), fmt.Errorf("input text cannot be empty")
	}

	model := "text-embedding-ada-002" // 默认模型

	reqBody := EmbeddingRequest{
		Input: text,
		Model: model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return pgvector.NewVector(nil), fmt.Errorf("failed to marshal request body: %w", err)
	}

	baseURL := s.config.OpenAI.BaseURL
	apiKey := s.config.OpenAI.APIKey

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return pgvector.NewVector(nil), fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return pgvector.NewVector(nil), fmt.Errorf("failed to send request to embedding API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return pgvector.NewVector(nil), fmt.Errorf("embedding API request failed with status %d", resp.StatusCode)
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return pgvector.NewVector(nil), fmt.Errorf("failed to decode embedding response: %w", err)
	}

	if len(embeddingResp.Data) == 0 || len(embeddingResp.Data[0].Embedding) == 0 {
		return pgvector.NewVector(nil), fmt.Errorf("no embedding data returned")
	}

	return pgvector.NewVector(embeddingResp.Data[0].Embedding), nil
}
