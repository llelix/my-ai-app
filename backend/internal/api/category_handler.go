package api

import (
	"net/http"

	"ai-knowledge-app/internal/models"
	"ai-knowledge-app/pkg/database"
	"ai-knowledge-app/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ========== 分类处理器 ==========

// CategoryHandler 分类处理器
type CategoryHandler struct{}

// NewCategoryHandler 创建分类处理器
func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
	Color       string `json:"color" binding:"omitempty,len=7"`
	Icon        string `json:"icon" binding:"omitempty,max=50"`
	ParentID    *uint  `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}

// GetCategories 获取分类列表
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	db := database.GetDatabase()

	var categories []models.Category
	query := db.Preload("Parent").Preload("Children")

	// 过滤条件
	if isActive := c.Query("is_active"); isActive != "" {
		query = query.Where("is_active = ?", isActive == "true")
	}

	// 排序
	query = query.Order("sort_order ASC, created_at ASC")

	if err := query.Find(&categories).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	utils.SuccessResponse(c, categories)
}

// GetCategory 获取单个分类
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var category models.Category
	if err := db.Preload("Parent").Preload("Children").
		Preload("Knowledges", "is_published = ?", true).
		First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Category not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch category")
		return
	}

	utils.SuccessResponse(c, category)
}

// CreateCategory 创建分类
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	db := database.GetDatabase()

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 检查父分类是否存在
	if req.ParentID != nil {
		var parent models.Category
		if err := db.First(&parent, *req.ParentID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid parent category")
			return
		}
	}

	// 检查分类名称是否已存在
	var existingCategory models.Category
	if err := db.Where("name = ?", req.Name).First(&existingCategory).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Category name already exists")
		return
	}

	// 创建分类
	category := models.Category{
		Name:        utils.CleanText(req.Name),
		Description: utils.CleanText(req.Description),
		Color:       req.Color,
		Icon:        req.Icon,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}

	if err := db.Create(&category).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create category")
		return
	}

	// 重新加载完整的分类对象
	db.Preload("Parent").First(&category, category.ID)

	utils.SuccessResponse(c, category)
}

// UpdateCategory 更新分类
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var category models.Category
	if err := db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Category not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch category")
		return
	}

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 检查父分类是否存在
	if req.ParentID != nil {
		var parent models.Category
		if err := db.First(&parent, *req.ParentID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid parent category")
			return
		}
		// 不能设置自己为父分类
		if *req.ParentID == category.ID {
			utils.ErrorResponse(c, http.StatusBadRequest, "Cannot set self as parent")
			return
		}
	}

	// 检查名称是否与其他分类冲突
	if req.Name != category.Name {
		var existingCategory models.Category
		if err := db.Where("name = ? AND id != ?", req.Name, category.ID).First(&existingCategory).Error; err == nil {
			utils.ErrorResponse(c, http.StatusConflict, "Category name already exists")
			return
		}
	}

	// 更新字段
	category.Name = utils.CleanText(req.Name)
	category.Description = utils.CleanText(req.Description)
	category.Color = req.Color
	category.Icon = req.Icon
	category.ParentID = req.ParentID
	category.SortOrder = req.SortOrder
	category.IsActive = req.IsActive

	if err := db.Save(&category).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update category")
		return
	}

	// 重新加载完整的分类对象
	db.Preload("Parent").Preload("Children").First(&category, category.ID)

	utils.SuccessResponse(c, category)
}

// DeleteCategory 删除分类
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	var category models.Category
	if err := db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Category not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch category")
		return
	}

	// 检查是否有子分类
	var childCount int64
	db.Model(&models.Category{}).Where("parent_id = ?", category.ID).Count(&childCount)
	if childCount > 0 {
		utils.ErrorResponse(c, http.StatusConflict, "Cannot delete category with subcategories")
		return
	}

	// 检查是否有关联的知识
	var knowledgeCount int64
	db.Model(&models.Knowledge{}).Where("category_id = ?", category.ID).Count(&knowledgeCount)
	if knowledgeCount > 0 {
		utils.ErrorResponse(c, http.StatusConflict, "Cannot delete category with associated knowledges")
		return
	}

	// 软删除
	if err := db.Delete(&category).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Category deleted successfully"})
}

// GetCategoryKnowledges 获取分类下的知识
func (h *CategoryHandler) GetCategoryKnowledges(c *gin.Context) {
	db := database.GetDatabase()
	id := c.Param("id")

	// 验证分类存在
	var category models.Category
	if err := db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Category not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch category")
		return
	}

	// 解析分页参数
	var pagination utils.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// 构建查询
	query := db.Model(&models.Knowledge{}).
		Preload("Category").
		Preload("Tags").
		Where("category_id = ? AND is_published = ?", category.ID, true)

	// 搜索条件
	if pagination.Search != "" {
		searchTerm := "%" + pagination.Search + "%"
		query = query.Where("title LIKE ? OR content LIKE ? OR summary LIKE ?",
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

	if err := query.Order("created_at DESC").
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

	// 添加分类信息到响应数据中
	responseData := map[string]interface{}{
		"pagination": response,
		"category":   category,
	}

	utils.SuccessResponse(c, responseData)
}