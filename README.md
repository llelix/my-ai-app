# AI知识库查询应用

## 项目架构

这是一个全栈AI知识库查询应用，采用前后端分离架构：

```
ai-knowledge-app/
├── backend/          # Go后端API服务
│   ├── cmd/         # 应用入口点
│   ├── internal/    # 内部包
│   │   ├── api/     # API路由和处理器
│   │   ├── config/  # 配置管理
│   │   ├── models/  # 数据模型
│   │   ├── service/ # 业务逻辑服务
│   │   └── ai/      # AI集成服务
│   ├── pkg/         # 公共包
│   ├── go.mod       # Go模块定义
│   └── go.sum       # 依赖锁定文件
├── frontend/        # React前端应用
│   ├── src/         # 源代码
│   │   ├── components/  # React组件
│   │   ├── pages/       # 页面组件
│   │   ├── services/    # API调用服务
│   │   └── hooks/       # 自定义Hook
│   ├── package.json  # Node.js依赖
│   └── public/       # 静态资源
└── docs/            # 项目文档
```

## 技术栈

### 后端 (Go)
- **Web框架**: Gin
- **数据库**: SQLite (开发) / PostgreSQL (生产)
- **ORM**: GORM
- **AI集成**: OpenAI/Claude API
- **配置管理**: Viper
- **日志**: Logrus

### 前端 (React)
- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **状态管理**: Zustand
- **UI组件**: Ant Design
- **HTTP客户端**: Axios
- **路由**: React Router

## 核心功能

1. **知识库管理**
   - 添加/编辑/删除知识条目
   - 知识条目分类和标签
   - 全文搜索功能

2. **AI查询**
   - 自然语言查询接口
   - 基于知识库的智能回答
   - 查询历史记录

3. **用户界面**
   - 响应式设计
   - 实时查询结果
   - 知识条目可视化展示

## API设计

### 知识库管理
- `GET /api/knowledge` - 获取知识条目列表
- `POST /api/knowledge` - 创建新知识条目
- `PUT /api/knowledge/{id}` - 更新知识条目
- `DELETE /api/knowledge/{id}` - 删除知识条目
- `GET /api/knowledge/search?q={query}` - 搜索知识条目

### AI查询
- `POST /api/ai/query` - AI查询接口
- `GET /api/ai/history` - 查询历史
- `DELETE /api/ai/history/{id}` - 删除查询历史

## 开发环境要求

- Go 1.21+
- Node.js 18+
- SQLite 3.x

## 运行项目
cd ./backend/cmd/server/&&go run .
cd ./frontend/&&npm run dev