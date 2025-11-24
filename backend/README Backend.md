# AI知识库查询应用 - 后端API

## 功能特性

✅ **完整的功能实现**
- 知识库CRUD操作
- 分类和标签管理
- AI智能查询（支持OpenAI兼容API）
- 查询历史和统计
- 全文搜索和过滤
- RESTful API设计

✅ **技术特性**
- Gin高性能Web框架
- GORM数据库ORM
- 中间件支持（日志、CORS、错误处理）
- 配置管理（Viper）
- 结构化日志（Logrus）
- SQLite/PostgreSQL双数据库支持

## 快速启动

### 1. 环境要求
- Go 1.21+
- SQLite 3.x

### 2. 安装依赖
```bash
go mod tidy
```

### 3. 配置环境变量
复制 `.env.example` 到 `.env` 并修改配置：
```bash
cp .env.example .env
```

重点是配置AI服务：
```env
# OpenAI兼容API配置
OPENAI_API_KEY=your_api_key_here
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-3.5-turbo

# 或者使用国内大模型
# 通义千问
OPENAI_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
OPENAI_API_KEY=your_dashscope_api_key
OPENAI_MODEL=qwen-plus

# 月之暗面
OPENAI_BASE_URL=https://api.moonshot.cn/v1
OPENAI_API_KEY=your_moonshot_api_key
OPENAI_MODEL=moonshot-v1-8k
```

### 4. 运行应用
```bash
# 开发模式
go run cmd/server/main.go

# 构建运行
go build -o ai-knowledge-app cmd/server/main.go
./ai-knowledge-app
```

## API文档

### 基础信息
- **Base URL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json`

### 主要端点

#### 知识库管理
- `GET /knowledge` - 获取知识列表
- `POST /knowledge` - 创建知识
- `GET /knowledge/:id` - 获取单个知识
- `PUT /knowledge/:id` - 更新知识
- `DELETE /knowledge/:id` - 删除知识
- `GET /knowledge/search` - 搜索知识

#### 分类管理
- `GET /categories` - 获取分类列表
- `POST /categories` - 创建分类
- `PUT /categories/:id` - 更新分类
- `DELETE /categories/:id` - 删除分类

#### 标签管理
- `GET /tags` - 获取标签列表
- `POST /tags` - 创建标签
- `GET /tags/popular` - 获取热门标签

#### AI查询
- `POST /ai/query` - AI智能查询
- `GET /ai/history` - 查询历史
- `GET /ai/models` - 获取支持的模型

### 示例请求

#### 创建知识
```bash
curl -X POST http://localhost:8080/api/v1/knowledge \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Go语言基础",
    "content": "Go是Google开发的编程语言...",
    "category_id": 1,
    "tags": ["Go", "编程"],
    "is_published": true
  }'
```

#### AI查询
```bash
curl -X POST http://localhost:8080/api/v1/ai/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "什么是Go语言？",
    "model": "gpt-3.5-turbo"
  }'
```

## 配置说明

### 数据库配置
```env
# SQLite（开发）
DB_TYPE=sqlite
DB_PATH=./data/app.db

# PostgreSQL（生产）
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=ai_knowledge_db
```

### AI模型支持
支持所有OpenAI API兼容的模型：
- OpenAI: `gpt-3.5-turbo`, `gpt-4`
- 通义千问: `qwen-plus`, `qwen-max`
- 月之暗面: `moonshot-v1-8k`, `moonshot-v1-32k`
- 智谱清言: `glm-3-turbo`, `glm-4`
- Claude: `claude-3-sonnet`, `claude-3-opus`

## 项目结构
```
backend/
├── cmd/server/          # 应用入口
├── internal/
│   ├── api/            # API处理器
│   ├── config/         # 配置管理
│   ├── models/         # 数据模型
│   ├── ai/            # AI服务
│   └── middleware/    # 中间件
├── pkg/
│   ├── database/      # 数据库工具
│   ├── logger/        # 日志工具
│   └── utils/         # 工具函数
├── .env.example       # 环境配置模板
└── go.mod            # Go模块
```

## 开发说明

### 数据迁移
应用启动时会自动执行：
- 数据库表创建
- 索引建立
- 种子数据插入

### 日志查看
```bash
# 查看应用日志
tail -f logs/app.log

# 查看错误日志
grep "ERROR" logs/app.log
```

### 性能优化
- 数据库查询预加载
- 分页查询支持
- 索引优化
- 连接池管理

## API错误码
- `200` - 成功
- `400` - 请求参数错误
- `401` - 未授权
- `404` - 资源不存在
- `422` - 数据验证失败
- `500` - 服务器内部错误
- `503` - 服务不可用

## 许可证
MIT License