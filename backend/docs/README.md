# API 文档

## Swagger 文档

本项目使用 Swagger 自动生成 API 文档，提供交互式的 API 测试界面。

### 访问 Swagger UI

启动后端服务器后，可以通过以下地址访问 Swagger 文档：

```
http://localhost:8080/swagger/index.html
```

### 主要 API 端点

#### 系统相关
- `GET /health` - 健康检查
- `GET /debug/config` - 调试配置信息

#### 知识库管理
- `GET /api/v1/knowledge` - 获取知识列表（支持分页、搜索、过滤）
- `GET /api/v1/knowledge/{id}` - 获取单个知识条目
- `POST /api/v1/knowledge` - 创建新的知识条目
- `PUT /api/v1/knowledge/{id}` - 更新知识条目
- `DELETE /api/v1/knowledge/{id}` - 删除知识条目
- `GET /api/v1/knowledge/search` - 搜索知识
- `GET /api/v1/knowledge/{id}/related` - 获取相关知识
- `POST /api/v1/knowledge/{id}/view` - 增加查看次数

#### AI 查询
- `POST /api/v1/ai/query` - AI 智能查询
- `GET /api/v1/ai/history` - 获取查询历史
- `DELETE /api/v1/ai/history/{id}` - 删除查询历史
- `GET /api/v1/ai/history/stats` - 获取查询统计
- `POST /api/v1/ai/feedback` - 提交反馈
- `GET /api/v1/ai/models` - 获取可用模型

#### 分类管理
- `GET /api/v1/categories` - 获取分类列表
- `GET /api/v1/categories/{id}` - 获取单个分类
- `POST /api/v1/categories` - 创建分类
- `PUT /api/v1/categories/{id}` - 更新分类
- `DELETE /api/v1/categories/{id}` - 删除分类
- `GET /api/v1/categories/{id}/knowledges` - 获取分类下的知识

#### 标签管理
- `GET /api/v1/tags` - 获取标签列表
- `GET /api/v1/tags/{id}` - 获取单个标签
- `POST /api/v1/tags` - 创建标签
- `PUT /api/v1/tags/{id}` - 更新标签
- `DELETE /api/v1/tags/{id}` - 删除标签
- `GET /api/v1/tags/{id}/knowledges` - 获取标签下的知识
- `GET /api/v1/tags/popular` - 获取热门标签

#### 文档管理
- `POST /api/v1/documents/upload` - 上传文档
- `GET /api/v1/documents` - 获取文档列表
- `GET /api/v1/documents/{id}` - 获取文档详情
- `DELETE /api/v1/documents/{id}` - 删除文档
- `PUT /api/v1/documents/{id}/description` - 更新文档描述
- `GET /api/v1/documents/{id}/download` - 下载文档

#### 统计分析
- `GET /api/v1/stats/overview` - 概览统计
- `GET /api/v1/stats/knowledge` - 知识库统计
- `GET /api/v1/stats/queries` - 查询统计

#### 文件上传
- `POST /api/v1/files/upload` - 文件上传

### 使用示例

#### 创建知识条目
```bash
curl -X POST "http://localhost:8080/api/v1/knowledge" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Go语言基础",
    "content": "Go是一种开源的编程语言...",
    "summary": "Go语言入门介绍",
    "category_id": 1,
    "tags": ["编程", "Go", "后端"],
    "is_published": true
  }'
```

#### AI 查询
```bash
curl -X POST "http://localhost:8080/api/v1/ai/query" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "如何学习Go语言？",
    "temperature": 0.7,
    "max_tokens": 2000
  }'
```

### 响应格式

所有 API 响应都遵循统一的格式：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 实际数据内容
  }
}
```

错误响应格式：
```json
{
  "code": 400,
  "message": "错误描述"
}
```

### 分页响应

列表类 API 支持分页，响应格式：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 10,
    "total_pages": 10
  }
}
```

### 认证

目前 API 不需要认证，但在生产环境中建议添加适当的认证机制。

### 开发工具

- 使用 `make swagger` 重新生成文档
- 使用 `make dev` 启动开发服务器
- 使用 `make docker` 启动 PostgreSQL 数据库