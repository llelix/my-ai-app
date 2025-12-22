import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntApp, theme as antdTheme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';

import { useThemeStore } from './store';
import MainLayout from './layouts/MainLayout';

// 页面组件（懒加载）
import KnowledgeList from './pages/Knowledge/KnowledgeList';
import KnowledgeDetail from './pages/Knowledge/KnowledgeDetail';
import KnowledgeCreate from './pages/Knowledge/KnowledgeCreate';
import KnowledgeEdit from './pages/Knowledge/KnowledgeEdit';
import TagList from './pages/Tag/TagList';
import AIChat from './pages/AI/AIChat';
import AIHistory from './pages/AI/AIHistory';
import Statistics from './pages/Statistics/Statistics';
import Settings from './pages/Settings/Settings';
import DocumentManagement from './pages/DocumentManagement';
import NotFound from './pages/NotFound/NotFound';

// 设置dayjs中文语言
dayjs.locale('zh-cn');

const App: React.FC = () => {
  const { theme, primaryColor } = useThemeStore();

  // Ant Design主题配置
  const themeConfig = {
    token: {
      colorPrimary: primaryColor,
      borderRadius: 8,
      wireframe: false,
    },
    algorithm: theme === 'dark' ?
      antdTheme.darkAlgorithm :
      antdTheme.defaultAlgorithm,
  };

  return (
    <ConfigProvider
      locale={zhCN}
      theme={themeConfig}
    >
      <AntApp>
        <Router>
          <Routes>
            {/* 主布局路由 */}
            <Route path="/" element={<MainLayout />}>
              {/* 默认重定向到知识库 */}
              <Route index element={<Navigate to="/knowledge" replace />} />

              {/* 知识库相关路由 */}
              <Route path="knowledge" element={<KnowledgeList />} />
              <Route path="knowledge/:id" element={<KnowledgeDetail />} />
              <Route path="knowledge/create" element={<KnowledgeCreate />} />
              <Route path="knowledge/:id/edit" element={<KnowledgeEdit />} />

              {/* 标签管理 */}
              <Route path="tags" element={<TagList />} />

              {/* AI对话 */}
              <Route path="ai/chat" element={<AIChat />} />
              <Route path="ai/history" element={<AIHistory />} />

              {/* 文档管理 */}
              <Route path="documents" element={<DocumentManagement />} />

              {/* 统计分析 */}
              <Route path="statistics" element={<Statistics />} />

              {/* 系统设置 */}
              <Route path="settings" element={<Settings />} />
            </Route>

            {/* 404页面 */}
            <Route path="*" element={<NotFound />} />
          </Routes>
        </Router>
      </AntApp>
    </ConfigProvider>
  );
};

export default App;