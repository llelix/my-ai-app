// 导出所有服务
export { default as apiService, healthCheck } from './api';
export { knowledgeService } from './knowledge';
export { aiService } from './ai';
export { documentService } from './documentService';
export { documentProcessingService } from './documentProcessingService';

// 标签服务
import { apiService } from './api';
import type {
  PaginationRequest,
  PaginationResponse
} from '../types';
export class TagService {
  // 获取标签列表
  async getTags(params?: { is_active?: boolean; search?: string }) {
    return apiService.get<any[]>('/tags', { params });
  }

  // 获取单个标签
  async getTag(id: number) {
    return apiService.get<any>(`/tags/${id}`);
  }

  // 创建标签
  async createTag(data: { name: string; color?: string }) {
    return apiService.post<any>('/tags', data);
  }

  // 更新标签
  async updateTag(id: number, data: { name: string; color?: string }) {
    return apiService.put<any>(`/tags/${id}`, data);
  }

  // 删除标签
  async deleteTag(id: number) {
    return apiService.delete(`/tags/${id}`);
  }

  // 获取标签下的知识
  async getTagKnowledges(id: number, params?: PaginationRequest) {
    return apiService.get<PaginationResponse<any>>(`/tags/${id}/knowledges`, { params });
  }

  // 获取热门标签
  async getPopularTags(limit?: number) {
    const params = limit ? { limit } : {};
    return apiService.get<any[]>('/tags/popular', { params });
  }
}

// 统计服务
export class StatsService {
  // 获取概览统计
  async getOverviewStats() {
    return apiService.get<{
      knowledge_count: number;
      tag_count: number;
      query_count: number;
    }>('/stats/overview');
  }

  // 获取知识库统计
  async getKnowledgeStats() {
    return apiService.get<{
      by_tags: any[];
    }>('/stats/knowledge');
  }

  // 获取查询统计
  async getQueryStats() {
    return apiService.get<{
      today_count: number;
      week_count: number;
      total_count: number;
      success_rate: number;
      popular_queries: any[];
    }>('/stats/query');
  }

  // 获取时间趋势统计
  async getTrendStats(params: { period: 'week' | 'month' | 'year' }) {
    return apiService.get<any[]>('/stats/trends', { params });
  }
}

// 文件服务
export class FileService {
  // 上传文件
  async uploadFile(file: File, onProgress?: (progress: number) => void) {
    const formData = new FormData();
    formData.append('file', file);

    return new Promise((resolve, reject) => {
      const config = {
        onUploadProgress: (progressEvent: any) => {
          if (onProgress && progressEvent.total) {
            const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
            onProgress(progress);
          }
        },
      };

      apiService.upload<{
        filename: string;
        size: number;
        mime_type: string;
        url: string;
      }>('/files/upload', formData, config)
        .then(resolve)
        .catch(reject);
    });
  }

  // 获取文件信息
  async getFileInfo(filename: string) {
    return apiService.get<any>(`/files/info/${filename}`);
  }

  // 删除文件
  async deleteFile(filename: string) {
    return apiService.delete(`/files/${filename}`);
  }
}

// 创建服务实例
export const tagService = new TagService();
export const statsService = new StatsService();
export const fileService = new FileService();