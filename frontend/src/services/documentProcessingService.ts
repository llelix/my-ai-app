import { apiService } from './api';
import type {
  DocumentProcessingStatus,
  BatchProcessingRequest,
  BatchProcessingStatus,
  ProcessingStatusType
} from '../types';

/**
 * 文档处理状态API服务
 * 提供文档预处理和向量化的状态查询、触发和监控功能
 */
export class DocumentProcessingService {
  private pollingIntervals: Map<string, number> = new Map();
  private readonly POLLING_INTERVAL = 2000; // 2秒轮询间隔

  /**
   * 触发文档预处理
   * @param documentId 文档ID
   * @returns 处理状态
   */
  async triggerPreprocessing(documentId: string): Promise<DocumentProcessingStatus> {
    // 使用新的预处理API端点
    const response = await apiService.post<{ message: string; document_id: string; status: string }>(
      `/processing/documents/${documentId}/process`
    );
    
    if (!response.data) {
      throw new Error('Failed to trigger document preprocessing');
    }
    
    // 转换为旧的格式以保持兼容性
    return {
      documentId: response.data.document_id,
      preprocessStatus: 'processing',
      vectorizationStatus: 'not_started',
      progress: 0,
      vectorizationProgress: 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
  }

  /**
   * 触发文档向量化（预留接口）
   * @param documentId 文档ID
   * @returns 处理状态
   */
  async triggerVectorization(documentId: string): Promise<DocumentProcessingStatus> {
    const response = await apiService.post<DocumentProcessingStatus>(
      `/documents/${documentId}/vectorize`
    );
    
    if (!response.data) {
      throw new Error('Failed to trigger document vectorization');
    }
    
    return response.data;
  }

  /**
   * 获取文档处理状态
   * @param documentId 文档ID
   * @returns 处理状态
   */
  async getProcessingStatus(documentId: string): Promise<DocumentProcessingStatus> {
    // 尝试使用新的预处理API端点
    try {
      const response = await apiService.get<{
        document_id: string;
        status: string;
        progress: number;
        error_message?: string;
        started_at?: string;
        completed_at?: string;
        processed_size: number;
        total_size: number;
      }>(`/processing/documents/${documentId}/status`);
      
      if (response.data) {
        // 转换为旧的格式以保持兼容性
        return {
          documentId: response.data.document_id,
          preprocessStatus: response.data.status as any,
          vectorizationStatus: 'not_started',
          progress: response.data.progress,
          vectorizationProgress: 0,
          error: response.data.error_message,
          createdAt: response.data.started_at || new Date().toISOString(),
          updatedAt: new Date().toISOString(),
          completedAt: response.data.completed_at
        };
      }
    } catch (error) {
      // 如果新API不可用，返回默认状态
      console.warn('New processing API not available, returning default status');
    }
    
    // 返回默认状态
    return {
      documentId,
      preprocessStatus: 'not_started',
      vectorizationStatus: 'not_started',
      progress: 0,
      vectorizationProgress: 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
  }

  /**
   * 获取文档向量化状态（预留接口）
   * @param documentId 文档ID
   * @returns 处理状态
   */
  async getVectorizationStatus(documentId: string): Promise<DocumentProcessingStatus> {
    const response = await apiService.get<DocumentProcessingStatus>(
      `/documents/${documentId}/vectorization-status`
    );
    
    if (!response.data) {
      throw new Error('Failed to get vectorization status');
    }
    
    return response.data;
  }

  /**
   * 批量触发文档预处理
   * @param documentIds 文档ID列表
   * @returns 批量处理状态
   */
  async batchTriggerPreprocessing(documentIds: string[]): Promise<BatchProcessingStatus> {
    const request: BatchProcessingRequest = { documentIds };
    const response = await apiService.post<BatchProcessingStatus>(
      '/documents/batch-preprocess',
      request
    );
    
    if (!response.data) {
      throw new Error('Failed to trigger batch preprocessing');
    }
    
    return response.data;
  }

  /**
   * 获取批量处理状态
   * @param batchId 批量处理ID
   * @returns 批量处理状态
   */
  async getBatchProcessingStatus(batchId: string): Promise<BatchProcessingStatus> {
    const response = await apiService.get<BatchProcessingStatus>(
      `/documents/batch-processing-status?batchId=${batchId}`
    );
    
    if (!response.data) {
      throw new Error('Failed to get batch processing status');
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
    callback: (status: DocumentProcessingStatus) => void,
    options: {
      interval?: number;
      maxAttempts?: number;
      stopOnComplete?: boolean;
    } = {}
  ): void {
    const {
      interval = this.POLLING_INTERVAL,
      maxAttempts = 300, // 默认最多轮询5分钟 (300 * 2秒)
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
          (stopOnComplete && this.isProcessingComplete(status));

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
            documentId,
            preprocessStatus: 'failed',
            vectorizationStatus: 'not_started',
            progress: 0,
            vectorizationProgress: 0,
            error: 'Failed to fetch processing status',
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
          });
        }
      }
    };

    // 立即执行一次
    pollFunction();

    // 设置定时器
    const intervalId = setInterval(pollFunction, interval);
    this.pollingIntervals.set(documentId, intervalId);
  }

  /**
   * 停止轮询文档处理状态
   * @param documentId 文档ID
   */
  stopPolling(documentId: string): void {
    const intervalId = this.pollingIntervals.get(documentId);
    if (intervalId) {
      clearInterval(intervalId);
      this.pollingIntervals.delete(documentId);
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
   * @param status 处理状态
   * @returns 是否完成
   */
  private isProcessingComplete(status: DocumentProcessingStatus): boolean {
    const preprocessComplete = ['completed', 'failed'].includes(status.preprocessStatus);
    const vectorizationComplete = ['completed', 'failed', 'not_started'].includes(status.vectorizationStatus);
    
    return preprocessComplete && vectorizationComplete;
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
   * 重试失败的处理
   * @param documentId 文档ID
   * @param processType 处理类型
   * @returns 处理状态
   */
  async retryProcessing(
    documentId: string, 
    processType: 'preprocess' | 'vectorize'
  ): Promise<DocumentProcessingStatus> {
    if (processType === 'preprocess') {
      return this.triggerPreprocessing(documentId);
    } else {
      return this.triggerVectorization(documentId);
    }
  }

  /**
   * 取消处理（如果支持）
   * @param documentId 文档ID
   * @param processType 处理类型
   */
  async cancelProcessing(
    documentId: string, 
    processType: 'preprocess' | 'vectorize'
  ): Promise<void> {
    // 停止轮询
    this.stopPolling(documentId);
    
    // 发送取消请求（如果后端支持）
    try {
      await apiService.post(`/documents/${documentId}/${processType}/cancel`);
    } catch (error) {
      console.warn(`Cancel ${processType} not supported or failed:`, error);
    }
  }

  /**
   * 清理资源（组件卸载时调用）
   */
  cleanup(): void {
    this.stopAllPolling();
  }
}

// 创建服务实例
export const documentProcessingService = new DocumentProcessingService();

// 默认导出
export default documentProcessingService;