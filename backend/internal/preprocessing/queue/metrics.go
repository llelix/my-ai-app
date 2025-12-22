package queue

import (
	"sync"
	"time"
)

// QueueMetrics 队列指标
type QueueMetrics struct {
	mu sync.RWMutex

	// 队列统计
	QueueSize     int `json:"queue_size"`
	ActiveWorkers int `json:"active_workers"`
	TotalWorkers  int `json:"total_workers"`

	// 任务统计
	TotalTasks     int64 `json:"total_tasks"`
	CompletedTasks int64 `json:"completed_tasks"`
	FailedTasks    int64 `json:"failed_tasks"`
	RetriedTasks   int64 `json:"retried_tasks"`

	// 性能统计
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	TotalProcessingTime   time.Duration `json:"total_processing_time"`

	// 时间统计
	StartTime time.Time `json:"start_time"`
	LastReset time.Time `json:"last_reset"`
}

// NewQueueMetrics 创建新的队列指标
func NewQueueMetrics() *QueueMetrics {
	now := time.Now()
	return &QueueMetrics{
		StartTime: now,
		LastReset: now,
	}
}

// IncrementTotal 增加总任务数
func (m *QueueMetrics) IncrementTotal() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalTasks++
}

// IncrementCompleted 增加完成任务数
func (m *QueueMetrics) IncrementCompleted(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CompletedTasks++
	m.TotalProcessingTime += duration
	if m.CompletedTasks > 0 {
		m.AverageProcessingTime = m.TotalProcessingTime / time.Duration(m.CompletedTasks)
	}
}

// IncrementFailed 增加失败任务数
func (m *QueueMetrics) IncrementFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FailedTasks++
}

// IncrementRetried 增加重试任务数
func (m *QueueMetrics) IncrementRetried() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RetriedTasks++
}

// UpdateQueueSize 更新队列大小
func (m *QueueMetrics) UpdateQueueSize(size int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.QueueSize = size
}

// UpdateWorkerCount 更新工作协程数量
func (m *QueueMetrics) UpdateWorkerCount(active, total int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveWorkers = active
	m.TotalWorkers = total
}

// GetSnapshot 获取指标快照
func (m *QueueMetrics) GetSnapshot() QueueMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m
}

// Reset 重置指标
func (m *QueueMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalTasks = 0
	m.CompletedTasks = 0
	m.FailedTasks = 0
	m.RetriedTasks = 0
	m.TotalProcessingTime = 0
	m.AverageProcessingTime = 0
	m.LastReset = time.Now()
}

// GetSuccessRate 获取成功率
func (m *QueueMetrics) GetSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.TotalTasks == 0 {
		return 0
	}
	return float64(m.CompletedTasks) / float64(m.TotalTasks) * 100
}

// GetFailureRate 获取失败率
func (m *QueueMetrics) GetFailureRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.TotalTasks == 0 {
		return 0
	}
	return float64(m.FailedTasks) / float64(m.TotalTasks) * 100
}

// GetThroughput 获取吞吐量（任务/秒）
func (m *QueueMetrics) GetThroughput() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	duration := time.Since(m.StartTime)
	if duration.Seconds() == 0 {
		return 0
	}
	return float64(m.CompletedTasks) / duration.Seconds()
}
