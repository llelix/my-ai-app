package api

import (
	"ai-knowledge-app/internal/service"
	"ai-knowledge-app/pkg/utils"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	service *service.DocumentService
}

func NewDocumentHandler(service *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{service: service}
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No file uploaded")
		return
	}

	doc, err := h.service.Upload(file)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload document")
		return
	}

	utils.SuccessResponse(c, doc)
}

func (h *DocumentHandler) List(c *gin.Context) {
	docs, err := h.service.List()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch documents")
		return
	}

	utils.SuccessResponse(c, docs)
}

func (h *DocumentHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	doc, err := h.service.GetByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Document not found")
		return
	}

	utils.SuccessResponse(c, doc)
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete document")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Document deleted successfully"})
}

func (h *DocumentHandler) UpdateDescription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	var req struct {
		Description string `json:"description" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	if err := h.service.UpdateDescription(uint(id), req.Description); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update description")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Description updated successfully"})
}

func (h *DocumentHandler) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	doc, err := h.service.GetByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Document not found")
		return
	}

	// Use the new GetObject method to support both MinIO and local storage
	reader, err := h.service.GetObject(doc.FilePath)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve file")
		return
	}
	defer reader.Close()

	// Set appropriate headers
	c.Header("Content-Disposition", "attachment; filename="+doc.OriginalName)
	c.Header("Content-Type", doc.MimeType)
	c.Header("Content-Length", strconv.FormatInt(doc.FileSize, 10))

	// Stream the file content
	c.DataFromReader(http.StatusOK, doc.FileSize, doc.MimeType, reader, nil)
}

// CheckFile 检查文件是否存在（秒传）
func (h *DocumentHandler) CheckFile(c *gin.Context) {
	// Debug logging
	fmt.Printf("DEBUG: CheckFile handler called with path: %s\n", c.Request.URL.Path)
	fmt.Printf("DEBUG: Query params - hash: %s, size: %s\n", c.Query("hash"), c.Query("size"))

	hash := c.Query("hash")
	sizeStr := c.Query("size")

	if hash == "" || sizeStr == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Missing hash or size parameter")
		return
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid size parameter")
		return
	}

	doc, exists := h.service.CheckFile(hash, size)

	response := gin.H{
		"exists": exists,
	}

	if exists {
		response["document"] = doc
	}

	utils.SuccessResponse(c, response)
}

// InitUpload 初始化分块上传
func (h *DocumentHandler) InitUpload(c *gin.Context) {
	var req struct {
		FileName string `json:"file_name" binding:"required"`
		FileSize int64  `json:"file_size" binding:"required"`
		FileHash string `json:"file_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	session, err := h.service.InitUpload(req.FileName, req.FileSize, req.FileHash)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to initialize upload")
		return
	}

	utils.SuccessResponse(c, session)
}

// UploadChunk 上传分块
func (h *DocumentHandler) UploadChunk(c *gin.Context) {
	sessionID := c.Param("sessionId")
	chunkIndexStr := c.Param("chunkIndex")

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid chunk index")
		return
	}

	// Read chunk data from request body
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to read chunk data")
		return
	}

	if err := h.service.UploadChunk(sessionID, chunkIndex, data); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload chunk")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Chunk uploaded successfully"})
}

// CompleteUpload 完成上传
func (h *DocumentHandler) CompleteUpload(c *gin.Context) {
	sessionID := c.Param("sessionId")

	doc, err := h.service.CompleteUpload(sessionID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to complete upload")
		return
	}

	utils.SuccessResponse(c, doc)
}

// GetUploadProgress 获取上传进度
func (h *DocumentHandler) GetUploadProgress(c *gin.Context) {
	sessionID := c.Param("sessionId")

	session, err := h.service.GetUploadProgress(sessionID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Upload session not found")
		return
	}

	utils.SuccessResponse(c, session)
}

// Preprocess 预处理文档
func (h *DocumentHandler) Preprocess(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID")
		return
	}

	// 检查文档是否存在
	doc, err := h.service.GetByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Document not found")
		return
	}

	// 启动预处理任务
	err = h.service.StartPreprocessing(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to start preprocessing")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":     "Preprocessing started successfully",
		"document_id": doc.ID,
		"status":      "processing",
	})
}
