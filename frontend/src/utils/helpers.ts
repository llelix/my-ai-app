import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import relativeTime from 'dayjs/plugin/relativeTime';
import { DATE_FORMATS } from './constants';

// 设置dayjs
dayjs.locale('zh-cn');
dayjs.extend(relativeTime);

// 时间相关工具函数
export const formatDate = (date: string | Date, format = DATE_FORMATS.DATETIME) => {
  return dayjs(date).format(format);
};

export const fromNow = (date: string | Date) => {
  return dayjs(date).fromNow();
};

export const isRecent = (date: string | Date, minutes = 5) => {
  return dayjs().diff(dayjs(date), 'minute') <= minutes;
};

// 文字处理工具
export const truncateText = (text: string, maxLength = 100) => {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
};

export const highlightText = (text: string, highlight: string) => {
  if (!highlight) return text;
  const regex = new RegExp(`(${highlight})`, 'gi');
  return text.replace(regex, '<mark>$1</mark>');
};

export const extractKeywords = (text: string, maxLength = 5) => {
  // 简单的关键词提取，可以用更复杂的算法替换
  const words = text.toLowerCase().match(/[\u4e00-\u9fa5]+|[a-z]+/g) || [];
  const wordCount: Record<string, number> = {};

  words.forEach(word => {
    if (word.length >= 2) {
      wordCount[word] = (wordCount[word] || 0) + 1;
    }
  });

  return Object.entries(wordCount)
    .sort(([, a], [, b]) => b - a)
    .slice(0, maxLength)
    .map(([word]) => word);
};

// 文件工具
export const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

export const getFileExtension = (filename: string) => {
  return filename.slice((filename.lastIndexOf('.') - 1 >>> 0) + 2);
};

export const isImageFile = (filename: string) => {
  const imageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'];
  return imageExtensions.includes(getFileExtension(filename).toLowerCase());
};

// 数组工具
export const groupBy = <T extends object, K extends string | number>(array: T[], key: keyof T): Record<K, T[]> => {
  return array.reduce((groups, item) => {
    const groupValue = String(item[key]);
    const group = groupValue as K;
    groups[group] = groups[group] || [];
    groups[group].push(item);
    return groups;
  }, {} as Record<K, T[]>);
};

export const uniqueBy = <T>(array: T[], key: keyof T) => {
  const seen = new Set();
  return array.filter(item => {
    const value = item[key];
    if (seen.has(value)) return false;
    seen.add(value);
    return true;
  });
};

export const sortBy = <T>(array: T[], key: keyof T, order: 'asc' | 'desc' = 'asc') => {
  return [...array].sort((a, b) => {
    const aVal = a[key];
    const bVal = b[key];
    if (aVal < bVal) return order === 'asc' ? -1 : 1;
    if (aVal > bVal) return order === 'asc' ? 1 : -1;
    return 0;
  });
};

// 对象工具
export const pick = <T extends object, K extends keyof T>(obj: T, keys: K[]): Pick<T, K> => {
  const result = {} as Pick<T, K>;
  keys.forEach(key => {
    if (key in obj) {
      result[key] = obj[key];
    }
  });
  return result;
};

export const omit = <T, K extends keyof T>(obj: T, keys: K[]): Omit<T, K> => {
  const result = { ...obj };
  keys.forEach(key => {
    delete result[key];
  });
  return result;
};

export const deepClone = <T>(obj: T): T => {
  if (obj === null || typeof obj !== 'object') return obj;
  if (obj instanceof Date) return new Date(obj.getTime()) as unknown as T;
  if (obj instanceof Array) return obj.map(item => deepClone(item)) as unknown as T;
  if (typeof obj === 'object') {
    const cloned = {} as T;
    Object.keys(obj).forEach(key => {
      (cloned as any)[key] = deepClone((obj as any)[key]);
    });
    return cloned;
  }
  return obj;
};

// 验证工具
export const validateEmail = (email: string) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

export const validateURL = (url: string) => {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
};

export const validatePhone = (phone: string) => {
  const phoneRegex = /^1[3-9]\d{9}$/;
  return phoneRegex.test(phone);
};

// 本地存储工具
export const storage = {
  get: <T>(key: string, defaultValue?: T): T | null => {
    try {
      const item = localStorage.getItem(key);
      return item ? JSON.parse(item) : defaultValue || null;
    } catch {
      return defaultValue || null;
    }
  },

  set: (key: string, value: any) => {
    try {
      localStorage.setItem(key, JSON.stringify(value));
    } catch (error) {
      console.error('存储失败:', error);
    }
  },

  remove: (key: string) => {
    try {
      localStorage.removeItem(key);
    } catch (error) {
      console.error('删除存储失败:', error);
    }
  },

  clear: () => {
    try {
      localStorage.clear();
    } catch (error) {
      console.error('清空存储失败:', error);
    }
  },
};

// 防抖和节流
export const debounce = <T extends (...args: any[]) => void>(
  func: T,
  wait: number
): (() => void) => {
  let timeout: ReturnType<typeof setTimeout>;
  return function executedFunction(...args: Parameters<T>) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
};

export const throttle = <T extends (...args: any[]) => void>(
  func: T,
  limit: number
): (() => void) => {
  let inThrottle: boolean;
  return function executedFunction(...args: Parameters<T>) {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => inThrottle = false, limit);
    }
  };
};

// 复制到剪贴板
export const copyToClipboard = async (text: string): Promise<boolean> => {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    // 降级方案
    const textArea = document.createElement('textarea');
    textArea.value = text;
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    try {
      document.execCommand('copy');
      document.body.removeChild(textArea);
      return true;
    } catch {
      document.body.removeChild(textArea);
      return false;
    }
  }
};

// 生成随机ID
export const generateId = (length = 8) => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
};

// 生成颜色
export const generateColor = () => {
  const colors = [
    '#f50', '#2db7f5', '#87d068', '#108ee9',
    '#f56a00', '#7265e6', '#ffbf00', '#00a2ae'
  ];
  return colors[Math.floor(Math.random() * colors.length)];
};

// 下载文件
export const downloadFile = (url: string, filename?: string) => {
  const link = document.createElement('a');
  link.href = url;
  link.download = filename || 'download';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};

// 数字格式化
export const formatNumber = (num: number) => {
  return new Intl.NumberFormat('zh-CN').format(num);
};

export const formatCurrency = (amount: number, currency = 'CNY') => {
  return new Intl.NumberFormat('zh-CN', {
    style: 'currency',
    currency,
  }).format(amount);
};

// 百分比格式化
export const formatPercentage = (value: number, decimals = 1) => {
  return `${(value * 100).toFixed(decimals)}%`;
};

// 错误处理
export const handleApiError = (error: any): string => {
  if (error?.response?.data?.message) {
    return error.response.data.message;
  }
  if (error?.message) {
    return error.message;
  }
  return '操作失败，请稍后重试';
};

// URL参数处理
export const getQueryParams = (url?: string) => {
  const searchParams = new URLSearchParams(url ? new URL(url).search : window.location.search);
  const params: Record<string, string> = {};
  searchParams.forEach((value, key) => {
    params[key] = value;
  });
  return params;
};

export const buildQueryString = (params: Record<string, any>) => {
  const searchParams = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      searchParams.append(key, String(value));
    }
  });
  return searchParams.toString();
};