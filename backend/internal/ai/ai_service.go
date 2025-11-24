package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-knowledge-app/internal/config"
	"ai-knowledge-app/internal/models"
	"ai-knowledge-app/pkg/database"
	"ai-knowledge-app/pkg/logger"
	"ai-knowledge-app/pkg/utils"

	"gorm.io/gorm"
)

// AIService AI服务接口
type AIService interface {
	Query(ctx context.Context, req QueryRequest) (*QueryResponse, error)
	GetModels() []string
}

// OpenAIService OpenAI兼容的AI服务
type OpenAIService struct {
	config *config.AIConfig
	client *http.Client
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
	Response      string        `json:"response"`
	Model         string        `json:"model"`
	Tokens        int           `json:"tokens"`
	Duration      time.Duration `json:"duration"`
	KnowledgeIDs  []uint        `json:"knowledge_ids,omitempty"`
	RelevantDocs  []string      `json:"relevant_docs,omitempty"`
}

// OpenAIRequest OpenAI API请求结构
type OpenAIRequest struct {
	Model       string                 `json:"model"`
	Messages    []OpenAIMessage         `json:"messages"`
	Temperature float64                `json:"temperature"`
	MaxTokens   int                    `json:"max_tokens"`
	Stream      bool                   `json:"stream"`
}

// OpenAIMessage OpenAI消息结构
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse OpenAI API响应结构
type OpenAIResponse struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage   `json:"usage"`
}

// OpenAIChoice OpenAI选择结构
type OpenAIChoice struct {
	Index   int           `json:"index"`
	Message OpenAIMessage `json:"message"`
	FinishReason string   `json:"finish_reason"`
}

// OpenAIUsage OpenAI使用情况结构
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewAIService 创建AI服务实例
func NewAIService(cfg *config.AIConfig) AIService {
	return &OpenAIService{
		config: cfg,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Query 执行AI查询
func (s *OpenAIService) Query(ctx context.Context, req QueryRequest) (*QueryResponse, error) {
	startTime := time.Now()

	// 获取相关的知识库内容
	relevantDocs, knowledgeIDs, err := s.searchRelevantKnowledge(req.Query)
	if err != nil {
		logger.WithError(err).Error("Failed to search relevant knowledge")
	}

	// 构建系统提示
	systemPrompt := s.buildSystemPrompt(relevantDocs)

	// 构建消息
	messages := []OpenAIMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.Query},
	}

	// 如果有上下文，添加到消息中
	for _, ctx := range req.Context {
		messages = append(messages, OpenAIMessage{
			Role:    "system",
			Content: fmt.Sprintf("参考信息: %s", ctx),
		})
	}

	// 选择模型
	model := req.Model
	if model == "" {
		model = s.config.OpenAI.Model
	}

	// 构建OpenAI请求
	openaiReq := OpenAIRequest{
		Model:       model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      false,
	}

	// 调用API
	response, err := s.callOpenAI(ctx, openaiReq)
	if err != nil {
		return nil, err
	}

	// 计算执行时间
	duration := time.Since(startTime)

	// 构建响应
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

// callOpenAI 调用OpenAI兼容API
func (s *OpenAIService) callOpenAI(ctx context.Context, req OpenAIRequest) (string, error) {
	// 确定使用的API配置
	baseURL := s.config.OpenAI.BaseURL
	apiKey := s.config.OpenAI.APIKey

	// 如果配置了Claude，优先使用Claude
	if s.config.Claude.APIKey != "" {
		baseURL = s.config.Claude.BaseURL
		apiKey = s.config.Claude.APIKey
		if req.Model == "" || req.Model == s.config.OpenAI.Model {
			req.Model = s.config.Claude.Model
		}
	}

	// 构建请求body
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	// 如果是Claude API，可能需要不同的授权头
	if strings.Contains(baseURL, "anthropic.com") {
		httpReq.Header.Set("x-api-key", apiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
	}

	// 发送请求
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// 检查HTTP状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var openaiResp OpenAIResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// 提取回复内容
	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return openaiResp.Choices[0].Message.Content, nil
}

// searchRelevantKnowledge 搜索相关知识
func (s *OpenAIService) searchRelevantKnowledge(query string) ([]string, []uint, error) {
	db := database.GetDatabase()

	// 关键词提取
	keywords := utils.ExtractKeywords(query)
	if len(keywords) == 0 {
		return nil, nil, nil
	}

	// 构建搜索查询
	var knowledges []models.Knowledge
	queryBuilder := db.Model(&models.Knowledge{}).
		Preload("Category").
		Preload("Tags").
		Where("is_published = ? AND (deleted_at IS NULL)", true)

	// 多条件搜索
	searchConditions := []string{}
	searchParams := []interface{}{}

	for _, keyword := range keywords {
		searchPattern := "%" + keyword + "%"
		searchConditions = append(searchConditions, "title LIKE ? OR content LIKE ? OR summary LIKE ? OR metadata.keywords LIKE ?")
		searchParams = append(searchParams, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if len(searchConditions) > 0 {
		queryString := "(" + strings.Join(searchConditions, " OR ") + ")"
		queryBuilder = queryBuilder.Where(queryString, searchParams...)
	}

	// 限制返回数量并按相关性排序（简单实现：按创建时间和查看次数）
	err := queryBuilder.
		Order("view_count DESC, created_at DESC").
		Limit(5).
		Find(&knowledges).Error

	if err != nil {
		return nil, nil, err
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

// GetModels 获取支持的模型列表
func (s *OpenAIService) GetModels() []string {
	models := []string{
		// OpenAI兼容的模型
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
		"gpt-4",
		"gpt-4-32k",
		"gpt-4-turbo-preview",

		// 国内常用模型
		"qwen-turbo",
		"qwen-plus",
		"qwen-max",
		"moonshot-v1-8k",
		"moonshot-v1-32k",
		"moonshot-v1-128k",
		"glm-3-turbo",
		"glm-4",

		// Claude模型
		"claude-3-sonnet-20240229",
		"claude-3-opus-20240229",
		"claude-3-haiku-20240307",
	}

	return models
}