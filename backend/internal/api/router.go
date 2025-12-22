package api

import (
	"net/http"
	"time"

	"ai-knowledge-app/internal/ai"
	"ai-knowledge-app/internal/config"
	"ai-knowledge-app/internal/middleware"
	"ai-knowledge-app/internal/service"
	"ai-knowledge-app/pkg/database"
	"ai-knowledge-app/pkg/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router API路由器
type Router struct {
	config           *config.Config
	knowledgeHandler *KnowledgeHandler
	aiHandler        *AIHandler
	tagHandler       *TagHandler
	documentHandler  *DocumentHandler
	vectorService    service.VectorService
}

// NewRouter 创建新的路由器
func NewRouter(config *config.Config, vectorService service.VectorService, minioClient *service.MinIOClient) *Router {
	// 创建AI服务
	aiService := ai.NewAIService(&config.AI)
	aiService.SetVectorService(vectorService)

	// 创建文档服务
	documentService := service.NewDocumentService(database.GetDatabase())
	if minioClient != nil {
		documentService.SetMinIOClient(minioClient)
	}

	// 创建处理器
	aiHandler := NewAIHandler()
	aiHandler.SetAIService(aiService)

	return &Router{
		config:           config,
		knowledgeHandler: NewKnowledgeHandler(vectorService),
		aiHandler:        aiHandler,
		tagHandler:       NewTagHandler(),
		documentHandler:  NewDocumentHandler(documentService),
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

	// Swagger文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

		// 文档管理路由
		documents := v1.Group("/documents")
		{
			documents.GET("/check", r.documentHandler.CheckFile)
			documents.POST("/upload", r.documentHandler.Upload)
			documents.GET("", r.documentHandler.List)
			documents.PUT("/:id/description", r.documentHandler.UpdateDescription)
			documents.GET("/:id/download", r.documentHandler.Download)
			documents.GET("/:id", r.documentHandler.Get)
			documents.DELETE("/:id", r.documentHandler.Delete)
			documents.POST("/:id/preprocess", r.documentHandler.Preprocess)
		}
	}

	// 404处理
	router.NoRoute(func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotFound, "API endpoint not found")
	})

	return router
}

// healthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务和数据库连接状态
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /health [get]
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
