package api

import (
	"context"
	"net/http"
	"time"

	"ai-knowledge-app/internal/ai"
	"ai-knowledge-app/internal/models"
	"ai-knowledge-app/pkg/database"
	"ai-knowledge-app/pkg/logger"
	"ai-knowledge-app/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ========== AI查询处理器 ==========

// AIHandler AI处理器
type AIHandler struct {
	aiService ai.AIService
}

// NewAIHandler 创建AI处理器
func NewAIHandler() *AIHandler {
	// 这里应该从配置中创建AI服务实例
	// 暂时返回空处理器，实际使用时需要注入配置
	return &AIHandler{
		aiService: nil, // 将在实际初始化时注入
	}
}

// SetAIService 设置AI服务
func (h *AIHandler) SetAIService(service ai.AIService) {
	h.aiService = service
}

// QueryRequest AI查询请求
type QueryRequest struct {
	Query       string   `json:"query" binding:"required,min=1,max=1000"`
	Model       string   `json:"model,omitempty"`
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Context     []string `json:"context,omitempty"`
}

// QueryResponse AI查询响应
type QueryResponse struct {
	Response          string             `json:"response"`
	Model             string             `json:"model"`
	Tokens            int                `json:"tokens"`
	Duration          int                `json:"duration"` // 毫秒
	KnowledgeIDs      []uint             `json:"knowledge_ids,omitempty"`
	RelevantDocs      []string           `json:"relevant_docs,omitempty"`
	RelatedKnowledges []models.Knowledge `json:"related_knowledges,omitempty"`
}

// Query AI查询接口
// @Summary AI智能查询
// @Description 基于存储的知识库进行AI智能查询
// @Tags ai
// @Accept json
// @Produce json
// @Param request body QueryRequest true "查询请求"
// @Success 200 {object} QueryResponse
// @Failure 400 {object} utils.Response
// @Failure 503 {object} utils.Response
// @Router /ai/query [post]
func (h *AIHandler) Query(c *gin.Context) {
	if h.aiService == nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "AI service is not configured")
		return
	}

	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 设置默认参数
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 2000
	}

	// 记录查询日志
	logger.GetLogger().WithFields(map[string]interface{}{
		"query":       req.Query,
		"model":       req.Model,
		"temperature": req.Temperature,
	}).Info("AI query request")

	// 调用AI服务
	ctx := context.Background()
	aiResp, err := h.aiService.Query(ctx, ai.QueryRequest{
		Query:       req.Query,
		Model:       req.Model,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Context:     req.Context,
	})

	if err != nil {
		logger.GetLogger().WithError(err).Error("AI query failed")

		// 保存失败的查询记录
		go h.saveFailedQuery(req, err)

		utils.ErrorResponse(c, http.StatusInternalServerError, "AI query failed: "+err.Error())
		return
	}

	// 获取相关知识详情
	var relatedKnowledges []models.Knowledge
	if len(aiResp.KnowledgeIDs) > 0 {
		db := database.GetDatabase()
		db.Where("id IN ? AND is_published = ?", aiResp.KnowledgeIDs, true).
			Find(&relatedKnowledges)
	}

	// 构建响应
	response := QueryResponse{
		Response:          aiResp.Response,
		Model:             aiResp.Model,
		Tokens:            aiResp.Tokens,
		Duration:          int(aiResp.Duration.Milliseconds()),
		KnowledgeIDs:      aiResp.KnowledgeIDs,
		RelevantDocs:      aiResp.RelevantDocs,
		RelatedKnowledges: relatedKnowledges,
	}

	utils.SuccessResponse(c, response)
}

// GetQueryHistory 获取查询历史
func (h *AIHandler) GetQueryHistory(c *gin.Context) {
	db := database.GetDatabase()

	// 解析分页参数
	var pagination utils.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 构建查询
	query := db.Model(&models.QueryHistory{}).
		Preload("Knowledge").
		Where("is_success = ?", true)

	// 搜索条件
	if pagination.Search != "" {
		searchTerm := "%" + pagination.Search + "%"
		query = query.Where("query LIKE ? OR response LIKE ?", searchTerm, searchTerm)
	}

	// 模型筛选
	if model := c.Query("model"); model != "" {
		query = query.Where("model = ?", model)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count query history")
		return
	}

	// 分页查询
	offset := utils.GetOffset(pagination.Page, pagination.PageSize)
	var histories []models.QueryHistory

	if err := query.Order("created_at DESC").
		Offset(offset).Limit(pagination.PageSize).Find(&histories).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch query history")
		return
	}

	// 构建响应
	response := utils.PaginationResponse{
		Items:      histories,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: utils.CalculateTotalPages(total, pagination.PageSize),
	}

	utils.SuccessResponse(c, response)
}

// DeleteQueryHistory 删除查询历史
func (h *AIHandler) DeleteQueryHistory(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var history models.QueryHistory
	if err := db.First(&history, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Query history not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch query history")
		return
	}

	if err := db.Delete(&history).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete query history")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Query history deleted successfully"})
}

// GetQueryStats 获取查询统计
func (h *AIHandler) GetQueryStats(c *gin.Context) {
	db := database.GetDatabase()

	// 今日查询数量
	var todayCount int64
	today := time.Now().Truncate(24 * time.Hour)
	db.Model(&models.QueryHistory{}).
		Where("created_at >= ? AND is_success = ?", today, true).
		Count(&todayCount)

	// 本周查询数量
	var weekCount int64
	weekStart := time.Now().AddDate(0, 0, -7)
	db.Model(&models.QueryHistory{}).
		Where("created_at >= ? AND is_success = ?", weekStart, true).
		Count(&weekCount)

	// 总查询数量和成功率
	var totalCount, successCount int64
	db.Model(&models.QueryHistory{}).Count(&totalCount)
	db.Model(&models.QueryHistory{}).Where("is_success = ?", true).Count(&successCount)

	successRate := float64(0)
	if totalCount > 0 {
		successRate = float64(successCount) / float64(totalCount) * 100
	}

	// 按模型统计
	var modelStats []struct {
		Model string `json:"model"`
		Count int64  `json:"count"`
	}

	db.Model(&models.QueryHistory{}).
		Select("model, count(*) as count").
		Where("is_success = ?", true).
		Group("model").
		Order("count desc").
		Scan(&modelStats)

	// 最常用的查询词
	var popularQueries []struct {
		Query string `json:"query"`
		Count int64  `json:"count"`
	}

	db.Model(&models.QueryHistory{}).
		Select("query, count(*) as count").
		Where("is_success = ? AND length(query) <= 100", true).
		Group("query").
		Order("count desc").
		Limit(10).
		Scan(&popularQueries)

	// 平均响应时间
	var avgDuration float64
	db.Model(&models.QueryHistory{}).
		Where("is_success = ?", true).
		Select("AVG(duration)").
		Scan(&avgDuration)

	stats := gin.H{
		"today_count":     todayCount,
		"week_count":      weekCount,
		"total_count":     totalCount,
		"success_count":   successCount,
		"success_rate":    successRate,
		"avg_duration":    avgDuration,
		"by_models":       modelStats,
		"popular_queries": popularQueries,
	}

	utils.SuccessResponse(c, stats)
}

// SubmitFeedback 提交反馈
type FeedbackRequest struct {
	QueryID   uint   `json:"query_id" binding:"required"`
	Rating    int    `json:"rating" binding:"required,min=1,max=5"`
	Comment   string `json:"comment"`
	IsHelpful bool   `json:"is_helpful"`
}

// SubmitFeedback 提交AI查询反馈
func (h *AIHandler) SubmitFeedback(c *gin.Context) {
	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 这里可以保存反馈信息到数据库
	// 暂时只记录日志
	logger.GetLogger().WithFields(map[string]interface{}{
		"query_id":   req.QueryID,
		"rating":     req.Rating,
		"comment":    req.Comment,
		"is_helpful": req.IsHelpful,
	}).Info("AI query feedback submitted")

	utils.SuccessResponse(c, gin.H{"message": "Feedback submitted successfully"})
}

// GetModels 获取支持的AI模型
func (h *AIHandler) GetModels(c *gin.Context) {
	if h.aiService == nil {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "AI service is not configured")
		return
	}

	models := h.aiService.GetModels()
	utils.SuccessResponse(c, gin.H{"models": models})
}

// saveFailedQuery 保存失败的查询
func (h *AIHandler) saveFailedQuery(req QueryRequest, err error) {
	db := database.GetDatabase()

	history := models.QueryHistory{
		Query:        req.Query,
		Response:     "",
		Model:        req.Model,
		Tokens:       0,
		Duration:     0,
		IsSuccess:    false,
		ErrorMessage: err.Error(),
	}

	if err := db.Create(&history).Error; err != nil {
		logger.GetLogger().WithError(err).Error("Failed to save failed query")
	}
}
