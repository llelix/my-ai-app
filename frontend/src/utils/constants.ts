// API常量
export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  TIMEOUT: 30000,
  RETRY_COUNT: 3,
};

// 路由常量
export const ROUTES = {
  HOME: '/',
  KNOWLEDGE: '/knowledge',
  KNOWLEDGE_DETAIL: '/knowledge/:id',
  KNOWLEDGE_CREATE: '/knowledge/create',
  KNOWLEDGE_EDIT: '/knowledge/:id/edit',
  CATEGORIES: '/categories',
  TAGS: '/tags',
  AI_CHAT: '/ai/chat',
  AI_HISTORY: '/ai/history',
  STATISTICS: '/statistics',
  SETTINGS: '/settings',
};

// 主题常量
export const THEMES = {
  COLORS: [
    '#1890ff', // 蓝色
    '#52c41a', // 绿色
    '#faad14', // 黄色
    '#f5222d', // 红色
    '#722ed1', // 紫色
    '#13c2c2', // 青色
    '#eb2f96', // 粉色
    '#fa8c16', // 橙色
  ],
  FONT_SIZES: {
    SMALL: '12px',
    MEDIUM: '14px',
    LARGE: '16px',
    XLARGE: '18px',
  },
};

// 表格配置
export const TABLE_CONFIG = {
  ROW_SELECTION_SIZE: 'default',
  SCROLL_Y: 'calc(100vh - 280px)',
  PAGE_SIZE_OPTIONS: ['10', '20', '50', '100'],
  DEFAULT_PAGE_SIZE: 10,
};

// 表单验证
export const VALIDATION_RULES = {
  REQUIRED: { required: true, message: '此字段为必填项' },
  EMAIL: { type: 'email', message: '请输入有效的邮箱地址' },
  URL: { type: 'url', message: '请输入有效的URL' },
  MIN_LENGTH: (min: number) => ({ min, message: `最少输入${min}个字符` }),
  MAX_LENGTH: (max: number) => ({ max, message: `最多输入${max}个字符` }),
  PATTERN: (pattern: RegExp, message: string) => ({ pattern, message }),
};

// 日志级别
export const LOG_LEVELS = {
  DEBUG: 'debug',
  INFO: 'info',
  WARN: 'warn',
  ERROR: 'error',
};

// 文件上传配置
export const UPLOAD_CONFIG = {
  MAX_SIZE: 10 * 1024 * 1024, // 10MB
  ACCEPTED_TYPES: [
    'image/jpeg',
    'image/png',
    'image/gif',
    'application/pdf',
    'text/plain',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  ],
  CHUNK_SIZE: 1024 * 1024, // 1MB
};

// 时间格式
export const DATE_FORMATS = {
  DATE: 'YYYY-MM-DD',
  DATETIME: 'YYYY-MM-DD HH:mm:ss',
  TIME: 'HH:mm:ss',
  MONTH: 'YYYY-MM',
  YEAR: 'YYYY',
};

// 本地存储键名
export const STORAGE_KEYS = {
  AUTH_TOKEN: 'auth_token',
  USER_INFO: 'user_info',
  THEME: 'theme',
  LANGUAGE: 'language',
  PREFERENCES: 'preferences',
};

// 分页默认值
export const PAGINATION = {
  DEFAULT_PAGE: 1,
  DEFAULT_PAGE_SIZE: 10,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
};

// AI配置
export const AI_CONFIG = {
  DEFAULT_MODEL: 'gpt-3.5-turbo',
  DEFAULT_TEMPERATURE: 0.7,
  DEFAULT_MAX_TOKENS: 2000,
  QUERY_TIMEOUT: 60000, // 1分钟
  HISTORY_LIMIT: 100,
};

// 缓存配置
export const CACHE_CONFIG = {
  TTL: 5 * 60 * 1000, // 5分钟
  MAX_SIZE: 100, // 最大缓存条目数
};

// 错误码
export const ERROR_CODES = {
  NETWORK_ERROR: 'NETWORK_ERROR',
  TIMEOUT_ERROR: 'TIMEOUT_ERROR',
  AUTH_ERROR: 'AUTH_ERROR',
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  SERVER_ERROR: 'SERVER_ERROR',
  NOT_FOUND: 'NOT_FOUND',
  PERMISSION_DENIED: 'PERMISSION_DENIED',
};

// 成功消息
export const SUCCESS_MESSAGES = {
  CREATE_SUCCESS: '创建成功',
  UPDATE_SUCCESS: '更新成功',
  DELETE_SUCCESS: '删除成功',
  SAVE_SUCCESS: '保存成功',
  UPLOAD_SUCCESS: '上传成功',
  COPY_SUCCESS: '复制成功',
};

// 默认配置
export const DEFAULT_CONFIG = {
  THEME: 'light',
  LANGUAGE: 'zh-CN',
  COLOR: '#1890ff',
  COMPACT_MODE: false,
};