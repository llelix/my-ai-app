package queue

import (
	"context"
	"sync"
	"time"

	"ai-knowledge-app/internal/preprocessing/core"
)

// ProcessingQueue 处理队列
type ProcessingQueue struct {
	tasks       chan *Task
	workers     int
	service     core.DocumentPreprocessingService
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex
	activeTasks map[string]*Task
	metrics     *QueueMetrics
	running     bool
}

// NewProcessingQueue 创建新的处理队列
func NewProcessingQueue(service core.DocumentPreprocessingService, workers, queueSize int) *ProcessingQueue {
	ctx, cancel := context.WithCancel(context.Background())

	return &ProcessingQueue{
		tasks:       make(chan *Task, queueSize),
		workers:     workers,
		service:     service,
		ctx:         ctx,
		cancel:      cancel,
		activeTasks: make(map[string]*Task),
		metrics:     NewQueueMetrics(),
	}
}

// Start 启动队列处理
func (q *ProcessingQueue) Start() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running {
		return
	}

	q.running = true

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
}

// Stop 停止队列处理
func (q *ProcessingQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.running {
		return
	}

	q.running = false
	q.cancel()
	close(q.tasks)
	q.wg.Wait()
}

// AddTask 添加任务到队列
func (q *ProcessingQueue) AddTask(task *Task) error {
	select {
	case q.tasks <- task:
		q.metrics.IncrementTotal()
		q.metrics.UpdateQueueSize(len(q.tasks))
		return nil
	case <-q.ctx.Done():
		return core.ErrTaskCancelled
	default:
		return core.ErrQueueFull
	}
}

// GetTask 获取任务状态
func (q *ProcessingQueue) GetTask(taskID string) (*Task, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if task, exists := q.activeTasks[taskID]; exists {
		return task, nil
	}

	return nil, core.ErrTaskNotFound
}

// GetMetrics 获取队列指标
func (q *ProcessingQueue) GetMetrics() QueueMetrics {
	return q.metrics.GetSnapshot()
}

// worker 工作协程
func (q *ProcessingQueue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case task, ok := <-q.tasks:
			if !ok {
				return
			}

			q.processTask(task)

		case <-q.ctx.Done():
			return
		}
	}
}

// processTask 处理单个任务
func (q *ProcessingQueue) processTask(task *Task) {
	// 记录活跃任务
	q.mu.Lock()
	q.activeTasks[task.ID] = task
	q.mu.Unlock()

	// 开始处理
	task.Start()

	// 更新指标
	activeCount := len(q.activeTasks)
	q.metrics.UpdateWorkerCount(activeCount, q.workers)

	// 执行任务
	err := q.executeTask(task)

	// 处理结果
	if err != nil {
		task.Fail(err)
		q.metrics.IncrementFailed()

		// 重试逻辑
		if task.CanRetry() {
			task.Retry()
			q.metrics.IncrementRetried()

			// 延迟后重新加入队列
			go func() {
				time.Sleep(30 * time.Second)
				q.AddTask(task)
			}()
		}
	} else {
		task.Complete()
		q.metrics.IncrementCompleted(task.Duration())
	}

	// 移除活跃任务
	q.mu.Lock()
	delete(q.activeTasks, task.ID)
	q.mu.Unlock()

	// 更新队列大小
	q.metrics.UpdateQueueSize(len(q.tasks))
}

// executeTask 执行具体任务
func (q *ProcessingQueue) executeTask(task *Task) error {
	ctx, cancel := context.WithTimeout(q.ctx, 10*time.Minute)
	defer cancel()

	switch task.Type {
	case TaskTypeProcess:
		return q.service.ProcessDocument(ctx, task.DocumentID)
	case TaskTypeReprocess:
		return q.service.ReprocessDocument(ctx, task.DocumentID)
	default:
		return core.NewProcessingError(task.DocumentID, "execute", core.ErrInvalidConfiguration)
	}
}
