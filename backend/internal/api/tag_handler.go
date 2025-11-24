package api

import (
	"net/http"
	"strconv"

	"ai-knowledge-app/internal/models"
	"ai-knowledge-app/pkg/database"
	"ai-knowledge-app/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ========== 标签处理器 ==========

// TagHandler 标签处理器
type TagHandler struct{}

// NewTagHandler 创建标签处理器
func NewTagHandler() *TagHandler {
	return &TagHandler{}
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=50"`
	Color string `json:"color" binding:"omitempty,len=7"`
}

// GetTags 获取标签列表
func (h *TagHandler) GetTags(c *gin.Context) {
	db := database.GetDatabase()

	var tags []models.Tag
	query := db.Model(&models.Tag{})

	// 过滤条件
	if isActive := c.Query("is_active"); isActive != "" {
		// 由于Tag模型没有IsActive字段，我们可以基于DeletedAt判断
		if isActive == "true" {
			query = query.Where("deleted_at IS NULL")
		}
	}

	// 搜索条件
	if search := c.Query("search"); search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("name LIKE ?", searchTerm)
	}

	// 排序
	query = query.Order("usage_count DESC, name ASC")

	if err := query.Find(&tags).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tags")
		return
	}

	utils.SuccessResponse(c, tags)
}

// GetTag 获取单个标签
func (h *TagHandler) GetTag(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var tag models.Tag
	if err := db.Preload("Knowledges", "is_published = ?", true).
		First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Tag not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tag")
		return
	}

	utils.SuccessResponse(c, tag)
}

// CreateTag 创建标签
func (h *TagHandler) CreateTag(c *gin.Context) {
	db := database.GetDatabase()

	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 检查标签名称是否已存在
	var existingTag models.Tag
	if err := db.Where("name = ?", req.Name).First(&existingTag).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Tag name already exists")
		return
	}

	// 创建标签
	tag := models.Tag{
		Name: utils.CleanText(req.Name),
		Color: req.Color,
	}

	if tag.Color == "" {
		tag.Color = generateRandomColor()
	}

	if err := db.Create(&tag).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create tag")
		return
	}

	utils.SuccessResponse(c, tag)
}

// UpdateTag 更新标签
func (h *TagHandler) UpdateTag(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var tag models.Tag
	if err := db.First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Tag not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tag")
		return
	}

	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 检查名称是否与其他标签冲突
	if req.Name != tag.Name {
		var existingTag models.Tag
		if err := db.Where("name = ? AND id != ?", req.Name, tag.ID).First(&existingTag).Error; err == nil {
			utils.ErrorResponse(c, http.StatusConflict, "Tag name already exists")
			return
		}
	}

	// 更新字段
	tag.Name = utils.CleanText(req.Name)
	if req.Color != "" {
		tag.Color = req.Color
	}

	if err := db.Save(&tag).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update tag")
		return
	}

	utils.SuccessResponse(c, tag)
}

// DeleteTag 删除标签
func (h *TagHandler) DeleteTag(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var tag models.Tag
	if err := db.First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Tag not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tag")
		return
	}

	// 检查是否有关联的知识
	var knowledgeCount int64
	db.Table("knowledge_tags").Where("tag_id = ?", tag.ID).Count(&knowledgeCount)
	if knowledgeCount > 0 {
		utils.ErrorResponse(c, http.StatusConflict, "Cannot delete tag with associated knowledges")
		return
	}

	// 软删除
	if err := db.Delete(&tag).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete tag")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Tag deleted successfully"})
}

// GetTagKnowledges 获取标签下的知识
func (h *TagHandler) GetTagKnowledges(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	// 验证标签存在
	var tag models.Tag
	if err := db.First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Tag not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tag")
		return
	}

	// 解析分页参数
	var pagination utils.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 构建查询
	query := db.Table("knowledges").
		Select("knowledges.*").
		Joins("INNER JOIN knowledge_tags ON knowledges.id = knowledge_tags.knowledge_id").
		Joins("INNER JOIN categories ON knowledges.category_id = categories.id").
		Where("knowledge_tags.tag_id = ? AND knowledges.is_published = ?", tag.ID, true).
		Preload("Category").
		Preload("Tags")

	// 搜索条件
	if pagination.Search != "" {
		searchTerm := "%" + pagination.Search + "%"
		query = query.Where("knowledges.title LIKE ? OR knowledges.content LIKE ? OR knowledges.summary LIKE ?",
			searchTerm, searchTerm, searchTerm)
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

	if err := query.Order("knowledges.created_at DESC").
		Offset(offset).Limit(pagination.PageSize).Find(&knowledges).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch knowledges")
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

	// 添加标签信息到响应数据中
	responseData := map[string]interface{}{
		"pagination": response,
		"tag":        tag,
	}

	utils.SuccessResponse(c, responseData)
}

// GetPopularTags 获取热门标签
func (h *TagHandler) GetPopularTags(c *gin.Context) {
	db := database.GetDatabase()

	// 获取limit参数
	limit := 20 // 默认20个
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var tags []models.Tag

	// 基于使用次数获取热门标签
	query := db.Model(&models.Tag{}).
		Where("usage_count > 0").
		Order("usage_count DESC, name ASC").
		Limit(limit)

	if err := query.Find(&tags).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch popular tags")
		return
	}

	// 如果基于使用次数的标签不够，补充一些常用标签
	if len(tags) < limit {
		remaining := limit - len(tags)
		var additionalTags []models.Tag

		// 获取未被选中的标签，按名称排序
		var tagIds []uint
		for _, t := range tags {
			tagIds = append(tagIds, t.ID)
		}

		additionalQuery := db.Model(&models.Tag{})
		if len(tagIds) > 0 {
			additionalQuery = additionalQuery.Where("id NOT IN ?", tagIds)
		}
		additionalQuery.Order("name ASC").Limit(remaining).Find(&additionalTags)

		tags = append(tags, additionalTags...)
	}

	utils.SuccessResponse(c, tags)
}