package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ai-knowledge-app/internal/config"
	"ai-knowledge-app/internal/models"
	"ai-knowledge-app/internal/service"
	"ai-knowledge-app/pkg/database"
	"ai-knowledge-app/pkg/logger"

	"github.com/pgvector/pgvector-go"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"gorm.io/gorm"
)

// AIService AI服务接口
type AIService interface {
	Query(ctx context.Context, req QueryRequest) (*QueryResponse, error)
	GetModels() []string
	SetVectorService(vectorService service.VectorService)
}

// OpenAIService OpenAI兼容的AI服务
type OpenAIService struct {
	config        *config.AIConfig
	llm           llms.Model
	vectorService service.VectorService
}

// QueryRequest AI查询请求
type QueryRequest struct {
	Query       string   `json:"query"`
	Model       string   `json:"model"`
	Temperature float64  `json:"temperature"`
	MaxTokens   int      `json:"max_tokens"`
	Context     []string `json:"context,omitempty"`
}

// QueryResponse AI查询响应
type QueryResponse struct {
	Response     string        `json:"response"`
	Model        string        `json:"model"`
	Tokens       int           `json:"tokens"`
	Duration     time.Duration `json:"duration"`
	KnowledgeIDs []uint        `json:"knowledge_ids,omitempty"`
	RelevantDocs []string      `json:"relevant_docs,omitempty"`
}

// NewAIService 创建AI服务实例
func NewAIService(cfg *config.AIConfig) AIService {
	// 创建LangChain-Go OpenAI LLM实例
	llm, err := openai.New(
		openai.WithModel(cfg.OpenAI.Model),
		openai.WithBaseURL(cfg.OpenAI.BaseURL),
		openai.WithToken(cfg.OpenAI.APIKey),
	)
	if err != nil {
		logger.GetLogger().WithError(err).Error("Failed to create OpenAI LLM")
		// 返回一个基本的实例，后续可以重试
		return &OpenAIService{
			config: cfg,
			llm:    nil,
		}
	}

	return &OpenAIService{
		config: cfg,
		llm:    llm,
	}
}

// SetVectorService 设置向量服务
func (s *OpenAIService) SetVectorService(vectorService service.VectorService) {
	s.vectorService = vectorService
}

// Query 执行AI查询
func (s *OpenAIService) Query(ctx context.Context, req QueryRequest) (*QueryResponse, error) {
	startTime := time.Now()

	// 检查LLM是否已初始化
	if s.llm == nil {
		// 尝试重新初始化LLM
		llm, err := openai.New(
			openai.WithModel(s.config.OpenAI.Model),
			openai.WithBaseURL(s.config.OpenAI.BaseURL),
			openai.WithToken(s.config.OpenAI.APIKey),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize LLM: %w", err)
		}
		s.llm = llm
	}

	// 获取相关的知识库内容
	relevantDocs, knowledgeIDs, err := s.searchRelevantKnowledge(ctx, req.Query)
	if err != nil {
		logger.GetLogger().WithError(err).Error("Failed to search relevant knowledge")
		// 继续执行，不要因为向量搜索失败而终止整个查询
	}

	// 构建系统提示
	systemPrompt := s.buildSystemPrompt(relevantDocs)

	// 使用LangChain-Go的提示模板
	promptTemplate := prompts.NewPromptTemplate(
		systemPrompt,
		[]string{"query"},
	)

	// 格式化提示
	formattedPrompt, err := promptTemplate.Format(map[string]any{
		"query": req.Query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to format prompt: %w", err)
	}

	// 使用LangChain-Go生成响应
	var response string
	if req.Temperature > 0 || req.MaxTokens > 0 {
		// 使用自定义选项
		options := []llms.CallOption{
			llms.WithTemperature(req.Temperature),
		}
		if req.MaxTokens > 0 {
			options = append(options, llms.WithMaxTokens(req.MaxTokens))
		}

		// 使用GenerateFromSinglePrompt支持选项
		completion, err := llms.GenerateFromSinglePrompt(ctx, s.llm, formattedPrompt, options...)
		if err != nil {
			logger.GetLogger().WithError(err).Error("AI query failed")
			return nil, fmt.Errorf("AI service error: %w", err)
		}
		response = completion
	} else {
		// 使用默认选项
		completion, err := llms.GenerateFromSinglePrompt(ctx, s.llm, formattedPrompt)
		if err != nil {
			logger.GetLogger().WithError(err).Error("AI query failed")
			return nil, fmt.Errorf("AI service error: %w", err)
		}
		response = completion
	}

	// 计算执行时间
	duration := time.Since(startTime)

	// 构建响应
	model := req.Model
	if model == "" {
		model = s.config.OpenAI.Model
	}
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	result := &QueryResponse{
		Response:     response,
		Model:        model,
		Tokens:       s.estimateTokens(response), // 简单的token估算
		Duration:     duration,
		KnowledgeIDs: knowledgeIDs,
		RelevantDocs: relevantDocs,
	}

	// 保存查询历史
	go s.saveQueryHistory(req, result)

	return result, nil
}

// searchRelevantKnowledge 搜索相关知识
func (s *OpenAIService) searchRelevantKnowledge(ctx context.Context, query string) ([]string, []uint, error) {
	// 检查向量服务是否可用
	if s.vectorService == nil {
		logger.GetLogger().Warn("Vector service is not available, skipping knowledge search")
		return []string{}, []uint{}, nil
	}

	db := database.GetDatabase()
	if db == nil {
		logger.GetLogger().Warn("Database is not available, skipping knowledge search")
		return []string{}, []uint{}, nil
	}

	// 1. 生成查询的向量
	queryEmbedding, err := s.vectorService.GenerateEmbedding(ctx, query)
	if err != nil {
		logger.GetLogger().WithError(err).Warn("Failed to generate query embedding, continuing without knowledge search")
		return []string{}, []uint{}, nil
	}

	// 2. 在数据库中进行向量相似度搜索
	var knowledges []models.Knowledge
	err = db.Model(&models.Knowledge{}).
		Select("*, (content_vector <-> ?) as distance", pgvector.NewVector(queryEmbedding.Slice())).
		Where("is_published = ? AND (deleted_at IS NULL)", true).
		Order("distance").
		Limit(5).
		Find(&knowledges).Error

	if err != nil {
		logger.GetLogger().WithError(err).Warn("Failed to search knowledge base, continuing without relevant documents")
		return []string{}, []uint{}, nil
	}

	// 提取文档内容和相关知识ID
	var docs []string
	var knowledgeIDs []uint

	for _, k := range knowledges {
		doc := fmt.Sprintf("标题: %s\n内容: %s", k.Title, k.Content)
		if k.Summary != "" {
			doc += fmt.Sprintf("\n摘要: %s", k.Summary)
		}
		docs = append(docs, doc)
		knowledgeIDs = append(knowledgeIDs, k.ID)
	}

	return docs, knowledgeIDs, nil
}

// buildSystemPrompt 构建系统提示
func (s *OpenAIService) buildSystemPrompt(relevantDocs []string) string {
	basePrompt := `你是一个专业的知识库助手，专注于根据提供的知识库内容回答用户的问题。

回答要求：
1. 基于提供的知识库内容进行回答
2. 如果知识库中没有相关信息，诚实地说明而不是编造
3. 回答要准确、简洁、有条理
4. 使用中文回答，语气友好专业
5. 如果信息不完整，可以建议用户查看相关知识条目`

	if len(relevantDocs) > 0 {
		contextSection := "\n\n相关知识库内容：\n"
		for i, doc := range relevantDocs {
			contextSection += fmt.Sprintf("\n--- 知识 %d ---\n%s\n", i+1, doc)
		}
		basePrompt += contextSection
	}

	return basePrompt
}

// estimateTokens 估算token数量（简单实现）
func (s *OpenAIService) estimateTokens(text string) int {
	// 简单的token估算：中文字符按1个token计算，英文单词按0.75个token计算
	chineseCount := 0
	englishWords := strings.Fields(text)

	// 计算中文字符
	for _, char := range text {
		if char >= 0x4e00 && char <= 0x9fff {
			chineseCount++
		}
	}

	// 估算token数
	return chineseCount + int(float64(len(englishWords))*0.75)
}

// saveQueryHistory 保存查询历史
func (s *OpenAIService) saveQueryHistory(req QueryRequest, resp *QueryResponse) {
	db := database.GetDatabase()

	// 提取相关的知识ID
	var knowledgeID *uint
	if len(resp.KnowledgeIDs) > 0 {
		knowledgeID = &resp.KnowledgeIDs[0]
	}

	// 创建查询历史记录
	history := models.QueryHistory{
		Query:       req.Query,
		Response:    resp.Response,
		KnowledgeID: knowledgeID,
		Model:       resp.Model,
		Tokens:      resp.Tokens,
		Duration:    int(resp.Duration.Milliseconds()),
		IsSuccess:   true,
	}

	if err := db.Create(&history).Error; err != nil {
		logger.WithError(err).Error("Failed to save query history")
	}

	// 更新相关知识的使用计数
	if len(resp.KnowledgeIDs) > 0 {
		for _, kid := range resp.KnowledgeIDs {
			db.Model(&models.Knowledge{}).Where("id = ?", kid).
				Update("view_count", gorm.Expr("view_count + ?", 1))
		}
	}
}

func (s *OpenAIService) GetModels() []string {
	// 构建API URL
	url := s.config.OpenAI.BaseURL
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += "v1/models"

	// 创建HTTP请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.GetLogger().WithError(err).Error("Failed to create request for models")
		return s.getDefaultModels()
	}

	// 添加认证头
	req.Header.Add("Authorization", "Bearer "+s.config.OpenAI.APIKey)
	req.Header.Add("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		logger.GetLogger().WithError(err).Error("Failed to fetch models")
		return s.getDefaultModels()
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		logger.GetLogger().WithField("status_code", resp.StatusCode).Error("Failed to fetch models, non-200 status code")
		return s.getDefaultModels()
	}

	// 解析响应
	var modelsResponse struct {
		Object string `json:"object"`
		Data   []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResponse); err != nil {
		logger.GetLogger().WithError(err).Error("Failed to decode models response")
		return s.getDefaultModels()
	}

	// 提取模型ID
	var modelIds []string
	for _, model := range modelsResponse.Data {
		modelIds = append(modelIds, model.ID)
	}

	// 如果没有获取到模型，返回默认模型
	if len(modelIds) == 0 {
		return s.getDefaultModels()
	}

	return modelIds
}

// getDefaultModels 返回默认模型列表
func (s *OpenAIService) getDefaultModels() []string {
	// 根据配置的base_url返回不同的默认模型
	if strings.Contains(s.config.OpenAI.BaseURL, "api.chatanywhere.tech") {
		return []string{
			"gpt-3.5-turbo",
			"gpt-3.5-turbo-16k",
			"gpt-4",
			"gpt-4-32k",
			"gpt-4-turbo",
			"deepseek-r1",
			"deepseek-coder",
		}
	}

	// OpenAI官方默认模型
	return []string{
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
		"gpt-4",
		"gpt-4-32k",
		"gpt-4-turbo-preview",
		"gpt-4-vision-preview",
	}
}
