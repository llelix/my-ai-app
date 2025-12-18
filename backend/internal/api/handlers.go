package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ai-knowledge-app/internal/models"
	"ai-knowledge-app/internal/service"
	"ai-knowledge-app/pkg/database"
	"ai-knowledge-app/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// Validate 验证器实例
var Validate = validator.New()

// ========== 知识库处理器 ==========

// KnowledgeHandler 知识库处理器
type KnowledgeHandler struct {
	vectorService service.VectorService
}

// NewKnowledgeHandler 创建知识库处理器
func NewKnowledgeHandler(vectorService service.VectorService) *KnowledgeHandler {
	return &KnowledgeHandler{
		vectorService: vectorService,
	}
}

// CreateKnowledgeRequest 创建知识请求
type CreateKnowledgeRequest struct {
	Title       string          `json:"title" binding:"required,min=1,max=255"`
	Content     string          `json:"content" binding:"required"`
	Summary     string          `json:"summary"`
	CategoryID  uint            `json:"category_id"`
	Tags        []string        `json:"tags"`
	Metadata    models.Metadata `json:"metadata"`
	IsPublished bool            `json:"is_published"`
}

// UpdateKnowledgeRequest 更新知识请求
type UpdateKnowledgeRequest struct {
	Title       string          `json:"title" binding:"omitempty,min=1,max=255"`
	Content     string          `json:"content"`
	Summary     string          `json:"summary"`
	CategoryID  uint            `json:"category_id"`
	Tags        []string        `json:"tags"`
	Metadata    models.Metadata `json:"metadata"`
	IsPublished *bool           `json:"is_published"`
}

// GetKnowledges 获取知识列表
// @Summary Get knowledge list
// @Description Get paginated list of knowledge entries
// @Tags knowledge
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} utils.PaginationResponse
// @Router /knowledge [get]
func (h *KnowledgeHandler) GetKnowledges(c *gin.Context) {
	db := database.GetDatabase()

	// 解析分页参数
	var pagination utils.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 构建查询
	query := db.Model(&models.Knowledge{}).Preload("Category").Preload("Tags")

	// 搜索条件
	if pagination.Search != "" {
		searchTerm := "%" + pagination.Search + "%"
		query = query.Where("title LIKE ? OR content LIKE ? OR summary LIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	// 只显示已发布的（前端可以指定是否包含未发布的）
	if !utils.ContainsString([]string{"true", "1"}, c.Query("include_unpublished")) {
		query = query.Where("is_published = ?", true)
	}

	// 分类过滤
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32); err == nil {
			query = query.Where("category_id = ?", categoryID)
		}
	}

	// 标签过滤
	if tagIDStr := c.Query("tag_id"); tagIDStr != "" {
		if tagID, err := strconv.ParseUint(tagIDStr, 10, 32); err == nil {
			query = query.Joins("INNER JOIN knowledge_tags ON knowledges.id = knowledge_tags.knowledge_id").
				Where("knowledge_tags.tag_id = ?", tagID)
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count knowledges")
		return
	}

	// 分页查询
	offset := utils.GetOffset(pagination.Page, pagination.PageSize)
	var knowledges []models.Knowledge

	// 排序
	orderClause := "created_at DESC"
	if pagination.Sort != "" {
		orderClause = fmt.Sprintf("%s %s", pagination.Sort, strings.ToUpper(pagination.Order))
	}
	query = query.Order(orderClause)

	if err := query.Offset(offset).Limit(pagination.PageSize).Find(&knowledges).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch knowledges")
		return
	}

	// 构建分页响应
	response := utils.PaginationResponse{
		Items:      knowledges,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: utils.CalculateTotalPages(total, pagination.PageSize),
	}

	utils.SuccessResponse(c, response)
}

// GetKnowledge 获取单个知识
// @Summary 获取单个知识条目
// @Description 根据ID获取知识条目详情
// @Tags knowledge
// @Accept json
// @Produce json
// @Param id path int true "知识ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /knowledge/{id} [get]
func (h *KnowledgeHandler) GetKnowledge(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var knowledge models.Knowledge
	if err := db.Preload("Category").Preload("Tags").First(&knowledge, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Knowledge not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch knowledge")
		return
	}

	utils.SuccessResponse(c, knowledge)
}

// CreateKnowledge 创建知识
// @Summary 创建新的知识条目
// @Description 创建新的知识条目，支持分类和标签
// @Tags knowledge
// @Accept json
// @Produce json
// @Param request body CreateKnowledgeRequest true "创建知识请求"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /knowledge [post]
func (h *KnowledgeHandler) CreateKnowledge(c *gin.Context) {
	db := database.GetDatabase()

	var req CreateKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 验证分类是否存在
	if req.CategoryID > 0 {
		var category models.Category
		if err := db.First(&category, req.CategoryID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category")
			return
		}
	}

	// 创建知识
	knowledge := models.Knowledge{
		Title:         utils.CleanText(req.Title),
		Content:       utils.CleanText(req.Content),
		ContentVector: nil, // 初始为空，后续异步生成
		Summary:       utils.CleanText(req.Summary),
		CategoryID:    req.CategoryID,
		Metadata:      req.Metadata,
		IsPublished:   req.IsPublished,
	}

	// 如果没有提供摘要，自动生成
	if knowledge.Summary == "" {
		knowledge.Summary = utils.TruncateText(knowledge.Content, 200)
	}

	// 保存知识
	if err := db.Create(&knowledge).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create knowledge: %v", err))
		return
	}

	// 异步生成和保存向量（不阻塞主流程）
	go func(knowledgeID uint) {
		embedding, err := h.vectorService.GenerateEmbedding(context.Background(), knowledge.Content)
		if err != nil {
			// 向量生成失败，不影响知识保存，只记录日志
			// logger.GetLogger().WithError(err).Warn("Failed to generate embedding for knowledge ID: ", knowledgeID)
			return
		}
		if err := db.Model(&models.Knowledge{}).Where("id = ?", knowledgeID).Update("content_vector", &embedding).Error; err != nil {
			// logger.GetLogger().WithError(err).Warn("Failed to save embedding for knowledge ID: ", knowledgeID)
		}
	}(knowledge.ID)

	// 处理标签
	if len(req.Tags) > 0 {
		if err := h.attachTags(&knowledge, req.Tags); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to attach tags: %v", err))
			return
		}
	}

	// 重新加载完整的知识对象
	db.Preload("Category").Preload("Tags").First(&knowledge, knowledge.ID)

	utils.SuccessResponse(c, knowledge)
}

// UpdateKnowledge 更新知识
// @Summary 更新知识条目
// @Description 更新指定ID的知识条目
// @Tags knowledge
// @Accept json
// @Produce json
// @Param id path int true "知识ID"
// @Param request body UpdateKnowledgeRequest true "更新知识请求"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /knowledge/{id} [put]
func (h *KnowledgeHandler) UpdateKnowledge(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var knowledge models.Knowledge
	if err := db.First(&knowledge, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Knowledge not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch knowledge")
		return
	}

	var req UpdateKnowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 验证分类是否存在
	if req.CategoryID > 0 {
		var category models.Category
		if err := db.First(&category, req.CategoryID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category")
			return
		}
		knowledge.CategoryID = req.CategoryID
	}

	// 更新字段
	if req.Title != "" {
		knowledge.Title = utils.CleanText(req.Title)
	}

	contentChanged := false
	if req.Content != "" && req.Content != knowledge.Content {
		knowledge.Content = utils.CleanText(req.Content)
		contentChanged = true
	}

	if req.Summary != "" {
		knowledge.Summary = utils.CleanText(req.Summary)
	} else if contentChanged {
		// 如果更新了内容但没有提供摘要，自动生成
		knowledge.Summary = utils.TruncateText(req.Content, 200)
	}
	if req.IsPublished != nil {
		knowledge.IsPublished = *req.IsPublished
	}

	// 更新元数据
	if req.Metadata.Author != "" {
		knowledge.Metadata.Author = req.Metadata.Author
	}
	if req.Metadata.Source != "" {
		knowledge.Metadata.Source = req.Metadata.Source
	}
	if req.Metadata.Language != "" {
		knowledge.Metadata.Language = req.Metadata.Language
	}
	if req.Metadata.Difficulty != "" {
		knowledge.Metadata.Difficulty = req.Metadata.Difficulty
	}
	if req.Metadata.Keywords != "" {
		knowledge.Metadata.Keywords = req.Metadata.Keywords
	}

	// 保存更新
	if err := db.Save(&knowledge).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update knowledge")
		return
	}

	// 如果内容有变化且不为空，更新向量
	if contentChanged && knowledge.Content != "" {
		embedding, err := h.vectorService.GenerateEmbedding(context.Background(), knowledge.Content)
		if err != nil {
			// 即使生成向量失败，也应保存知识的其他更新
			// 但记录一个错误日志
			// logger.GetLogger().WithError(err).Warn("Failed to update embedding for knowledge ID: ", knowledge.ID)
		} else {
			if err := db.Model(&knowledge).Update("content_vector", embedding).Error; err != nil {
				// logger.GetLogger().WithError(err).Warn("Failed to save embedding for knowledge ID: ", knowledge.ID)
			}
		}
	}

	// 处理标签
	if len(req.Tags) > 0 {
		// 清除现有标签关联
		db.Model(&knowledge).Association("Tags").Clear()
		// 添加新标签
		if err := h.attachTags(&knowledge, req.Tags); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to attach tags")
			return
		}
	}

	// 重新加载完整的知识对象
	db.Preload("Category").Preload("Tags").First(&knowledge, knowledge.ID)

	utils.SuccessResponse(c, knowledge)
}

// DeleteKnowledge 删除知识
// @Summary 删除知识条目
// @Description 软删除指定ID的知识条目
// @Tags knowledge
// @Accept json
// @Produce json
// @Param id path int true "知识ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /knowledge/{id} [delete]
func (h *KnowledgeHandler) DeleteKnowledge(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var knowledge models.Knowledge
	if err := db.First(&knowledge, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Knowledge not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch knowledge")
		return
	}

	// 软删除
	if err := db.Delete(&knowledge).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete knowledge")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Knowledge deleted successfully"})
}

// SearchKnowledges 搜索知识
func (h *KnowledgeHandler) SearchKnowledges(c *gin.Context) {
	db := database.GetDatabase()

	query := c.Query("q")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Search query is required")
		return
	}

	// 解析分页参数
	var pagination utils.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 构建搜索查询
	searchTerm := "%" + strings.ToLower(query) + "%"
	dbQuery := db.Model(&models.Knowledge{}).
		Preload("Category").
		Preload("Tags").
		Where("(LOWER(title) LIKE ? OR LOWER(content) LIKE ? OR LOWER(summary) LIKE ? OR LOWER(metadata.keywords) LIKE ?) AND is_published = ?",
			searchTerm, searchTerm, searchTerm, searchTerm, true)

	// 获取总数
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count search results")
		return
	}

	// 分页查询
	offset := utils.GetOffset(pagination.Page, pagination.PageSize)
	var knowledges []models.Knowledge

	if err := dbQuery.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&knowledges).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to search knowledges")
		return
	}

	// 构建响应
	response := utils.PaginationResponse{
		Items:      knowledges,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: utils.CalculateTotalPages(total, pagination.PageSize),
	}

	utils.SuccessResponse(c, response)
}

// GetRelatedKnowledges 获取相关知识
func (h *KnowledgeHandler) GetRelatedKnowledges(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var knowledge models.Knowledge
	if err := db.First(&knowledge, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Knowledge not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch knowledge")
		return
	}

	// 获取limit参数
	limitStr := c.DefaultQuery("limit", "5")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	// 基于分类和标签查找相关知识
	var relatedKnowledges []models.Knowledge

	// 同分类的知识
	db.Preload("Category").Preload("Tags").
		Where("category_id = ? AND id != ? AND is_published = ?",
			knowledge.CategoryID, knowledge.ID, true).
		Order("created_at DESC").
		Limit(limit).
		Find(&relatedKnowledges)

	// 如果同分类的知识不够，添加同标签的知识
	if len(relatedKnowledges) < limit {
		var tagIDs []uint
		for _, tag := range knowledge.Tags {
			tagIDs = append(tagIDs, tag.ID)
		}

		if len(tagIDs) > 0 {
			var tagKnowledges []models.Knowledge
			db.Table("knowledges").
				Select("knowledges.*").
				Joins("INNER JOIN knowledge_tags ON knowledges.id = knowledge_tags.knowledge_id").
				Where("knowledge_tags.tag_id IN ? AND knowledges.id != ? AND knowledges.id NOT IN (?) AND knowledges.is_published = ?",
					tagIDs, knowledge.ID,
					func() []uint {
						existingIDs := []uint{knowledge.ID}
						for _, k := range relatedKnowledges {
							existingIDs = append(existingIDs, k.ID)
						}
						return existingIDs
					}(), true).
				Order("created_at DESC").
				Limit(limit - len(relatedKnowledges)).
				Scan(&tagKnowledges)

			relatedKnowledges = append(relatedKnowledges, tagKnowledges...)
		}
	}

	utils.SuccessResponse(c, relatedKnowledges)
}

// IncrementViewCount 增加查看次数
func (h *KnowledgeHandler) IncrementViewCount(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var knowledge models.Knowledge
	if err := db.First(&knowledge, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Knowledge not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch knowledge")
		return
	}

	// 增加查看次数
	if err := db.Model(&knowledge).Update("view_count", knowledge.ViewCount+1).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update view count")
		return
	}

	utils.SuccessResponse(c, gin.H{"view_count": knowledge.ViewCount + 1})
}

// attachTags 为知识附加标签
func (h *KnowledgeHandler) attachTags(knowledge *models.Knowledge, tagNames []string) error {
	db := database.GetDatabase()
	var tags []models.Tag

	for _, tagName := range tagNames {
		tagName = utils.CleanText(tagName)
		if tagName == "" {
			continue
		}

		var tag models.Tag
		// 查找或创建标签
		if err := db.Where("name = ?", tagName).First(&tag).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 创建新标签
				tag = models.Tag{
					Name:  tagName,
					Color: generateRandomColor(),
				}
				if err := db.Create(&tag).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		tags = append(tags, tag)
	}

	// 关联标签
	return db.Model(knowledge).Association("Tags").Append(&tags)
}

// generateRandomColor 生成随机颜色
func generateRandomColor() string {
	colors := []string{
		"#ff6b6b", "#4ecdc4", "#45b7d1", "#f9ca24", "#6c5ce7",
		"#a29bfe", "#fd79a8", "#fdcb6e", "#e17055", "#00b894",
		"#00cec9", "#0984e3", "#74b9ff", "#a29bfe", "#dfe6e9",
	}
	return colors[len(colors)%len(colors)]
}
