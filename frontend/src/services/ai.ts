import { apiService } from './api';
import type {
  AIQueryRequest,
  AIQueryResponse,
  QueryHistory,
  QueryStats,
  PaginationResponse,
  PaginationRequest,
  FeedbackRequest
} from '../types';

export class AIService {
  // AI查询
  async query(data: AIQueryRequest) {
    return apiService.post<AIQueryResponse>('/ai/query', data);
  }

  // 获取查询历史
  async getQueryHistory(params?: PaginationRequest & { model?: string }) {
    return apiService.get<PaginationResponse<QueryHistory>>('/ai/history', { params });
  }

  // 删除查询历史
  async deleteQueryHistory(id: number) {
    return apiService.delete(`/ai/history/${id}`);
  }

  // 获取查询统计
  async getQueryStats() {
    return apiService.get<QueryStats>('/ai/history/stats');
  }

  // 提交反馈
  async submitFeedback(data: FeedbackRequest) {
    return apiService.post('/ai/feedback', data);
  }

  // 获取支持的模型列表
  async getModels() {
    return apiService.get<{ models: string[] }>('/ai/models');
  }

  // 流式查询（如果后端支持的话）
  async streamQuery(data: AIQueryRequest, onChunk: (chunk: string) => void) {
    // 这里可以实现流式响应处理
    return new Promise<AIQueryResponse>((resolve, reject) => {
      const eventSource = new EventSource(
        `/api/v1/ai/stream?query=${encodeURIComponent(data.query)}&model=${data.model || 'gpt-3.5-turbo'}`
      );

      let fullResponse = '';

      eventSource.onmessage = (event) => {
        const chunk = event.data;
        if (chunk === '[DONE]') {
          eventSource.close();
          resolve({
            response: fullResponse,
            model: data.model || 'gpt-3.5-turbo',
            tokens: 0, // 这里需要后端返回
            duration: 0,
          } as AIQueryResponse);
        } else {
          try {
            const data = JSON.parse(chunk);
            fullResponse += data.content || '';
            onChunk(data.content || '');
          } catch (error) {
            console.error('解析流数据失败:', error);
          }
        }
      };

      eventSource.onerror = (error) => {
        eventSource.close();
        reject(error);
      };
    });
  }

  // 智能搜索（结合AI的增强搜索）
  async smartSearch(query: string, filters?: {
    category_id?: number;
    tags?: string[];
    date_range?: [string, string];
  }) {
    return apiService.post<{
      results: any[];
      suggestions: string[];
      related_topics: string[];
    }>('/ai/smart-search', {
      query,
      filters,
    });
  }

  // 生成知识摘要
  async generateSummary(content: string, maxLength = 200) {
    return apiService.post<{ summary: string }>('/ai/summarize', {
      content,
      max_length: maxLength,
    });
  }

  // 自动生成标签
  async generateTags(title: string, content: string) {
    return apiService.post<{ tags: string[] }>('/ai/generate-tags', {
      title,
      content,
    });
  }

  // 知识推荐
  async getRecommendations(knowledgeId: number, limit = 5) {
    return apiService.get<{
      knowledges: any[];
      reason: string;
    }>(`/ai/recommendations/${knowledgeId}`, {
      params: { limit },
    });
  }
}

export const aiService = new AIService();