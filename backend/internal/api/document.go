package api

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"ai-knowledge-app/internal/service"
	"ai-knowledge-app/pkg/utils"
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

	c.File(doc.FilePath)
}


