package api

import (
	"net/http"
	"time"

	"ai-knowledge-app/internal/ai"
	"ai-knowledge-app/internal/config"
	"ai-knowledge-app/internal/middleware"
	"ai-knowledge-app/internal/models"
	"ai-knowledge-app/internal/service"
	"ai-knowledge-app/pkg/database"
	"ai-knowledge-app/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Router API路由器
type Router struct {
	config           *config.Config
	knowledgeHandler *KnowledgeHandler
	aiHandler        *AIHandler
	categoryHandler  *CategoryHandler
	tagHandler       *TagHandler
	vectorService    service.VectorService
}

// NewRouter 创建新的路由器
func NewRouter(config *config.Config, vectorService service.VectorService) *Router {
	// 创建AI服务
	aiService := ai.NewAIService(&config.AI)
	aiService.SetVectorService(vectorService)

	// 创建处理器
	aiHandler := NewAIHandler()
	aiHandler.SetAIService(aiService)

	return &Router{
		config:           config,
		knowledgeHandler: NewKnowledgeHandler(vectorService),
		aiHandler:        aiHandler,
		categoryHandler:  NewCategoryHandler(),
		tagHandler:       NewTagHandler(),
		vectorService:    vectorService,
	}
}

// SetupRoutes 设置路由
func (r *Router) SetupRoutes() *gin.Engine {
	// 设置Gin模式
	gin.SetMode(r.config.Server.Mode)

	// 创建路由引擎
	router := gin.New()

	// 添加全局中间件
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.ValidateRequest())

	// CORS配置
	router.Use(middleware.CORS(
		r.config.CORS.AllowedOrigins,
		r.config.CORS.AllowedMethods,
		r.config.CORS.AllowedHeaders,
	))

	// 健康检查端点
	router.GET("/health", r.healthCheck)
	router.GET("/debug/config", r.debugConfig)

	// API版本分组
	v1 := router.Group("/api/v1")
	{
		// 知识库相关路由
		knowledge := v1.Group("/knowledge")
		{
			knowledge.GET("", r.knowledgeHandler.GetKnowledges)
			knowledge.GET("/:id", r.knowledgeHandler.GetKnowledge)
			knowledge.POST("", r.knowledgeHandler.CreateKnowledge)
			knowledge.PUT("/:id", r.knowledgeHandler.UpdateKnowledge)
			knowledge.DELETE("/:id", r.knowledgeHandler.DeleteKnowledge)
			knowledge.GET("/search", r.knowledgeHandler.SearchKnowledges)
			knowledge.GET("/:id/related", r.knowledgeHandler.GetRelatedKnowledges)
			knowledge.POST("/:id/view", r.knowledgeHandler.IncrementViewCount)
		}

		// 分类相关路由
		categories := v1.Group("/categories")
		{
			categories.GET("", r.categoryHandler.GetCategories)
			categories.GET("/:id", r.categoryHandler.GetCategory)
			categories.POST("", r.categoryHandler.CreateCategory)
			categories.PUT("/:id", r.categoryHandler.UpdateCategory)
			categories.DELETE("/:id", r.categoryHandler.DeleteCategory)
			categories.GET("/:id/knowledges", r.categoryHandler.GetCategoryKnowledges)
		}

		// 标签相关路由
		tags := v1.Group("/tags")
		{
			tags.GET("", r.tagHandler.GetTags)
			tags.GET("/:id", r.tagHandler.GetTag)
			tags.POST("", r.tagHandler.CreateTag)
			tags.PUT("/:id", r.tagHandler.UpdateTag)
			tags.DELETE("/:id", r.tagHandler.DeleteTag)
			tags.GET("/:id/knowledges", r.tagHandler.GetTagKnowledges)
			tags.GET("/popular", r.tagHandler.GetPopularTags)
		}

		// AI查询相关路由
		ai := v1.Group("/ai")
		{
			ai.POST("/query", r.aiHandler.Query)
			ai.GET("/history", r.aiHandler.GetQueryHistory)
			ai.DELETE("/history/:id", r.aiHandler.DeleteQueryHistory)
			ai.GET("/history/stats", r.aiHandler.GetQueryStats)
			ai.POST("/feedback", r.aiHandler.SubmitFeedback)
			ai.GET("/models", r.aiHandler.GetModels)
		}

		// 统计相关路由
		stats := v1.Group("/stats")
		{
			stats.GET("/overview", r.getOverviewStats)
			stats.GET("/knowledge", r.getKnowledgeStats)
			stats.GET("/queries", r.getQueryStats)
		}

		// 文件上传路由
		files := v1.Group("/files")
		{
			files.POST("/upload", r.uploadFile)
		}
	}

	// 404处理
	router.NoRoute(func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotFound, "API endpoint not found")
	})

	return router
}

// healthCheck 健康检查
func (r *Router) healthCheck(c *gin.Context) {
	// 检查数据库连接
	db := database.GetDatabase()
	sqlDB, err := db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "database ping failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

// debugConfig 调试配置信息
func (r *Router) debugConfig(c *gin.Context) {
	// 只返回安全的配置信息（不包含敏感信息）
	config := gin.H{
		"server": gin.H{
			"host": r.config.Server.Host,
			"port": r.config.Server.Port,
			"mode": r.config.Server.Mode,
		},
		"database": gin.H{
			"type": r.config.Database.Type,
			"host": r.config.Database.Host,
			"port": r.config.Database.Port,
		},
		"ai": gin.H{
			"provider": r.config.AI.Provider,
			"openai": gin.H{
				"base_url": r.config.AI.OpenAI.BaseURL,
				"model":    r.config.AI.OpenAI.Model,
				"has_key":  r.config.AI.OpenAI.APIKey != "",
			},
			"claude": gin.H{
				"base_url": r.config.AI.Claude.BaseURL,
				"model":    r.config.AI.Claude.Model,
				"has_key":  r.config.AI.Claude.APIKey != "",
			},
		},
	}

	utils.SuccessResponse(c, config)
}

// getOverviewStats 获取概览统计
func (r *Router) getOverviewStats(c *gin.Context) {
	db := database.GetDatabase()

	var knowledgeCount, categoryCount, tagCount, queryCount int64

	// 统计知识条目数量
	db.Model(&models.Knowledge{}).Count(&knowledgeCount)

	// 统计分类数量
	db.Model(&models.Category{}).Count(&categoryCount)

	// 统计标签数量
	db.Model(&models.Tag{}).Count(&tagCount)

	// 统计查询数量
	db.Model(&models.QueryHistory{}).Count(&queryCount)

	stats := gin.H{
		"knowledge_count": knowledgeCount,
		"category_count":  categoryCount,
		"tag_count":       tagCount,
		"query_count":     queryCount,
	}

	utils.SuccessResponse(c, stats)
}

// getKnowledgeStats 获取知识库统计
func (r *Router) getKnowledgeStats(c *gin.Context) {
	db := database.GetDatabase()

	// 按分类统计
	var categoryStats []struct {
		CategoryID uint   `json:"category_id"`
		CategoryName string `json:"category_name"`
		Count      int64  `json:"count"`
	}

	db.Table("knowledges").
		Select("category_id, categories.name as category_name, count(*) as count").
		Joins("left join categories on knowledges.category_id = categories.id").
		Group("category_id, categories.name").
		Scan(&categoryStats)

	// 按标签统计
	var tagStats []struct {
		TagID      uint   `json:"tag_id"`
		TagName    string `json:"tag_name"`
		Count      int64  `json:"count"`
	}

	db.Table("tags").
		Select("tags.id as tag_id, tags.name as tag_name, count(knowledge_tags.tag_id) as count").
		Joins("left join knowledge_tags on tags.id = knowledge_tags.tag_id").
		Group("tags.id, tags.name").
		Order("count desc").
		Limit(10).
		Scan(&tagStats)

	stats := gin.H{
		"by_category": categoryStats,
		"by_tags":     tagStats,
	}

	utils.SuccessResponse(c, stats)
}

// getQueryStats 获取查询统计
func (r *Router) getQueryStats(c *gin.Context) {
	db := database.GetDatabase()

	// 今日查询数量
	var todayCount int64
	today := time.Now().Truncate(24 * time.Hour)
	db.Model(&models.QueryHistory{}).
		Where("created_at >= ?", today).
		Count(&todayCount)

	// 本周查询数量
	var weekCount int64
	weekStart := time.Now().AddDate(0, 0, -7)
	db.Model(&models.QueryHistory{}).
		Where("created_at >= ?", weekStart).
		Count(&weekCount)

	// 查询成功率
	var successCount, totalCount int64
	db.Model(&models.QueryHistory{}).Count(&totalCount)
	db.Model(&models.QueryHistory{}).Where("is_success = ?", true).Count(&successCount)

	successRate := float64(0)
	if totalCount > 0 {
		successRate = float64(successCount) / float64(totalCount) * 100
	}

	// 最常用的查询词
	var popularQueries []struct {
		Query string `json:"query"`
		Count int64  `json:"count"`
	}

	db.Model(&models.QueryHistory{}).
		Select("query, count(*) as count").
		Where("is_success = ?", true).
		Group("query").
		Order("count desc").
		Limit(10).
		Scan(&popularQueries)

	stats := gin.H{
		"today_count":    todayCount,
		"week_count":     weekCount,
		"total_count":    totalCount,
		"success_rate":   successRate,
		"popular_queries": popularQueries,
	}

	utils.SuccessResponse(c, stats)
}

// uploadFile 文件上传处理
func (r *Router) uploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No file uploaded")
		return
	}

	// 检查文件大小（限制为10MB）
	if file.Size > 10*1024*1024 {
		utils.ErrorResponse(c, http.StatusBadRequest, "File too large (max 10MB)")
		return
	}

	// 保存文件
	filename, err := utils.SaveUploadedFile(file, "uploads")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file")
		return
	}

	result := gin.H{
		"filename": filename,
		"size":     file.Size,
		"mime_type": file.Header.Get("Content-Type"),
		"url":      "/uploads/" + filename,
	}

	utils.SuccessResponse(c, result)
}