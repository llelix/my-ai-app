import axios from 'axios';
import { API_BASE_URL } from '../types';

interface Document {
  id: number;
  name: string;
  original_name: string;
  file_path: string;
  file_size: number;
  file_hash: string;
  mime_type: string;
  extension: string;
  description: string;
  status: string;
  created_at: string;
  updated_at: string;
}

interface UploadSession {
  id: string;
  file_name: string;
  file_size: number;
  file_hash: string;
  chunk_size: number;
  total_chunks: number;
  uploaded_size: number;
  temp_dir: string;
  status: string;
  expires_at: string;
  created_at: string;
  updated_at: string;
}

// 计算文件SHA256哈希
async function calculateFileHash(file: File): Promise<string> {
  const buffer = await file.arrayBuffer();
  const hashBuffer = await crypto.subtle.digest('SHA-256', buffer);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}

export const documentService = {
  // 检查文件是否存在（秒传）
  async checkFile(hash: string, size: number): Promise<{ exists: boolean; document?: Document }> {
    const response = await axios.get(`${API_BASE_URL}/documents/check`, {
      params: { hash, size }
    });
    return response.data;
  },

  // 初始化分片上传
  async initUpload(fileName: string, fileSize: number, fileHash: string): Promise<UploadSession> {
    const response = await axios.post(`${API_BASE_URL}/documents/init`, {
      file_name: fileName,
      file_size: fileSize,
      file_hash: fileHash
    });
    return response.data.data;
  },

  // 上传分片
  async uploadChunk(sessionId: string, chunkIndex: number, chunkData: ArrayBuffer): Promise<void> {
    await axios.post(`${API_BASE_URL}/documents/chunk/${sessionId}/${chunkIndex}`, chunkData, {
      headers: { 'Content-Type': 'application/octet-stream' }
    });
  },

  // 完成上传
  async completeUpload(sessionId: string): Promise<Document> {
    const response = await axios.post(`${API_BASE_URL}/documents/complete/${sessionId}`);
    return response.data.data;
  },

  // 获取上传进度
  async getUploadProgress(sessionId: string): Promise<UploadSession> {
    const response = await axios.get(`${API_BASE_URL}/documents/progress/${sessionId}`);
    return response.data.data;
  },

  // 分片上传文件
  async uploadWithResume(
    file: File, 
    onProgress?: (progress: number) => void
  ): Promise<Document> {
    const fileHash = await calculateFileHash(file);
    
    // 检查是否可以秒传
    const checkResult = await this.checkFile(fileHash, file.size);
    if (checkResult.exists && checkResult.document) {
      onProgress?.(100);
      return checkResult.document;
    }

    // 初始化上传会话
    const session = await this.initUpload(file.name, file.size, fileHash);
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
    return await this.completeUpload(session.id);
  },

  // 传统上传方法
  async upload(file: File): Promise<Document> {
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await axios.post(`${API_BASE_URL}/documents/upload`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    return response.data.data;
  },

  async list(): Promise<Document[]> {
    const response = await axios.get(`${API_BASE_URL}/documents`);
    return response.data.data;
  },

  async get(id: number): Promise<Document> {
    const response = await axios.get(`${API_BASE_URL}/documents/${id}`);
    return response.data.data;
  },

  async delete(id: number): Promise<void> {
    await axios.delete(`${API_BASE_URL}/documents/${id}`);
  },

  async updateDescription(id: number, description: string): Promise<void> {
    await axios.put(`${API_BASE_URL}/documents/${id}/description`, { description });
  },

  getDownloadUrl(id: number): string {
    return `${API_BASE_URL}/documents/${id}/download`;
  }
};

export type { Document, UploadSession };
