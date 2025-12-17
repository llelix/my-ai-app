package ai

import (
	"context"
	"testing"

	"ai-knowledge-app/internal/config"
	"ai-knowledge-app/pkg/logger"
)

func TestGetModels(t *testing.T) {
	// 初始化测试日志
	logConfig := &config.LogConfig{
		Level:  "info",
		Format: "text",
	}
	if err := logger.InitLogger(logConfig); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// 创建测试配置
	testConfig := &config.AIConfig{
		Provider: "openai",
		OpenAI: config.OpenAIConfig{
			APIKey:  "sk-Ee16GOiSAepaEfdC0jZmwiHphZ67RwPygS7Cd3ZmPGJ5NlI7c",
			BaseURL: "https://api.chatanywhere.tech/",
			Model:   "deepseek-r1",
		},
	}

	// 创建AI服务实例
	service := NewAIService(testConfig).(*OpenAIService)

	// 测试GetModels方法
	models := service.GetModels()

	// 验证返回的模型列表不为空
	if len(models) == 0 {
		t.Error("GetModels() returned empty model list")
	}

	// 验证至少包含一些常见的模型
	expectedModels := []string{"gpt-3.5-turbo", "gpt-4", "deepseek-r1"}
	foundExpectedModel := false

	for _, expected := range expectedModels {
		for _, model := range models {
			if model == expected {
				foundExpectedModel = true
				break
			}
		}
		if foundExpectedModel {
			break
		}
	}

	if !foundExpectedModel {
		t.Errorf("GetModels() did not contain any expected models. Got: %v", models)
	}

	t.Logf("Available models: %v", models)
}

func TestGetModelsWithInvalidConfig(t *testing.T) {
	// 创建无效配置测试
	testConfig := &config.AIConfig{
		Provider: "openai",
		OpenAI: config.OpenAIConfig{
			APIKey:  "invalid-key",
			BaseURL: "https://invalid-url.com/",
			Model:   "test-model",
		},
	}

	// 创建AI服务实例
	service := NewAIService(testConfig).(*OpenAIService)

	// 测试GetModels方法，应该返回默认模型
	models := service.GetModels()

	// 验证即使配置无效，也返回默认模型
	if len(models) == 0 {
		t.Error("GetModels() with invalid config should return default models")
	}

	t.Logf("Default models for invalid config: %v", models)
}

func TestQueryIntegration(t *testing.T) {
	// 跳过集成测试，除非明确启用
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 创建测试配置
	testConfig := &config.AIConfig{
		Provider: "openai",
		OpenAI: config.OpenAIConfig{
			APIKey:  "sk-Ee16GOiSAepaEfdC0jZmwiHphZ67RwPygS7Cd3ZmPGJ5NlI7c",
			BaseURL: "https://api.chatanywhere.tech/",
			Model:   "deepseek-r1",
		},
	}

	// 创建AI服务实例
	service := NewAIService(testConfig)

	// 测试查询功能
	ctx := context.Background()
	req := QueryRequest{
		Query:       "Hello, how are you?",
		Model:       "deepseek-r1",
		Temperature: 0.7,
		MaxTokens:   100,
	}

	resp, err := service.Query(ctx, req)
	if err != nil {
		t.Errorf("Query() failed: %v", err)
		return
	}

	if resp.Response == "" {
		t.Error("Query() returned empty response")
	}

	if resp.Model == "" {
		t.Error("Query() returned empty model name")
	}

	t.Logf("Query response: %s", resp.Response)
	t.Logf("Used model: %s", resp.Model)
	t.Logf("Token count: %d", resp.Tokens)
	t.Logf("Duration: %v", resp.Duration)
}