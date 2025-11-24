import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import type { AppState } from '../types';

// 全局状态store
interface AppStore extends AppState {
  // Actions
  setLoading: (loading: boolean) => void;
  setUser: (user: AppState['user']) => void;
  logout: () => void;
  reset: () => void;
}

export const useAppStore = create<AppStore>()(
  devtools(
    persist(
      (set) => ({
        // Initial state
        loading: false,
        user: null,

        // Actions
        setLoading: (loading) => set({ loading }, false, 'setLoading'),

        setUser: (user) => set({ user }, false, 'setUser'),

        logout: () => {
          localStorage.removeItem('auth_token');
          set({ user: null }, false, 'logout');
        },

        reset: () => {
          set({ loading: false, user: null }, false, 'reset');
        },
      }),
      {
        name: 'app-store',
        partialize: (state) => ({
          user: state.user,
        }),
      }
    ),
    {
      name: 'app-store',
    }
  )
);

// 知识库相关状态
interface KnowledgeState {
  knowledges: any[];
  currentKnowledge: any | null;
  categories: any[];
  tags: any[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  filters: {
    search: string;
    category_id: number | null;
    tag_id: number | null;
  };
  loading: boolean;

  // Actions
  setKnowledges: (knowledges: any[]) => void;
  setCurrentKnowledge: (knowledge: any | null) => void;
  setCategories: (categories: any[]) => void;
  setTags: (tags: any[]) => void;
  setPagination: (pagination: Partial<KnowledgeState['pagination']>) => void;
  setFilters: (filters: Partial<KnowledgeState['filters']>) => void;
  setLoading: (loading: boolean) => void;
  reset: () => void;
}

export const useKnowledgeStore = create<KnowledgeState>()(
  devtools(
    (set) => ({
      // Initial state
      knowledges: [],
      currentKnowledge: null,
      categories: [],
      tags: [],
      pagination: {
        page: 1,
        page_size: 10,
        total: 0,
        total_pages: 0,
      },
      filters: {
        search: '',
        category_id: null,
        tag_id: null,
      },
      loading: false,

      // Actions
      setKnowledges: (knowledges) => set({ knowledges }, false, 'setKnowledges'),

      setCurrentKnowledge: (currentKnowledge) =>
        set({ currentKnowledge }, false, 'setCurrentKnowledge'),

      setCategories: (categories) => set({ categories }, false, 'setCategories'),

      setTags: (tags) => set({ tags }, false, 'setTags'),

      setPagination: (pagination) =>
        set((state) => ({ pagination: { ...state.pagination, ...pagination } }), false, 'setPagination'),

      setFilters: (filters) =>
        set((state) => ({ filters: { ...state.filters, ...filters } }), false, 'setFilters'),

      setLoading: (loading) => set({ loading }, false, 'setLoading'),

      reset: () =>
        set(
          {
            knowledges: [],
            currentKnowledge: null,
            pagination: {
              page: 1,
              page_size: 10,
              total: 0,
              total_pages: 0,
            },
            filters: {
              search: '',
              category_id: null,
              tag_id: null,
            },
            loading: false,
          },
          false,
          'reset'
        ),
    }),
    {
      name: 'knowledge-store',
    }
  )
);

// AI查询相关状态
interface AIState {
  // 查询相关
  queryHistory: any[];
  currentQuery: string;
  currentResponse: string;
  isQuerying: boolean;

  // 配置相关
  selectedModel: string;
  temperature: number;
  maxTokens: number;

  // 统计相关
  queryStats: any;

  // Actions
  setQueryHistory: (history: any[]) => void;
  setCurrentQuery: (query: string) => void;
  setCurrentResponse: (response: string) => void;
  setIsQuerying: (isQuerying: boolean) => void;
  setSelectedModel: (model: string) => void;
  setTemperature: (temperature: number) => void;
  setMaxTokens: (tokens: number) => void;
  setQueryStats: (stats: any) => void;
  addToHistory: (item: any) => void;
  reset: () => void;
}

export const useAIStore = create<AIState>()(
  devtools(
    persist(
      (set) => ({
        // Initial state
        queryHistory: [],
        currentQuery: '',
        currentResponse: '',
        isQuerying: false,
        selectedModel: 'gpt-3.5-turbo',
        temperature: 0.7,
        maxTokens: 2000,
        queryStats: null,

        // Actions
        setQueryHistory: (queryHistory) => set({ queryHistory }, false, 'setQueryHistory'),

        setCurrentQuery: (currentQuery) => set({ currentQuery }, false, 'setCurrentQuery'),

        setCurrentResponse: (currentResponse) => set({ currentResponse }, false, 'setCurrentResponse'),

        setIsQuerying: (isQuerying) => set({ isQuerying }, false, 'setIsQuerying'),

        setSelectedModel: (selectedModel) => set({ selectedModel }, false, 'setSelectedModel'),

        setTemperature: (temperature) => set({ temperature }, false, 'setTemperature'),

        setMaxTokens: (maxTokens) => set({ maxTokens }, false, 'setMaxTokens'),

        setQueryStats: (queryStats) => set({ queryStats }, false, 'setQueryStats'),

        addToHistory: (item) =>
          set(
            (state) => ({
              queryHistory: [item, ...state.queryHistory.slice(0, 99)], // 保留最近100条
            }),
            false,
            'addToHistory'
          ),

        reset: () =>
          set(
            {
              currentQuery: '',
              currentResponse: '',
              isQuerying: false,
            },
            false,
            'reset'
          ),
      }),
      {
        name: 'ai-store',
        partialize: (state) => ({
          selectedModel: state.selectedModel,
          temperature: state.temperature,
          maxTokens: state.maxTokens,
        }),
      }
    ),
    {
      name: 'ai-store',
    }
  )
);

// 主题相关状态
interface ThemeState {
  theme: 'light' | 'dark';
  primaryColor: string;
  compactMode: boolean;

  // Actions
  setTheme: (theme: 'light' | 'dark') => void;
  setPrimaryColor: (color: string) => void;
  setCompactMode: (compactMode: boolean) => void;
  toggleTheme: () => void;
}

export const useThemeStore = create<ThemeState>()(
  devtools(
    persist(
      (set) => ({
        // Initial state
        theme: 'light',
        primaryColor: '#1890ff',
        compactMode: false,

        // Actions
        setTheme: (theme) => set({ theme }, false, 'setTheme'),

        setPrimaryColor: (primaryColor) => set({ primaryColor }, false, 'setPrimaryColor'),

        setCompactMode: (compactMode) => set({ compactMode }, false, 'setCompactMode'),

        toggleTheme: () =>
          set(
            (state) => ({ theme: state.theme === 'light' ? 'dark' : 'light' }),
            false,
            'toggleTheme'
          ),
      }),
      {
        name: 'theme-store',
      }
    ),
    {
      name: 'theme-store',
    }
  )
);