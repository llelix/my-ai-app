import { apiService } from './api';
import type {
  ProcessDocumentAsyncRequest,
  ProcessDocumentAsyncResponse,
  ProcessingStatusResponse,
  TaskStatusResponse,
  BatchProcessRequest,
  BatchProcessResponse,
  DocumentChunksResponse,
  SupportedFormatsResponse,
  QueueStatsResponse,
  ProcessingStatisticsResponse,
  PollingOptions,
  ProcessingStatusType
} from '../types/processing';

/**
 * 文档预处理API服务
 * 提供完整的文档预处理功能，包括同步/异步处理、状态查询、任务管理等
 */
export class ProcessingService {
  private pollingIntervals: Map<string, number> = new Map();
  private readonly DEFAULT_POLLING_INTERVAL = 2000; // 2秒轮询间隔

  /**
   * 同步处理文档
   * @param documentId 文档ID
   * @returns 处理结果
   */
  async processDocument(documentId: string): Promise<{ message: string; document_id: string; status: string }> {
    const response = await apiService.post<{ message: string; document_id: string; status: string }>(
      `/processing/documents/${documentId}/process`
    );
    
    if (!response.data) {
      throw new Error('Failed to process document');
    }
    
    return response.data;
  }

  /**
   * 异步处理文档
   * @param documentId 文档ID
   * @param options 处理选项
   * @returns 任务信息
   */
  async processDocumentAsync(
    documentId: string, 
    options: ProcessDocumentAsyncRequest = {}
  ): Promise<ProcessDocumentAsyncResponse> {
    const response = await apiService.post<ProcessDocumentAsyncResponse>(
      `/processing/documents/${documentId}/process-async`,
      options
    );
    
    if (!response.data) {
      throw new Error('Failed to create processing task');
    }
    
    return response.data;
  }

  /**
   * 获取文档处理状态
   * @param documentId 文档ID
   * @returns 处理状态
   */
  async getProcessingStatus(documentId: string): Promise<ProcessingStatusResponse> {
    const response = await apiService.get<ProcessingStatusResponse>(
      `/processing/documents/${documentId}/status`
    );
    
    if (!response.data) {
      throw new Error('Failed to get processing status');
    }
    
    return response.data;
  }

  /**
   * 获取任务状态
   * @param taskId 任务ID
   * @returns 任务状态
   */
  async getTaskStatus(taskId: string): Promise<TaskStatusResponse> {
    const response = await apiService.get<TaskStatusResponse>(
      `/processing/tasks/${taskId}/status`
    );
    
    if (!response.data) {
      throw new Error('Failed to get task status');
    }
    
    return response.data;
  }

  /**
   * 取消任务
   * @param taskId 任务ID
   */
  async cancelTask(taskId: string): Promise<{ message: string; task_id: string }> {
    const response = await apiService.post<{ message: string; task_id: string }>(
      `/processing/tasks/${taskId}/cancel`
    );
    
    if (!response.data) {
      throw new Error('Failed to cancel task');
    }
    
    return response.data;
  }

  /**
   * 批量处理文档
   * @param request 批量处理请求
   * @returns 批量处理结果
   */
  async batchProcessDocuments(request: BatchProcessRequest): Promise<BatchProcessResponse> {
    const response = await apiService.post<BatchProcessResponse>(
      '/processing/documents/batch-process',
      request
    );
    
    if (!response.data) {
      throw new Error('Failed to batch process documents');
    }
    
    return response.data;
  }

  /**
   * 获取文档分块
   * @param documentId 文档ID
   * @returns 文档分块列表
   */
  async getDocumentChunks(documentId: string): Promise<DocumentChunksResponse> {
    const response = await apiService.get<DocumentChunksResponse>(
      `/processing/documents/${documentId}/chunks`
    );
    
    if (!response.data) {
      throw new Error('Failed to get document chunks');
    }
    
    return response.data;
  }

  /**
   * 重新处理文档
   * @param documentId 文档ID
   * @returns 处理结果
   */
  async reprocessDocument(documentId: string): Promise<{ message: string; document_id: string; status: string }> {
    const response = await apiService.post<{ message: string; document_id: string; status: string }>(
      `/processing/documents/${documentId}/reprocess`
    );
    
    if (!response.data) {
      throw new Error('Failed to reprocess document');
    }
    
    return response.data;
  }

  /**
   * 获取队列统计
   * @returns 队列统计信息
   */
  async getQueueStats(): Promise<QueueStatsResponse> {
    const response = await apiService.get<QueueStatsResponse>(
      '/processing/queue/stats'
    );
    
    if (!response.data) {
      throw new Error('Failed to get queue stats');
    }
    
    return response.data;
  }

  /**
   * 获取处理统计
   * @returns 处理统计信息
   */
  async getProcessingStatistics(): Promise<ProcessingStatisticsResponse> {
    const response = await apiService.get<ProcessingStatisticsResponse>(
      '/processing/statistics'
    );
    
    if (!response.data) {
      throw new Error('Failed to get processing statistics');
    }
    
    return response.data;
  }

  /**
   * 获取支持的格式
   * @returns 支持的格式列表
   */
  async getSupportedFormats(): Promise<SupportedFormatsResponse> {
    const response = await apiService.get<SupportedFormatsResponse>(
      '/processing/formats'
    );
    
    if (!response.data) {
      throw new Error('Failed to get supported formats');
    }
    
    return response.data;
  }

  /**
   * 开始轮询文档处理状态
   * @param documentId 文档ID
   * @param callback 状态更新回调函数
   * @param options 轮询选项
   */
  startPolling(
    documentId: string,
    callback: (status: ProcessingStatusResponse) => void,
    options: PollingOptions = {}
  ): void {
    const {
      interval = this.DEFAULT_POLLING_INTERVAL,
      maxAttempts = 300, // 默认最多轮询10分钟 (300 * 2秒)
      stopOnComplete = true
    } = options;

    // 如果已经在轮询，先停止
    this.stopPolling(documentId);

    let attempts = 0;
    const pollFunction = async () => {
      try {
        attempts++;
        const status = await this.getProcessingStatus(documentId);
        callback(status);

        // 检查是否应该停止轮询
        const shouldStop = 
          attempts >= maxAttempts ||
          (stopOnComplete && this.isProcessingComplete(status.status));

        if (shouldStop) {
          this.stopPolling(documentId);
        }
      } catch (error) {
        console.error(`Error polling status for document ${documentId}:`, error);
        
        // 错误重试逻辑：如果达到最大尝试次数，停止轮询
        if (attempts >= maxAttempts) {
          this.stopPolling(documentId);
          // 通知回调函数发生了错误
          callback({
            document_id: documentId,
            status: 'failed',
            progress: 0,
            error_message: 'Failed to fetch processing status',
            processed_size: 0,
            total_size: 0
          });
        }
      }
    };

    // 立即执行一次
    pollFunction();

    // 设置定时器
    const intervalId = window.setInterval(pollFunction, interval);
    this.pollingIntervals.set(documentId, intervalId);
  }

  /**
   * 开始轮询任务状态
   * @param taskId 任务ID
   * @param callback 状态更新回调函数
   * @param options 轮询选项
   */
  startTaskPolling(
    taskId: string,
    callback: (status: TaskStatusResponse) => void,
    options: PollingOptions = {}
  ): void {
    const {
      interval = this.DEFAULT_POLLING_INTERVAL,
      maxAttempts = 300,
      stopOnComplete = true
    } = options;

    // 如果已经在轮询，先停止
    this.stopPolling(taskId);

    let attempts = 0;
    const pollFunction = async () => {
      try {
        attempts++;
        const status = await this.getTaskStatus(taskId);
        callback(status);

        const shouldStop = 
          attempts >= maxAttempts ||
          (stopOnComplete && this.isProcessingComplete(status.status));

        if (shouldStop) {
          this.stopPolling(taskId);
        }
      } catch (error) {
        console.error(`Error polling task ${taskId}:`, error);
        
        if (attempts >= maxAttempts) {
          this.stopPolling(taskId);
        }
      }
    };

    pollFunction();
    const intervalId = window.setInterval(pollFunction, interval);
    this.pollingIntervals.set(taskId, intervalId);
  }

  /**
   * 停止轮询
   * @param id 文档ID或任务ID
   */
  stopPolling(id: string): void {
    const intervalId = this.pollingIntervals.get(id);
    if (intervalId) {
      clearInterval(intervalId);
      this.pollingIntervals.delete(id);
    }
  }

  /**
   * 停止所有轮询
   */
  stopAllPolling(): void {
    this.pollingIntervals.forEach((intervalId) => {
      clearInterval(intervalId);
    });
    this.pollingIntervals.clear();
  }

  /**
   * 检查处理是否完成
   * @param status 处理状态类型
   * @returns 是否完成
   */
  private isProcessingComplete(status: ProcessingStatusType): boolean {
    return ['completed', 'failed'].includes(status);
  }

  /**
   * 检查状态是否为处理中
   * @param status 处理状态类型
   * @returns 是否为处理中
   */
  isProcessing(status: ProcessingStatusType): boolean {
    return ['pending', 'processing'].includes(status);
  }

  /**
   * 检查状态是否为成功完成
   * @param status 处理状态类型
   * @returns 是否成功完成
   */
  isCompleted(status: ProcessingStatusType): boolean {
    return status === 'completed';
  }

  /**
   * 检查状态是否为失败
   * @param status 处理状态类型
   * @returns 是否失败
   */
  isFailed(status: ProcessingStatusType): boolean {
    return status === 'failed';
  }

  /**
   * 获取状态的显示文本
   * @param status 处理状态类型
   * @returns 显示文本
   */
  getStatusText(status: ProcessingStatusType): string {
    const statusMap: Record<ProcessingStatusType, string> = {
      'not_started': '未开始',
      'pending': '等待中',
      'processing': '处理中',
      'completed': '已完成',
      'failed': '失败'
    };
    
    return statusMap[status] || '未知状态';
  }

  /**
   * 获取状态的颜色类型（用于UI显示）
   * @param status 处理状态类型
   * @returns Ant Design 状态颜色
   */
  getStatusColor(status: ProcessingStatusType): 'default' | 'processing' | 'success' | 'error' | 'warning' {
    const colorMap: Record<ProcessingStatusType, 'default' | 'processing' | 'success' | 'error' | 'warning'> = {
      'not_started': 'default',
      'pending': 'warning',
      'processing': 'processing',
      'completed': 'success',
      'failed': 'error'
    };
    
    return colorMap[status] || 'default';
  }

  /**
   * 清理资源（组件卸载时调用）
   */
  cleanup(): void {
    this.stopAllPolling();
  }
}

// 创建服务实例
export const processingService = new ProcessingService();

// 默认导出
export default processingService;
