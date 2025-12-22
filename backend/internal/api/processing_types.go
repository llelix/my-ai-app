package api

import (
	"ai-knowledge-app/internal/preprocessing/core"
	"time"
)

// ProcessDocumentAsyncRequest 异步处理文档请求
type ProcessDocumentAsyncRequest struct {
	Priority int `json:"priority" example:"1" validate:"min=1,max=10"` // 任务优先级，1-10，数字越大优先级越高
}

// ProcessDocumentAsyncResponse 异步处理文档响应
type ProcessDocumentAsyncResponse struct {
	TaskID     string    `json:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"` // 任务ID
	DocumentID string    `json:"document_id" example:"123"`                              // 文档ID
	Status     string    `json:"status" example:"pending"`                               // 任务状态
	Priority   int       `json:"priority" example:"1"`                                   // 任务优先级
	CreatedAt  time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`              // 创建时间
}

// ProcessingStatusResponse 处理状态响应
type ProcessingStatusResponse struct {
	DocumentID    string     `json:"document_id" example:"123"`                             // 文档ID
	Status        string     `json:"status" example:"processing"`                           // 处理状态：pending, processing, completed, failed
	Progress      float64    `json:"progress" example:"75.5"`                               // 处理进度百分比
	ErrorMessage  string     `json:"error_message,omitempty" example:""`                    // 错误信息
	StartedAt     *time.Time `json:"started_at,omitempty" example:"2023-01-01T00:00:00Z"`   // 开始时间
	CompletedAt   *time.Time `json:"completed_at,omitempty" example:"2023-01-01T01:00:00Z"` // 完成时间
	ProcessedSize int64      `json:"processed_size" example:"1024000"`                      // 已处理大小（字节）
	TotalSize     int64      `json:"total_size" example:"2048000"`                          // 总大小（字节）
}

// TaskStatusResponse 任务状态响应
type TaskStatusResponse struct {
	TaskID       string     `json:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"` // 任务ID
	DocumentID   string     `json:"document_id" example:"123"`                              // 文档ID
	Status       string     `json:"status" example:"processing"`                            // 任务状态
	Priority     int        `json:"priority" example:"1"`                                   // 任务优先级
	Progress     float64    `json:"progress" example:"75.5"`                                // 处理进度百分比
	ErrorMessage string     `json:"error_message,omitempty" example:""`                     // 错误信息
	CreatedAt    time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`              // 创建时间
	StartedAt    *time.Time `json:"started_at,omitempty" example:"2023-01-01T00:00:00Z"`    // 开始时间
	CompletedAt  *time.Time `json:"completed_at,omitempty" example:"2023-01-01T01:00:00Z"`  // 完成时间
}

// BatchProcessRequest 批量处理请求
type BatchProcessRequest struct {
	DocumentIDs []string `json:"document_ids" binding:"required" example:"[\"123\",\"456\",\"789\"]"` // 文档ID列表
	Priority    int      `json:"priority" example:"1" validate:"min=1,max=10"`                        // 任务优先级
	Async       bool     `json:"async" example:"true"`                                                // 是否异步处理
}

// BatchProcessResponse 批量处理响应
type BatchProcessResponse struct {
	Success        bool     `json:"success" example:"true"`                                             // 是否成功
	ProcessedCount int      `json:"processed_count" example:"3"`                                        // 处理的文档数量
	TaskIDs        []string `json:"task_ids,omitempty" example:"[\"task1\",\"task2\",\"task3\"]"`       // 任务ID列表（异步模式）
	FailedIDs      []string `json:"failed_ids,omitempty" example:"[\"456\"]"`                           // 失败的文档ID列表
	ErrorMessage   string   `json:"error_message,omitempty" example:"Some documents failed to process"` // 错误信息
}

// DocumentChunksResponse 文档分块响应
type DocumentChunksResponse struct {
	DocumentID string               `json:"document_id" example:"123"` // 文档ID
	ChunkCount int                  `json:"chunk_count" example:"5"`   // 分块数量
	Chunks     []core.DocumentChunk `json:"chunks"`                    // 分块列表
}

// SupportedFormatsResponse 支持格式响应
type SupportedFormatsResponse struct {
	Formats []string `json:"formats" example:"[\"pdf\",\"docx\",\"txt\",\"md\"]"` // 支持的格式列表
	Count   int      `json:"count" example:"4"`                                   // 格式数量
}

// QueueStatsResponse 队列统计响应
type QueueStatsResponse struct {
	PendingTasks    int     `json:"pending_tasks" example:"5"`        // 待处理任务数
	ProcessingTasks int     `json:"processing_tasks" example:"2"`     // 处理中任务数
	CompletedTasks  int     `json:"completed_tasks" example:"100"`    // 已完成任务数
	FailedTasks     int     `json:"failed_tasks" example:"3"`         // 失败任务数
	TotalTasks      int     `json:"total_tasks" example:"110"`        // 总任务数
	AverageWaitTime float64 `json:"average_wait_time" example:"30.5"` // 平均等待时间（秒）
	WorkerCount     int     `json:"worker_count" example:"4"`         // 工作协程数
}

// ProcessingStatisticsResponse 处理统计响应
type ProcessingStatisticsResponse struct {
	TotalDocuments       int     `json:"total_documents" example:"1000"`            // 总文档数
	ProcessedDocuments   int     `json:"processed_documents" example:"950"`         // 已处理文档数
	FailedDocuments      int     `json:"failed_documents" example:"50"`             // 处理失败文档数
	ProcessingRate       float64 `json:"processing_rate" example:"95.0"`            // 处理成功率（百分比）
	AverageProcessTime   float64 `json:"average_process_time" example:"120.5"`      // 平均处理时间（秒）
	TotalProcessedSize   int64   `json:"total_processed_size" example:"1073741824"` // 总处理大小（字节）
	ProcessingThroughput float64 `json:"processing_throughput" example:"8388608"`   // 处理吞吐量（字节/秒）
}
