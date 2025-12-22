package core

import (
	"crypto/rand"
	"encoding/hex"
	"slices"
)

// GenerateID 生成唯一ID
func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ValidateDocumentChunk 验证文档块数据
func ValidateDocumentChunk(chunk *DocumentChunk) error {
	if chunk.DocumentID == "" {
		return NewValidationError("document_id", chunk.DocumentID, "document ID is required")
	}

	if chunk.Content == "" {
		return NewValidationError("content", chunk.Content, "content is required")
	}

	if chunk.ChunkIndex < 0 {
		return NewValidationError("chunk_index", chunk.ChunkIndex, "chunk index must be non-negative")
	}

	if chunk.StartOffset < 0 {
		return NewValidationError("start_offset", chunk.StartOffset, "start offset must be non-negative")
	}

	if chunk.EndOffset <= chunk.StartOffset {
		return NewValidationError("end_offset", chunk.EndOffset, "end offset must be greater than start offset")
	}

	return nil
}

// ValidateProcessingStatus 验证处理状态数据
func ValidateProcessingStatus(status *ProcessingStatus) error {
	if status.DocumentID == "" {
		return NewValidationError("document_id", status.DocumentID, "document ID is required")
	}

	validStatuses := []ProcessingStatusType{
		StatusPending, StatusProcessing, StatusCompleted, StatusFailed, StatusNotStarted,
	}

	// 验证预处理状态
	if !slices.Contains(validStatuses, status.PreprocessStatus) {
		return NewValidationError("preprocess_status", status.PreprocessStatus, "invalid preprocess status")
	}

	// 验证向量化状态
	if !slices.Contains(validStatuses, status.VectorizationStatus) {
		return NewValidationError("vectorization_status", status.VectorizationStatus, "invalid vectorization status")
	}

	if status.Progress < 0 || status.Progress > 100 {
		return NewValidationError("progress", status.Progress, "progress must be between 0 and 100")
	}

	if status.VectorizationProgress < 0 || status.VectorizationProgress > 100 {
		return NewValidationError("vectorization_progress", status.VectorizationProgress, "vectorization progress must be between 0 and 100")
	}

	return nil
}
