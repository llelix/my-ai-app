package api

import (
	"ai-knowledge-app/internal/preprocessing/core"
	"ai-knowledge-app/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProcessingHandler 文档预处理处理器
type ProcessingHandler struct {
	service core.DocumentPreprocessingService
}

// NewProcessingHandler 创建新的预处理处理器
func NewProcessingHandler(service core.DocumentPreprocessingService) *ProcessingHandler {
	return &ProcessingHandler{
		service: service,
	}
}

// ProcessDocument 处理文档
// @Summary 处理文档
// @Description 启动文档预处理任务，将文档转换为可搜索的格式
// @Tags processing
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} map[string]interface{} "处理任务已启动"
// @Failure 400 {object} map[string]interface{} "无效的文档ID"
// @Failure 404 {object} map[string]interface{} "文档未找到"
// @Failure 500 {object} map[string]interface{} "处理失败"
// @Router /api/v1/processing/documents/{id}/process [post]
func (h *ProcessingHandler) ProcessDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	documentID := strconv.FormatUint(id, 10)
	err = h.service.ProcessDocument(c.Request.Context(), documentID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process document: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":     "Document processing started successfully",
		"document_id": documentID,
		"status":      "processing",
	})
}

// ProcessDocumentAsync 异步处理文档
// @Summary 异步处理文档
// @Description 启动异步文档预处理任务，返回任务ID用于跟踪进度
// @Tags processing
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Param request body ProcessDocumentAsyncRequest false "处理选项"
// @Success 200 {object} ProcessDocumentAsyncResponse "异步任务已创建"
// @Failure 400 {object} map[string]interface{} "无效的文档ID或请求参数"
// @Failure 404 {object} map[string]interface{} "文档未找到"
// @Failure 500 {object} map[string]interface{} "任务创建失败"
// @Router /api/v1/processing/documents/{id}/process-async [post]
func (h *ProcessingHandler) ProcessDocumentAsync(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	var req ProcessDocumentAsyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有请求体，使用默认优先级
		req.Priority = 1
	}

	documentID := strconv.FormatUint(id, 10)
	task, err := h.service.ProcessDocumentAsync(documentID, req.Priority)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create processing task: "+err.Error())
		return
	}

	response := ProcessDocumentAsyncResponse{
		TaskID:     task.ID,
		DocumentID: documentID,
		Status:     string(task.Status),
		Priority:   task.Priority,
		CreatedAt:  task.CreatedAt,
	}

	utils.SuccessResponse(c, response)
}

// GetProcessingStatus 获取处理状态
// @Summary 获取文档处理状态
// @Description 获取指定文档的预处理状态信息
// @Tags processing
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} ProcessingStatusResponse "处理状态信息"
// @Failure 400 {object} map[string]interface{} "无效的文档ID"
// @Failure 404 {object} map[string]interface{} "处理状态未找到"
// @Failure 500 {object} map[string]interface{} "获取状态失败"
// @Router /api/v1/processing/documents/{id}/status [get]
func (h *ProcessingHandler) GetProcessingStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	documentID := strconv.FormatUint(id, 10)
	status, err := h.service.GetProcessingStatus(c.Request.Context(), documentID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Processing status not found: "+err.Error())
		return
	}

	response := ProcessingStatusResponse{
		DocumentID:    status.DocumentID,
		Status:        string(status.PreprocessStatus),
		Progress:      status.Progress,
		ErrorMessage:  status.Error,
		StartedAt:     &status.CreatedAt,
		CompletedAt:   status.CompletedAt,
		ProcessedSize: 0, // 需要从其他地方获取
		TotalSize:     0, // 需要从其他地方获取
	}

	utils.SuccessResponse(c, response)
}

// GetTaskStatus 获取任务状态
// @Summary 获取处理任务状态
// @Description 根据任务ID获取异步处理任务的详细状态
// @Tags processing
// @Accept json
// @Produce json
// @Param taskId path string true "任务ID"
// @Success 200 {object} TaskStatusResponse "任务状态信息"
// @Failure 400 {object} map[string]interface{} "无效的任务ID"
// @Failure 404 {object} map[string]interface{} "任务未找到"
// @Failure 500 {object} map[string]interface{} "获取任务状态失败"
// @Router /api/v1/processing/tasks/{taskId}/status [get]
func (h *ProcessingHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Task ID is required")
		return
	}

	task, err := h.service.GetTaskStatus(taskID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Task not found: "+err.Error())
		return
	}

	response := TaskStatusResponse{
		TaskID:       task.ID,
		DocumentID:   task.DocumentID,
		Status:       string(task.Status),
		Priority:     task.Priority,
		Progress:     0, // ProcessingTask 没有 Progress 字段
		ErrorMessage: task.Error,
		CreatedAt:    task.CreatedAt,
		StartedAt:    nil, // ProcessingTask 没有 StartedAt 字段
		CompletedAt:  nil, // ProcessingTask 没有 CompletedAt 字段
	}

	utils.SuccessResponse(c, response)
}

// CancelTask 取消处理任务
// @Summary 取消处理任务
// @Description 取消指定的异步处理任务
// @Tags processing
// @Accept json
// @Produce json
// @Param taskId path string true "任务ID"
// @Success 200 {object} map[string]interface{} "任务已取消"
// @Failure 400 {object} map[string]interface{} "无效的任务ID"
// @Failure 404 {object} map[string]interface{} "任务未找到"
// @Failure 500 {object} map[string]interface{} "取消任务失败"
// @Router /api/v1/processing/tasks/{taskId}/cancel [post]
func (h *ProcessingHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Task ID is required")
		return
	}

	err := h.service.CancelTask(taskID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to cancel task: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "Task cancelled successfully",
		"task_id": taskID,
	})
}

// BatchProcessDocuments 批量处理文档
// @Summary 批量处理文档
// @Description 批量启动多个文档的预处理任务
// @Tags processing
// @Accept json
// @Produce json
// @Param request body BatchProcessRequest true "批量处理请求"
// @Success 200 {object} BatchProcessResponse "批量处理结果"
// @Failure 400 {object} map[string]interface{} "无效的请求参数"
// @Failure 500 {object} map[string]interface{} "批量处理失败"
// @Router /api/v1/processing/documents/batch-process [post]
func (h *ProcessingHandler) BatchProcessDocuments(c *gin.Context) {
	var req BatchProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	if len(req.DocumentIDs) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Document IDs are required")
		return
	}

	if req.Async {
		// 异步批量处理
		tasks, err := h.service.BatchProcessDocumentsAsync(req.DocumentIDs, req.Priority)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create batch processing tasks: "+err.Error())
			return
		}

		var taskIDs []string
		for _, task := range tasks {
			taskIDs = append(taskIDs, task.ID)
		}

		response := BatchProcessResponse{
			Success:        true,
			ProcessedCount: len(req.DocumentIDs),
			TaskIDs:        taskIDs,
		}

		utils.SuccessResponse(c, response)
	} else {
		// 同步批量处理
		err := h.service.BatchProcessDocuments(c.Request.Context(), req.DocumentIDs)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to batch process documents: "+err.Error())
			return
		}

		response := BatchProcessResponse{
			Success:        true,
			ProcessedCount: len(req.DocumentIDs),
		}

		utils.SuccessResponse(c, response)
	}
}

// GetDocumentChunks 获取文档分块
// @Summary 获取文档分块
// @Description 获取指定文档预处理后的文本分块
// @Tags processing
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} DocumentChunksResponse "文档分块列表"
// @Failure 400 {object} map[string]interface{} "无效的文档ID"
// @Failure 404 {object} map[string]interface{} "文档分块未找到"
// @Failure 500 {object} map[string]interface{} "获取分块失败"
// @Router /api/v1/processing/documents/{id}/chunks [get]
func (h *ProcessingHandler) GetDocumentChunks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	documentID := strconv.FormatUint(id, 10)
	chunks, err := h.service.GetDocumentChunks(c.Request.Context(), documentID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Document chunks not found: "+err.Error())
		return
	}

	response := DocumentChunksResponse{
		DocumentID: documentID,
		ChunkCount: len(chunks),
		Chunks:     chunks,
	}

	utils.SuccessResponse(c, response)
}

// GetQueueStats 获取队列统计
// @Summary 获取处理队列统计
// @Description 获取预处理队列的统计信息，包括待处理、处理中、已完成的任务数量
// @Tags processing
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "队列统计信息"
// @Failure 500 {object} map[string]interface{} "获取统计失败"
// @Router /api/v1/processing/queue/stats [get]
func (h *ProcessingHandler) GetQueueStats(c *gin.Context) {
	stats := h.service.GetQueueStats()
	utils.SuccessResponse(c, stats)
}

// GetProcessingStatistics 获取处理统计
// @Summary 获取处理统计信息
// @Description 获取文档预处理的详细统计信息，包括成功率、平均处理时间等
// @Tags processing
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "处理统计信息"
// @Failure 500 {object} map[string]interface{} "获取统计失败"
// @Router /api/v1/processing/statistics [get]
func (h *ProcessingHandler) GetProcessingStatistics(c *gin.Context) {
	stats, err := h.service.GetProcessingStatistics(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get processing statistics: "+err.Error())
		return
	}

	utils.SuccessResponse(c, stats)
}

// GetSupportedFormats 获取支持的格式
// @Summary 获取支持的文档格式
// @Description 获取系统支持的文档格式列表
// @Tags processing
// @Accept json
// @Produce json
// @Success 200 {object} SupportedFormatsResponse "支持的格式列表"
// @Router /api/v1/processing/formats [get]
func (h *ProcessingHandler) GetSupportedFormats(c *gin.Context) {
	formats := h.service.GetSupportedFormats()

	response := SupportedFormatsResponse{
		Formats: formats,
		Count:   len(formats),
	}

	utils.SuccessResponse(c, response)
}

// ReprocessDocument 重新处理文档
// @Summary 重新处理文档
// @Description 重新启动文档预处理任务，会覆盖之前的处理结果
// @Tags processing
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} map[string]interface{} "重新处理已启动"
// @Failure 400 {object} map[string]interface{} "无效的文档ID"
// @Failure 404 {object} map[string]interface{} "文档未找到"
// @Failure 500 {object} map[string]interface{} "重新处理失败"
// @Router /api/v1/processing/documents/{id}/reprocess [post]
func (h *ProcessingHandler) ReprocessDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	documentID := strconv.FormatUint(id, 10)
	err = h.service.ReprocessDocument(c.Request.Context(), documentID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to reprocess document: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":     "Document reprocessing started successfully",
		"document_id": documentID,
		"status":      "processing",
	})
}
