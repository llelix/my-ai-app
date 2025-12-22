package core

import (
	"errors"
	"fmt"
)

var (
	// ErrDocumentNotFound 文档未找到错误
	ErrDocumentNotFound = errors.New("document not found")

	// ErrInvalidDocumentFormat 无效文档格式错误
	ErrInvalidDocumentFormat = errors.New("invalid document format")

	// ErrProcessingFailed 处理失败错误
	ErrProcessingFailed = errors.New("document processing failed")

	// ErrInvalidConfiguration 无效配置错误
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrTaskNotFound 任务未找到错误
	ErrTaskNotFound = errors.New("task not found")

	// ErrTaskCancelled 任务已取消错误
	ErrTaskCancelled = errors.New("task cancelled")

	// ErrQueueFull 队列已满错误
	ErrQueueFull = errors.New("processing queue is full")
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Value   any
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// NewValidationError 创建验证错误
func NewValidationError(field string, value any, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// ProcessingError 处理错误
type ProcessingError struct {
	DocumentID string
	Stage      string
	Err        error
}

func (e *ProcessingError) Error() string {
	return fmt.Sprintf("processing error for document '%s' at stage '%s': %v", e.DocumentID, e.Stage, e.Err)
}

func (e *ProcessingError) Unwrap() error {
	return e.Err
}

// NewProcessingError 创建处理错误
func NewProcessingError(documentID, stage string, err error) *ProcessingError {
	return &ProcessingError{
		DocumentID: documentID,
		Stage:      stage,
		Err:        err,
	}
}
