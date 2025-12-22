// 预处理相关类型定义

// 处理状态类型
export type ProcessingStatusType = 
  | 'pending' 
  | 'processing' 
  | 'completed' 
  | 'failed' 
  | 'not_started';

// 异步处理请求
export interface ProcessDocumentAsyncRequest {
  priority?: number; // 任务优先级，1-10，数字越大优先级越高
}

// 异步处理响应
export interface ProcessDocumentAsyncResponse {
  task_id: string;
  document_id: string;
  status: ProcessingStatusType;
  priority: number;
  created_at: string;
}

// 处理状态响应
export interface ProcessingStatusResponse {
  document_id: string;
  status: ProcessingStatusType;
  progress: number;
  error_message?: string;
  started_at?: string;
  completed_at?: string;
  processed_size: number;
  total_size: number;
}

// 任务状态响应
export interface TaskStatusResponse {
  task_id: string;
  document_id: string;
  status: ProcessingStatusType;
  priority: number;
  progress: number;
  error_message?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
}

// 批量处理请求
export interface BatchProcessRequest {
  document_ids: string[];
  priority?: number;
  async?: boolean;
}

// 批量处理响应
export interface BatchProcessResponse {
  success: boolean;
  processed_count: number;
  task_ids?: string[];
  failed_ids?: string[];
  error_message?: string;
}

// 文档分块
export interface DocumentChunk {
  id: string;
  document_id: string;
  content: string;
  chunk_index: number;
  start_offset: number;
  end_offset: number;
  metadata: Record<string, any>;
  created_at: string;
  updated_at: string;
}

// 文档分块响应
export interface DocumentChunksResponse {
  document_id: string;
  chunk_count: number;
  chunks: DocumentChunk[];
}

// 支持格式响应
export interface SupportedFormatsResponse {
  formats: string[];
  count: number;
}

// 队列统计响应
export interface QueueStatsResponse {
  pending_tasks: number;
  processing_tasks: number;
  completed_tasks: number;
  failed_tasks: number;
  total_tasks: number;
  average_wait_time: number;
  worker_count: number;
}

// 处理统计响应
export interface ProcessingStatisticsResponse {
  total_documents: number;
  processed_documents: number;
  failed_documents: number;
  processing_rate: number;
  average_process_time: number;
  total_processed_size: number;
  processing_throughput: number;
}

// 轮询选项
export interface PollingOptions {
  interval?: number;
  maxAttempts?: number;
  stopOnComplete?: boolean;
}

// 处理选项
export interface ProcessingOptions {
  priority?: number;
  async?: boolean;
  retryOnFailure?: boolean;
  maxRetries?: number;
}