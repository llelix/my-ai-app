// API响应基础类型
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
}

// 分页请求类型
export interface PaginationRequest {
  page: number;
  page_size: number;
  search?: string;
  sort?: string;
  order?: 'asc' | 'desc';
}

// 分页响应类型
export interface PaginationResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// 知识条目相关类型
export interface Knowledge {
  id: number;
  title: string;
  content: string;
  summary: string;
  category_id: number;
  tags: Tag[];
  metadata: Metadata;
  is_published: boolean;
  view_count: number;
  created_at: string;
  updated_at: string;
  category?: Category;
}

export interface CreateKnowledgeRequest {
  title: string;
  content: string;
  summary?: string;
  category_id: number;
  tags: string[];
  metadata: Metadata;
  is_published: boolean;
}

export interface UpdateKnowledgeRequest {
  title?: string;
  content?: string;
  summary?: string;
  category_id?: number;
  tags?: string[];
  metadata?: Metadata;
  is_published?: boolean;
}

// 分类相关类型
export interface Category {
  id: number;
  name: string;
  description: string;
  color: string;
  icon: string;
  parent_id?: number;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  parent?: Category;
  children?: Category[];
  knowledges?: Knowledge[];
}

export interface CreateCategoryRequest {
  name: string;
  description?: string;
  color?: string;
  icon?: string;
  parent_id?: number;
  sort_order?: number;
  is_active?: boolean;
}

// 标签相关类型
export interface Tag {
  id: number;
  name: string;
  color: string;
  usage_count: number;
  created_at: string;
  updated_at: string;
  knowledges?: Knowledge[];
}

export interface CreateTagRequest {
  name: string;
  color?: string;
}

// 元数据类型
export interface Metadata {
  author?: string;
  source?: string;
  language?: string;
  difficulty?: 'easy' | 'medium' | 'hard';
  keywords?: string;
  word_count?: number;
}

// AI查询相关类型
export interface AIQueryRequest {
  query: string;
  model?: string;
  temperature?: number;
  max_tokens?: number;
  context?: string[];
}

export interface AIQueryResponse {
  response: string;
  model: string;
  tokens: number;
  duration: number;
  knowledge_ids?: number[];
  relevant_docs?: string[];
  related_knowledges?: Knowledge[];
}

// 查询历史类型
export interface QueryHistory {
  id: number;
  query: string;
  response: string;
  knowledge_id?: number;
  model: string;
  tokens: number;
  duration: number;
  is_success: boolean;
  error_message?: string;
  created_at: string;
  updated_at: string;
  knowledge?: Knowledge;
}

// 查询统计类型
export interface QueryStats {
  today_count: number;
  week_count: number;
  total_count: number;
  success_count: number;
  success_rate: number;
  avg_duration: number;
  by_models: Array<{
    model: string;
    count: number;
  }>;
  popular_queries: Array<{
    query: string;
    count: number;
  }>;
}

// 概览统计类型
export interface OverviewStats {
  knowledge_count: number;
  category_count: number;
  tag_count: number;
  query_count: number;
}

// 应用状态类型
export interface AppState {
  loading: boolean;
  user: {
    name: string;
    avatar?: string;
  } | null;
}

// 路由相关类型
export interface RouteConfig {
  path: string;
  component: React.ComponentType;
  title: string;
  icon?: string;
  children?: RouteConfig[];
}

// 主题相关类型
export interface ThemeConfig {
  primaryColor: string;
  darkMode: boolean;
  compactMode: boolean;
}

// 表单验证规则类型
export interface ValidationRule {
  required?: boolean;
  message?: string;
  min?: number;
  max?: number;
  pattern?: RegExp;
  validator?: (value: any) => boolean | Promise<boolean>;
}

// 错误类型
export interface AppError {
  code: number;
  message: string;
  details?: any;
}

// 文件上传类型
export interface UploadFile {
  filename: string;
  size: number;
  mime_type: string;
  url: string;
}

// 反馈类型
export interface FeedbackRequest {
  query_id: number;
  rating: number;
  comment?: string;
  is_helpful: boolean;
}

// 导出的常量
export const API_BASE_URL = '/api/v1';

export const DEFAULT_PAGINATION: PaginationRequest = {
  page: 1,
  page_size: 10,
  order: 'desc',
};

export const AI_MODELS = [
  'gpt-3.5-turbo',
  'gpt-4',
  'qwen-plus',
  'qwen-max',
  'moonshot-v1-8k',
  'moonshot-v1-32k',
  'glm-3-turbo',
  'glm-4',
];

export const DIFFICULTY_OPTIONS = [
  { label: '简单', value: 'easy' },
  { label: '中等', value: 'medium' },
  { label: '困难', value: 'hard' },
];