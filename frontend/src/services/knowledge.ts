import { apiService } from './api';
import type {
  Knowledge,
  CreateKnowledgeRequest,
  UpdateKnowledgeRequest,
  PaginationRequest,
  PaginationResponse
} from '../types';

export class KnowledgeService {
  // 获取知识列表
  async getKnowledges(params?: PaginationRequest & {
    category_id?: number;
    tag_id?: number;
    include_unpublished?: boolean;
  }) {
    return apiService.get<PaginationResponse<Knowledge>>('/knowledge', { params });
  }

  // 获取单个知识
  async getKnowledge(id: number) {
    return apiService.get<Knowledge>(`/knowledge/${id}`);
  }

  // 创建知识
  async createKnowledge(data: CreateKnowledgeRequest) {
    return apiService.post<Knowledge>('/knowledge', data);
  }

  // 更新知识
  async updateKnowledge(id: number, data: UpdateKnowledgeRequest) {
    return apiService.put<Knowledge>(`/knowledge/${id}`, data);
  }

  // 删除知识
  async deleteKnowledge(id: number) {
    return apiService.delete(`/knowledge/${id}`);
  }

  // 搜索知识
  async searchKnowledges(params: PaginationRequest & { q: string }) {
    return apiService.get<PaginationResponse<Knowledge>>('/knowledge/search', { params });
  }

  // 获取相关知识
  async getRelatedKnowledges(id: number, limit?: number) {
    const params = limit ? { limit } : {};
    return apiService.get<Knowledge[]>(`/knowledge/${id}/related`, { params });
  }

  // 增加查看次数
  async incrementViewCount(id: number) {
    return apiService.post<{ view_count: number }>(`/knowledge/${id}/view`);
  }

  // 批量操作
  async batchDelete(ids: number[]) {
    return apiService.post('/knowledge/batch-delete', { ids });
  }

  async batchUpdate(ids: number[], data: Partial<UpdateKnowledgeRequest>) {
    return apiService.post('/knowledge/batch-update', { ids, data });
  }
}

export const knowledgeService = new KnowledgeService();