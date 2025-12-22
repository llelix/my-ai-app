import { useState, useEffect, useCallback, useRef } from 'react';
import { message } from 'antd';
import { processingService } from '../services';
import type {
  ProcessingStatusResponse,
  TaskStatusResponse,
  ProcessingStatusType,
  PollingOptions,
  ProcessDocumentAsyncRequest
} from '../types/processing';

// 处理状态Hook
export interface UseProcessingStatusResult {
  status: ProcessingStatusResponse | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  startPolling: (options?: PollingOptions) => void;
  stopPolling: () => void;
}

/**
 * 使用文档处理状态的Hook
 * @param documentId 文档ID
 * @param autoRefresh 是否自动刷新
 */
export function useProcessingStatus(
  documentId: string | null,
  autoRefresh = false
): UseProcessingStatusResult {
  const [status, setStatus] = useState<ProcessingStatusResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pollingRef = useRef<boolean>(false);

  const refresh = useCallback(async () => {
    if (!documentId) return;

    setLoading(true);
    setError(null);

    try {
      const result = await processingService.getProcessingStatus(documentId);
      setStatus(result);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to get processing status';
      setError(errorMessage);
      console.error('Error fetching processing status:', err);
    } finally {
      setLoading(false);
    }
  }, [documentId]);

  const startPolling = useCallback((options?: PollingOptions) => {
    if (!documentId || pollingRef.current) return;

    pollingRef.current = true;
    processingService.startPolling(
      documentId,
      (newStatus) => {
        setStatus(newStatus);
        setError(null);
      },
      {
        stopOnComplete: true,
        ...options
      }
    );
  }, [documentId]);

  const stopPolling = useCallback(() => {
    if (!documentId) return;
    
    pollingRef.current = false;
    processingService.stopPolling(documentId);
  }, [documentId]);

  // 初始加载
  useEffect(() => {
    if (documentId) {
      refresh();
      
      if (autoRefresh) {
        startPolling();
      }
    }

    return () => {
      stopPolling();
    };
  }, [documentId, autoRefresh, refresh, startPolling, stopPolling]);

  return {
    status,
    loading,
    error,
    refresh,
    startPolling,
    stopPolling
  };
}

// 任务状态Hook
export interface UseTaskStatusResult {
  task: TaskStatusResponse | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  startPolling: (options?: PollingOptions) => void;
  stopPolling: () => void;
  cancel: () => Promise<void>;
}

/**
 * 使用任务状态的Hook
 * @param taskId 任务ID
 * @param autoRefresh 是否自动刷新
 */
export function useTaskStatus(
  taskId: string | null,
  autoRefresh = false
): UseTaskStatusResult {
  const [task, setTask] = useState<TaskStatusResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pollingRef = useRef<boolean>(false);

  const refresh = useCallback(async () => {
    if (!taskId) return;

    setLoading(true);
    setError(null);

    try {
      const result = await processingService.getTaskStatus(taskId);
      setTask(result);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to get task status';
      setError(errorMessage);
      console.error('Error fetching task status:', err);
    } finally {
      setLoading(false);
    }
  }, [taskId]);

  const startPolling = useCallback((options?: PollingOptions) => {
    if (!taskId || pollingRef.current) return;

    pollingRef.current = true;
    processingService.startTaskPolling(
      taskId,
      (newTask) => {
        setTask(newTask);
        setError(null);
      },
      {
        stopOnComplete: true,
        ...options
      }
    );
  }, [taskId]);

  const stopPolling = useCallback(() => {
    if (!taskId) return;
    
    pollingRef.current = false;
    processingService.stopPolling(taskId);
  }, [taskId]);

  const cancel = useCallback(async () => {
    if (!taskId) return;

    try {
      await processingService.cancelTask(taskId);
      message.success('任务已取消');
      await refresh();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to cancel task';
      message.error(`取消任务失败: ${errorMessage}`);
      console.error('Error canceling task:', err);
    }
  }, [taskId, refresh]);

  // 初始加载
  useEffect(() => {
    if (taskId) {
      refresh();
      
      if (autoRefresh) {
        startPolling();
      }
    }

    return () => {
      stopPolling();
    };
  }, [taskId, autoRefresh, refresh, startPolling, stopPolling]);

  return {
    task,
    loading,
    error,
    refresh,
    startPolling,
    stopPolling,
    cancel
  };
}

// 处理操作Hook
export interface UseProcessingActionsResult {
  processDocument: (documentId: string) => Promise<void>;
  processDocumentAsync: (documentId: string, options?: ProcessDocumentAsyncRequest) => Promise<string>;
  reprocessDocument: (documentId: string) => Promise<void>;
  batchProcess: (documentIds: string[], async?: boolean, priority?: number) => Promise<void>;
  loading: boolean;
  error: string | null;
}

/**
 * 使用处理操作的Hook
 */
export function useProcessingActions(): UseProcessingActionsResult {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const processDocument = useCallback(async (documentId: string) => {
    setLoading(true);
    setError(null);

    try {
      await processingService.processDocument(documentId);
      message.success('文档处理已开始');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to process document';
      setError(errorMessage);
      message.error(`处理失败: ${errorMessage}`);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const processDocumentAsync = useCallback(async (
    documentId: string, 
    options?: ProcessDocumentAsyncRequest
  ): Promise<string> => {
    setLoading(true);
    setError(null);

    try {
      const result = await processingService.processDocumentAsync(documentId, options);
      message.success('异步处理任务已创建');
      return result.task_id;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create processing task';
      setError(errorMessage);
      message.error(`创建任务失败: ${errorMessage}`);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const reprocessDocument = useCallback(async (documentId: string) => {
    setLoading(true);
    setError(null);

    try {
      await processingService.reprocessDocument(documentId);
      message.success('文档重新处理已开始');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to reprocess document';
      setError(errorMessage);
      message.error(`重新处理失败: ${errorMessage}`);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const batchProcess = useCallback(async (
    documentIds: string[], 
    async = false, 
    priority = 1
  ) => {
    setLoading(true);
    setError(null);

    try {
      const result = await processingService.batchProcessDocuments({
        document_ids: documentIds,
        async,
        priority
      });

      if (result.success) {
        message.success(`批量处理已开始，共处理 ${result.processed_count} 个文档`);
      } else {
        throw new Error(result.error_message || 'Batch processing failed');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to batch process documents';
      setError(errorMessage);
      message.error(`批量处理失败: ${errorMessage}`);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    processDocument,
    processDocumentAsync,
    reprocessDocument,
    batchProcess,
    loading,
    error
  };
}

// 状态工具函数
export const ProcessingUtils = {
  /**
   * 检查状态是否为处理中
   */
  isProcessing: (status: ProcessingStatusType): boolean => {
    return processingService.isProcessing(status);
  },

  /**
   * 检查状态是否为完成
   */
  isCompleted: (status: ProcessingStatusType): boolean => {
    return processingService.isCompleted(status);
  },

  /**
   * 检查状态是否为失败
   */
  isFailed: (status: ProcessingStatusType): boolean => {
    return processingService.isFailed(status);
  },

  /**
   * 获取状态显示文本
   */
  getStatusText: (status: ProcessingStatusType): string => {
    return processingService.getStatusText(status);
  },

  /**
   * 获取状态颜色
   */
  getStatusColor: (status: ProcessingStatusType) => {
    return processingService.getStatusColor(status);
  },

  /**
   * 格式化文件大小
   */
  formatFileSize: (bytes: number): string => {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  },

  /**
   * 格式化处理时间
   */
  formatDuration: (seconds: number): string => {
    if (seconds < 60) {
      return `${Math.round(seconds)}秒`;
    } else if (seconds < 3600) {
      const minutes = Math.floor(seconds / 60);
      const remainingSeconds = Math.round(seconds % 60);
      return `${minutes}分${remainingSeconds}秒`;
    } else {
      const hours = Math.floor(seconds / 3600);
      const minutes = Math.floor((seconds % 3600) / 60);
      return `${hours}小时${minutes}分`;
    }
  }
};