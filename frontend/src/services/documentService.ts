import { apiService } from './api';
import type {
  Document,
  UploadSession,
  PaginationRequest,
  PaginationResponse
} from '../types';

// 计算文件SHA256哈希
async function calculateFileHash(file: File): Promise<string> {
  const buffer = await file.arrayBuffer();
  const hashBuffer = await crypto.subtle.digest('SHA-256', buffer);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}

export class DocumentService {
  // 检查文件是否存在（秒传）
  async checkFile(hash: string, size: number) {
    return apiService.get<{ exists: boolean; document?: Document }>('/documents/check', {
      params: { hash, size }
    });
  }

  // 初始化分片上传
  async initUpload(fileName: string, fileSize: number, fileHash: string) {
    return apiService.post<UploadSession>('/documents/init', {
      file_name: fileName,
      file_size: fileSize,
      file_hash: fileHash
    });
  }

  // 上传分片
  async uploadChunk(sessionId: string, chunkIndex: number, chunkData: ArrayBuffer) {
    return apiService.getInstance().post(`/documents/chunk/${sessionId}/${chunkIndex}`, chunkData, {
      headers: { 'Content-Type': 'application/octet-stream' }
    });
  }

  // 完成上传
  async completeUpload(sessionId: string) {
    return apiService.post<Document>(`/documents/complete/${sessionId}`);
  }

  // 获取上传进度
  async getUploadProgress(sessionId: string) {
    return apiService.get<UploadSession>(`/documents/progress/${sessionId}`);
  }

  // 分片上传文件
  async uploadWithResume(
    file: File, 
    onProgress?: (progress: number) => void
  ) {
    const fileHash = await calculateFileHash(file);
    
    // 检查是否可以秒传
    const checkResult = await this.checkFile(fileHash, file.size);
    if (checkResult.data?.exists && checkResult.data?.document) {
      onProgress?.(100);
      return checkResult.data.document;
    }

    // 初始化上传会话
    const sessionResult = await this.initUpload(file.name, file.size, fileHash);
    const session = sessionResult.data;
    if (!session) {
      throw new Error('Failed to initialize upload session');
    }
    
    const chunkSize = session.chunk_size;

    // 分片上传
    for (let i = 0; i < session.total_chunks; i++) {
      const start = i * chunkSize;
      const end = Math.min(start + chunkSize, file.size);
      const chunk = file.slice(start, end);
      const chunkBuffer = await chunk.arrayBuffer();
      
      await this.uploadChunk(session.id, i, chunkBuffer);
      
      const progress = ((i + 1) / session.total_chunks) * 100;
      onProgress?.(progress);
    }

    // 完成上传
    const result = await this.completeUpload(session.id);
    if (!result.data) {
      throw new Error('Failed to complete upload');
    }
    return result.data;
  }

  // 传统上传方法
  async upload(file: File, description?: string) {
    const formData = new FormData();
    formData.append('file', file);
    if (description) {
      formData.append('description', description);
    }
    
    return apiService.upload<Document>('/documents/upload', formData);
  }

  // 获取文档列表
  async getDocuments(params?: PaginationRequest & {
    status?: string;
    mime_type?: string;
  }) {
    return apiService.get<PaginationResponse<Document>>('/documents', { params });
  }

  // 获取单个文档
  async getDocument(id: number) {
    return apiService.get<Document>(`/documents/${id}`);
  }

  // 删除文档
  async deleteDocument(id: number) {
    return apiService.delete(`/documents/${id}`);
  }

  // 更新文档描述
  async updateDescription(id: number, description: string) {
    return apiService.put(`/documents/${id}/description`, { description });
  }

  // 获取下载链接
  getDownloadUrl(id: number): string {
    return `/api/v1/documents/${id}/download`;
  }

  // 下载文档
  async downloadDocument(id: number, filename?: string) {
    return apiService.download(`/documents/${id}/download`, filename);
  }

  // 批量删除文档
  async batchDelete(ids: number[]) {
    return apiService.post('/documents/batch-delete', { ids });
  }

  // 获取文档统计
  async getDocumentStats() {
    return apiService.get<{
      total_count: number;
      total_size: number;
      by_type: Array<{ mime_type: string; count: number; size: number }>;
      by_status: Array<{ status: string; count: number }>;
    }>('/documents/stats');
  }
}

// 创建服务实例
export const documentService = new DocumentService();
