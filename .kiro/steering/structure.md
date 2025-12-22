# 项目结构与组织

## 根目录布局

```
ai-knowledge-app/
├── backend/          # Go后端API服务
├── frontend/         # React前端应用
├── README.md         # 项目概述和设置说明
└── .kiro/           # Kiro AI助手配置
```

## 后端结构 (`backend/`)

### 标准Go项目布局

```
backend/
├── cmd/
│   └── server/           # 应用程序入口点
│       ├── main.go       # 主应用程序入口
│       └── logs/         # 应用程序日志
├── internal/             # 私有应用程序代码
│   ├── api/             # HTTP处理器和路由
│   ├── ai/              # AI服务集成
│   ├── config/          # 配置管理
│   ├── middleware/      # HTTP中间件
│   ├── models/          # 数据模型和数据库模式
│   └── service/         # 业务逻辑服务
├── pkg/                 # 公共库代码
│   ├── database/        # 数据库连接和工具
│   ├── logger/          # 日志工具
│   └── utils/           # 通用工具函数
├── data/                # 应用程序数据存储
│   └── documents/       # 文档存储，按日期层次结构
├── .env.example         # 环境配置模板
├── docker-compose.yml   # PostgreSQL + pgvector设置
├── go.mod              # Go模块定义
└── go.sum              # 依赖校验和
```

### 关键后端模式

- **整洁架构**：使用 `internal/` 进行业务逻辑的关注点分离
- **处理器模式**：`internal/api/` 中的API处理器，具有清晰的路由结构
- **服务层**：业务逻辑封装在 `internal/service/` 中
- **仓储模式**：通过GORM模型抽象数据库操作
- **中间件链**：横切关注点（日志、CORS、验证）在 `internal/middleware/` 中

## 前端结构 (`frontend/`)

### React应用布局

```
frontend/
├── src/
│   ├── components/      # 可重用UI组件（当前为空）
│   ├── pages/           # 按功能组织的页面级组件
│   │   ├── AI/          # AI聊天和历史页面
│   │   ├── Category/    # 分类管理
│   │   ├── Knowledge/   # 知识CRUD操作
│   │   ├── NotFound/    # 404错误页面
│   │   ├── Settings/    # 应用设置
│   │   ├── Statistics/  # 分析和报告
│   │   └── Tag/         # 标签管理
│   ├── layouts/         # 布局组件（MainLayout）
│   ├── services/        # API通信层
│   ├── store/           # Zustand状态管理
│   ├── types/           # TypeScript类型定义
│   ├── utils/           # 工具函数和常量
│   ├── preprocessing/   # 文档处理工具
│   ├── App.tsx          # 主应用组件
│   └── main.tsx         # 应用入口点
├── public/              # 静态资源
├── dist/                # 生产构建输出
├── package.json         # Node.js依赖和脚本
├── vite.config.ts       # Vite构建配置
└── tsconfig.json        # TypeScript配置
```

### 关键前端模式

- **基于功能的组织**：页面按功能区域分组
- **服务层**：API调用在 `services/` 目录中抽象
- **类型安全**：全面使用TypeScript和共享类型
- **组件组合**：Ant Design组件与自定义样式
- **状态管理**：Zustand用于轻量级、可预测的状态

## API结构

### RESTful端点组织

```
/api/v1/
├── /knowledge          # 知识库CRUD操作
├── /categories         # 分类管理
├── /tags              # 标签管理
├── /ai                # AI查询和历史
├── /documents         # 文档上传和处理
├── /stats             # 分析和统计
└── /files             # 文件上传工具
```

### 数据库模式组织

- **核心实体**：Knowledge、Category、Tag，具有适当的关系
- **AI功能**：QueryHistory、Feedback用于AI交互
- **文档处理**：Document、DocumentChunk、UploadSession
- **向量存储**：与pgvector集成进行语义搜索

## 配置管理

### 基于环境的配置

- **开发环境**：SQLite数据库、调试日志、本地开发CORS
- **生产环境**：PostgreSQL + pgvector、结构化日志、安全头
- **AI服务**：可配置的提供商（OpenAI、Claude、自定义端点）

### 文件组织约定

- **Go文件**：文件使用snake_case，导出函数使用PascalCase
- **React文件**：组件使用PascalCase，工具使用camelCase
- **API路由**：RESTful命名，具有清晰的资源层次结构
- **数据库**：一致的命名，适当的索引和关系

## 开发工作流

### 代码组织原则

1. **关注点分离**：层之间的清晰边界
2. **功能分组**：相关功能组合在一起
3. **依赖方向**：依赖向内流动（整洁架构）
4. **配置外部化**：基于环境的配置
5. **错误处理**：一致的错误响应和日志记录

### 文件命名约定

- **后端**：文件使用 `snake_case.go`，类型使用 `PascalCase`
- **前端**：组件使用 `PascalCase.tsx`，工具使用 `camelCase.ts`
- **配置**：`.env` 文件使用 `UPPER_CASE` 变量
- **文档**：相关目录中的 `README.md` 文件