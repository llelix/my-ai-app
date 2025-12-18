# 技术栈与构建系统

## 架构

前后端分离的全栈应用：
- **后端**：基于Go的REST API服务器
- **前端**：React SPA + TypeScript
- **数据库**：SQLite（开发环境）/ PostgreSQL（生产环境）+ pgvector扩展
- **AI集成**：OpenAI兼容API（OpenAI、Claude、本地模型）

## 后端技术栈

### 核心框架与库
- **Web框架**：Gin（高性能HTTP Web框架）
- **ORM**：GORM，支持自动迁移
- **配置管理**：Viper，基于环境的配置管理
- **日志记录**：Logrus，结构化日志和文件轮转（lumberjack）
- **数据验证**：go-playground/validator，请求验证
- **AI集成**：langchaingo，LLM交互
- **向量数据库**：pgvector，语义搜索能力

### 关键依赖
- `github.com/gin-gonic/gin` - Web框架
- `gorm.io/gorm` - ORM，支持SQLite/PostgreSQL驱动
- `github.com/spf13/viper` - 配置管理
- `github.com/sirupsen/logrus` - 结构化日志
- `github.com/tmc/langchaingo` - LLM集成
- `github.com/pgvector/pgvector-go` - 向量操作

## 前端技术栈

### 核心框架与库
- **框架**：React 19 + TypeScript
- **构建工具**：Vite，快速开发和构建
- **UI库**：Ant Design 6.0，一致的UI组件
- **状态管理**：Zustand，轻量级状态管理
- **路由**：React Router DOM v7
- **HTTP客户端**：Axios，API通信
- **Markdown**：react-markdown，语法高亮

### 关键依赖
- `react` & `react-dom` - React核心框架
- `antd` - UI组件库
- `zustand` - 状态管理
- `react-router-dom` - 客户端路由
- `axios` - HTTP客户端
- `react-markdown` - Markdown渲染
- `react-syntax-highlighter` - 代码语法高亮

## 开发命令

### 后端命令
```bash
# 进入后端目录
cd backend

# 安装依赖
go mod tidy

# 运行开发服务器
cd cmd/server && go run .

# 生产环境构建
go build -o ai-knowledge-app cmd/server/main.go

# 运行测试
go test ./...

# 使用Docker启动PostgreSQL
docker-compose up -d postgres
```

### 前端命令
```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 运行开发服务器（代理到后端）
npm run dev

# 生产环境构建
npm run build

# 预览生产构建
npm run preview

# 运行代码检查
npm run lint
```

## 环境配置

### 后端环境变量
- 复制 `.env.example` 到 `.env` 并配置：
  - `SERVER_PORT=8080` - API服务器端口
  - `DB_TYPE=sqlite|postgres` - 数据库类型
  - `OPENAI_API_KEY` - AI服务API密钥
  - `OPENAI_BASE_URL` - AI服务端点
  - `OPENAI_MODEL` - 使用的AI模型

### 开发代理
- 前端Vite开发服务器将 `/api` 请求代理到 `http://localhost:8080`
- 实现无缝的全栈开发体验

## 数据库设置

### SQLite（开发环境）
- 自动在 `./data/app.db` 创建数据库
- 无需额外设置

### PostgreSQL（生产环境）
- 使用提供的 `docker-compose.yml` 启动本地PostgreSQL + pgvector
- 需要pgvector扩展进行向量相似性搜索
- 自动迁移处理模式创建

## 构建与部署

### 开发工作流
1. 启动后端：`cd backend/cmd/server && go run .`
2. 启动前端：`cd frontend && npm run dev`
3. 访问应用：`http://localhost:5173`

### 生产构建
1. 后端：`go build -o ai-knowledge-app cmd/server/main.go`
2. 前端：`npm run build`（输出到 `dist/`）
3. 通过后端或独立Web服务器提供前端静态文件