import axios, { type InternalAxiosRequestConfig } from 'axios';
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import type { ApiResponse } from '../types';
import { API_BASE_URL } from '../types';

class ApiService {
  private instance: AxiosInstance;

  constructor() {
    this.instance = axios.create({
      baseURL: API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // 请求拦截器
    this.instance.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        // 可以在这里添加token等认证信息
        const token = localStorage.getItem('auth_token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }

        // 添加请求ID用于追踪
        config.headers['X-Request-ID'] = this.generateRequestId();

        console.log(`🚀 API Request: ${config.method?.toUpperCase()} ${config.url}`);
        return config;
      },
      (error) => {
        console.error('❌ Request Error:', error);
        return Promise.reject(error);
      }
    );

    // 响应拦截器
    this.instance.interceptors.response.use(
      (response: AxiosResponse<ApiResponse>) => {
        console.log(`✅ API Response: ${response.status} ${response.config.url}`);
        return response;
      },
      (error) => {
        console.error('❌ Response Error:', error);

        // 处理不同类型的错误
        if (error.response) {
          // 服务器响应错误
          const status = error.response.status;
          const message = error.response.data?.message || '请求失败';

          switch (status) {
            case 401:
              // 未授权，可能需要重新登录
              localStorage.removeItem('auth_token');
              window.location.href = '/login';
              break;
            case 403:
              // 权限不足
              console.error('权限不足:', message);
              break;
            case 404:
              console.error('资源不存在:', message);
              break;
            case 422:
              console.error('数据验证失败:', message);
              break;
            case 500:
              console.error('服务器内部错误:', message);
              break;
            default:
              console.error(`请求失败 (${status}):`, message);
          }
        } else if (error.request) {
          // 网络错误
          console.error('网络错误，请检查网络连接:', error.message);
        } else {
          // 其他错误
          console.error('未知错误:', error.message);
        }

        return Promise.reject(error);
      }
    );
  }

  private generateRequestId(): string {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }

  // GET请求
  async get<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.instance.get<ApiResponse<T>>(url, config);
    return response.data;
  }

  // POST请求
  async post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.instance.post<ApiResponse<T>>(url, data, config);
    return response.data;
  }

  // PUT请求
  async put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.instance.put<ApiResponse<T>>(url, data, config);
    return response.data;
  }

  // DELETE请求
  async delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.instance.delete<ApiResponse<T>>(url, config);
    return response.data;
  }

  // PATCH请求
  async patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.instance.patch<ApiResponse<T>>(url, data, config);
    return response.data;
  }

  // 文件上传
  async upload<T = any>(url: string, formData: FormData, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const uploadConfig: AxiosRequestConfig = {
      ...config,
      headers: {
        'Content-Type': 'multipart/form-data',
        ...config?.headers,
      },
    };

    const response = await this.instance.post<ApiResponse<T>>(url, formData, uploadConfig);
    return response.data;
  }

  // 下载文件
  async download(url: string, filename?: string, config?: AxiosRequestConfig): Promise<void> {
    const response = await this.instance.get(url, {
      ...config,
      responseType: 'blob',
    });

    const blob = new Blob([response.data]);
    const downloadUrl = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = filename || 'download';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(downloadUrl);
  }

  // 获取原始axios实例（用于特殊需求）
  getInstance(): AxiosInstance {
    return this.instance;
  }
}

// 创建API服务实例
export const apiService = new ApiService();

// 健康检查
export const healthCheck = () => {
  return apiService.get('/health');
};

// 导出默认实例
export default apiService;