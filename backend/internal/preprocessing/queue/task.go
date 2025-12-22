package queue

import (
	"time"

	"ai-knowledge-app/internal/preprocessing/core"
)

// Task 处理任务
type Task struct {
	ID          string                    `json:"id"`
	DocumentID  string                    `json:"document_id"`
	Type        TaskType                  `json:"type"`
	Status      core.ProcessingStatusType `json:"status"`
	Priority    int                       `json:"priority"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	StartedAt   *time.Time                `json:"started_at,omitempty"`
	CompletedAt *time.Time                `json:"completed_at,omitempty"`
	Error       string                    `json:"error,omitempty"`
	Retries     int                       `json:"retries"`
	MaxRetries  int                       `json:"max_retries"`
}

// TaskType 任务类型
type TaskType string

const (
	TaskTypeProcess   TaskType = "process"
	TaskTypeReprocess TaskType = "reprocess"
	TaskTypeBatch     TaskType = "batch"
	TaskTypeVectorize TaskType = "vectorize" // 预留
)

// NewTask 创建新任务
func NewTask(documentID string, taskType TaskType, priority int) *Task {
	return &Task{
		ID:         core.GenerateID(),
		DocumentID: documentID,
		Type:       taskType,
		Status:     core.StatusPending,
		Priority:   priority,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		MaxRetries: 3,
	}
}

// Start 开始任务
func (t *Task) Start() {
	now := time.Now()
	t.Status = core.StatusProcessing
	t.StartedAt = &now
	t.UpdatedAt = now
}

// Complete 完成任务
func (t *Task) Complete() {
	now := time.Now()
	t.Status = core.StatusCompleted
	t.CompletedAt = &now
	t.UpdatedAt = now
}

// Fail 任务失败
func (t *Task) Fail(err error) {
	t.Status = core.StatusFailed
	t.Error = err.Error()
	t.UpdatedAt = time.Now()
}

// CanRetry 是否可以重试
func (t *Task) CanRetry() bool {
	return t.Retries < t.MaxRetries
}

// Retry 重试任务
func (t *Task) Retry() {
	t.Retries++
	t.Status = core.StatusPending
	t.Error = ""
	t.UpdatedAt = time.Now()
	t.StartedAt = nil
	t.CompletedAt = nil
}

// Duration 获取处理时长
func (t *Task) Duration() time.Duration {
	if t.StartedAt == nil {
		return 0
	}

	endTime := time.Now()
	if t.CompletedAt != nil {
		endTime = *t.CompletedAt
	}

	return endTime.Sub(*t.StartedAt)
}
