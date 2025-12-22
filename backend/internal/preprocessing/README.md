# 文档预处理模块

## 模块概述

文档预处理模块负责处理上传的文档，包括格式转换、文本分块、质量验证等功能。该模块采用清晰的分层架构，便于维护和扩展。

## 目录结构

```
preprocessing/
├── core/                    # 核心类型和接口定义
│   ├── types.go            # 业务类型定义
│   ├── interfaces.go       # 核心接口定义
│   ├── errors.go           # 错误类型定义
│   └── utils.go            # 工具函数
├── processor/              # 处理器实现
│   ├── interfaces.go       # 处理器接口
│   ├── mineru_processor.go # MinerU文档转换处理器
│   ├── text_chunker.go     # 文本分块器
│   └── quality_validator.go # 质量验证器
├── repository/             # 数据访问层
│   ├── models.go           # 数据库模型
│   ├── chunk_repository.go # 文档块存储库
│   └── status_repository.go # 状态存储库
├── queue/                  # 异步处理队列
│   ├── processing_queue.go # 处理队列实现
│   ├── task.go            # 任务定义
│   └── metrics.go         # 队列监控指标
├── monitoring/             # 监控和指标
│   ├── metrics.go         # 性能指标
│   └── health.go          # 健康检查
├── service.go             # 主服务入口
├── config.go              # 配置管理
└── README.md              # 本文档
```

## 架构设计

### 分层架构

1. **Core层** (`core/`): 定义核心业务类型、接口和错误
2. **Processor层** (`processor/`): 实现具体的处理逻辑
3. **Repository层** (`repository/`): 数据访问和持久化
4. **Queue层** (`queue/`): 异步任务处理
5. **Monitoring层** (`monitoring/`): 监控和指标收集

### 依赖关系

```
Service → Processor → Core
       → Repository → Core
       → Queue → Core
       → Monitoring → Core
```

## 核心组件

### 1. 文档处理服务 (Service)

主要服务接口，提供：
- 同步/异步文档处理
- 处理状态查询
- 批量处理
- 统计信息

### 2. MinerU处理器 (MinerUProcessor)

负责文档格式转换：
- PDF → Markdown
- 支持多种解析方法
- 图像提取
- 表格和公式处理

### 3. 文本分块器 (TextChunker)

将长文本分割为可管理的块：
- 可配置块大小和重叠
- 智能分割点识别
- 保持语义完整性

### 4. 处理队列 (ProcessingQueue)

异步任务处理：
- 优先级队列
- 工作协程池
- 任务状态跟踪
- 错误重试机制

## 使用示例

### 基本使用

```go
// 创建服务
service := preprocessing.NewService(
    mineruProcessor,
    textChunker,
    db,
)

// 同步处理文档
err := service.ProcessDocument(ctx, documentID)

// 异步处理文档
task, err := service.ProcessDocumentAsync(documentID, priority)

// 查询处理状态
status, err := service.GetProcessingStatus(ctx, documentID)
```

### 配置示例

```go
// 转换选项
options := &core.ConversionOptions{
    Language:      "zh",
    Backend:       "pipeline",
    ParseMethod:   "auto",
    FormulaEnable: true,
    TableEnable:   true,
    ExtractImages: true,
}

// 分块选项
chunkOptions := &core.ChunkingOptions{
    ChunkSize:    1000,
    ChunkOverlap: 200,
    Separators:   []string{"\n\n", "\n", ".", "!", "?"},
}
```

## 扩展点

### 添加新的处理器

1. 在 `processor/interfaces.go` 中定义接口
2. 在 `processor/` 目录下实现具体处理器
3. 在服务中注册新处理器

### 添加新的存储后端

1. 在 `core/interfaces.go` 中定义存储接口
2. 在 `repository/` 目录下实现具体存储
3. 通过依赖注入使用新存储

## 监控和指标

### 性能指标

- 处理时间统计
- 队列长度监控
- 成功/失败率
- 资源使用情况

### 健康检查

- 数据库连接状态
- 队列健康状态
- 处理器可用性

## 错误处理

### 错误类型

- `ValidationError`: 数据验证错误
- `ProcessingError`: 处理过程错误
- 标准错误: 文档未找到、配置错误等

### 错误恢复

- 自动重试机制
- 错误状态记录
- 失败任务清理

## 测试

### 单元测试

每个组件都有对应的测试文件：
- `*_test.go`: 单元测试
- 使用模拟对象进行隔离测试

### 集成测试

- 端到端处理流程测试
- 数据库集成测试
- 队列处理测试

## 性能优化

### 批处理

- 批量数据库操作
- 批量文档处理
- 减少网络往返

### 并发处理

- 工作协程池
- 并发安全的数据结构
- 资源池管理

### 内存管理

- 流式处理大文件
- 及时释放资源
- 内存使用监控