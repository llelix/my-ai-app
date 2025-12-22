# 设计文档

## 概述

文档预处理系统是RAG应用的核心组件，负责将用户上传的各种格式文档转换为结构化的Markdown格式，并进行智能分块处理。系统采用MinerU作为文档转换引擎，使用LangChain Go进行文本分块，最终将处理结果存储到document_chunks表中，为后续的向量化和检索提供优化的数据结构。

## 架构

系统采用分层架构设计，包含以下主要层次：

```
┌─────────────────────────────────────────────────────────────┐
│                    API 层 (HTTP Handlers)                   │
├─────────────────────────────────────────────────────────────┤
│                    服务层 (Business Logic)                  │
├─────────────────────────────────────────────────────────────┤
│           处理引擎层 (MinerU + LangChain Go)                │
├─────────────────────────────────────────────────────────────┤
│                    数据访问层 (Repository)                  │
├─────────────────────────────────────────────────────────────┤
│              存储层 (Database + Object Storage)             │
└─────────────────────────────────────────────────────────────┘
```

### 核心组件

1. **文档预处理服务 (DocumentPreprocessingService)**
   - 协调整个预处理流程
   - 管理异步任务队列
   - 处理错误和重试逻辑

2. **MinerU集成模块 (MinerUProcessor)**
   - 封装MinerU http API调用
   - 处理不同文档格式的转换
   - 管理转换配置和参数

3. **文本分块器 (TextChunker)**
   - 使用LangChain Go的RecursiveCharacter分割器
   - 实现段落级别的智能分块
   - 支持自定义分块参数

4. **向量化处理器 (VectorizationProcessor)** - 预留接口
   - 为文档块生成向量嵌入
   - 支持多种嵌入模型
   - 批量向量化处理
   - *注：当前阶段仅定义接口，具体实现将在后续迭代中完成*

5. **文档块存储库 (DocumentChunkRepository)**
   - 管理document_chunks表的CRUD操作
   - 处理批量插入和更新
   - 维护文档块的关联关系

## 组件和接口

### 1. HTTP API接口

```go
// DocumentPreprocessingHandler 处理文档预处理的HTTP请求
type DocumentPreprocessingHandler struct {
    service *DocumentPreprocessingService
}

// ProcessDocument 处理单个文档
// POST /api/v1/documents/{id}/preprocess
func (h *DocumentPreprocessingHandler) ProcessDocument(c *gin.Context)

// ProcessVectorization 处理单个文档的向量化（预留）
// POST /api/v1/documents/{id}/vectorize
func (h *DocumentPreprocessingHandler) ProcessVectorization(c *gin.Context)

// GetProcessingStatus 获取处理状态
// GET /api/v1/documents/{id}/processing-status
func (h *DocumentPreprocessingHandler) GetProcessingStatus(c *gin.Context)

// GetVectorizationStatus 获取向量化状态（预留）
// GET /api/v1/documents/{id}/vectorization-status
func (h *DocumentPreprocessingHandler) GetVectorizationStatus(c *gin.Context)

// BatchProcessDocuments 批量处理文档
// POST /api/v1/documents/batch-preprocess
func (h *DocumentPreprocessingHandler) BatchProcessDocuments(c *gin.Context)

// GetBatchProcessingStatus 获取批量处理状态
// GET /api/v1/documents/batch-processing-status
func (h *DocumentPreprocessingHandler) GetBatchProcessingStatus(c *gin.Context)
```

### 2. 服务层接口

```go
// DocumentPreprocessingService 文档预处理服务接口
type DocumentPreprocessingService interface {
    ProcessDocument(ctx context.Context, documentID string) error
    GetProcessingStatus(ctx context.Context, documentID string) (*ProcessingStatus, error)
    BatchProcessDocuments(ctx context.Context, documentIDs []string) error
}

// ProcessingStatus 处理状态
type ProcessingStatus struct {
    DocumentID         string
    PreprocessStatus   ProcessingStatusType
    VectorizationStatus ProcessingStatusType  // 预留字段
    Progress           float64
    VectorizationProgress float64             // 预留字段
    Error              string
    VectorizationError string                 // 预留字段
    CreatedAt          time.Time
    UpdatedAt          time.Time
    CompletedAt        *time.Time
    ProcessingTime     time.Duration
}

// ProcessingStatusType 处理状态类型
type ProcessingStatusType string

const (
    StatusPending    ProcessingStatusType = "pending"
    StatusProcessing ProcessingStatusType = "processing"
    StatusCompleted  ProcessingStatusType = "completed"
    StatusFailed     ProcessingStatusType = "failed"
    StatusNotStarted ProcessingStatusType = "not_started"  // 用于向量化预留状态
)
```

### 3. MinerU处理器接口

```go
// MinerUProcessor MinerU文档转换处理器
type MinerUProcessor interface {
    ConvertToMarkdown(ctx context.Context, filePath string, options *ConversionOptions) (*MarkdownResult, error)
    SupportedFormats() []string
}

// ConversionOptions 转换选项
type ConversionOptions struct {
    Language        string
    Backend         string // pipeline, vlm-transformers, vlm-vllm-engine
    ParseMethod     string // auto, txt, ocr
    FormulaEnable   bool
    TableEnable     bool
    ExtractImages   bool
}

// MarkdownResult 转换结果
type MarkdownResult struct {
    Content     string
    Images      []ImageInfo
    Metadata    map[string]interface{}
    ProcessTime time.Duration
}
```

### 4. 文本分块器接口

```go
// TextChunker 文本分块器接口
type TextChunker interface {
    ChunkText(ctx context.Context, text string, options *ChunkingOptions) ([]DocumentChunk, error)
}

// ChunkingOptions 分块选项
type ChunkingOptions struct {
    ChunkSize    int
    ChunkOverlap int
    Separators   []string
}

// DocumentChunk 文档块
type DocumentChunk struct {
    ID          string
    DocumentID  string
    Content     string
    ChunkIndex  int
    StartOffset int
    EndOffset   int
    Metadata    map[string]interface{}
}
```

### 5. 向量化处理器接口（预留）

```go
// VectorizationProcessor 向量化处理器接口（预留扩展）
type VectorizationProcessor interface {
    // GenerateEmbeddings 为文档块生成向量嵌入
    GenerateEmbeddings(ctx context.Context, chunks []DocumentChunk, options *EmbeddingOptions) ([]DocumentEmbedding, error)
    
    // BatchGenerateEmbeddings 批量生成向量嵌入
    BatchGenerateEmbeddings(ctx context.Context, chunkBatches [][]DocumentChunk, options *EmbeddingOptions) ([][]DocumentEmbedding, error)
    
    // GetEmbeddingDimensions 获取嵌入向量的维度
    GetEmbeddingDimensions() int
    
    // SupportedModels 获取支持的嵌入模型列表
    SupportedModels() []string
}

// EmbeddingOptions 向量化选项
type EmbeddingOptions struct {
    Model       string  // 嵌入模型名称
    BatchSize   int     // 批处理大小
    Normalize   bool    // 是否标准化向量
    Dimensions  int     // 向量维度（可选）
}

// DocumentEmbedding 文档嵌入
type DocumentEmbedding struct {
    ChunkID     string
    Vector      []float32
    Model       string
    Dimensions  int
    CreatedAt   time.Time
}

// 注：此接口当前仅作为预留设计，具体实现将在向量化模块开发时完成
```

## 数据模型

### 1. 文档块表结构

```sql
CREATE TABLE document_chunks (
    id VARCHAR(36) PRIMARY KEY,
    document_id VARCHAR(36) NOT NULL,
    content TEXT NOT NULL,
    chunk_index INTEGER NOT NULL,
    start_offset INTEGER NOT NULL,
    end_offset INTEGER NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    INDEX idx_document_chunks_document_id (document_id),
    INDEX idx_document_chunks_chunk_index (document_id, chunk_index)
);
```

### 2. 处理状态表结构

```sql
CREATE TABLE document_processing_status (
    id VARCHAR(36) PRIMARY KEY,
    document_id VARCHAR(36) NOT NULL UNIQUE,
    status ENUM('pending', 'processing', 'completed', 'failed') NOT NULL,
    progress DECIMAL(5,2) DEFAULT 0.00,
    error_message TEXT,
    processing_options JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    INDEX idx_processing_status_document_id (document_id),
    INDEX idx_processing_status_status (status)
);
```

### 3. 文档嵌入表结构（预留）

```sql
-- 预留的向量嵌入表结构，用于未来向量化功能扩展
CREATE TABLE document_embeddings (
    id VARCHAR(36) PRIMARY KEY,
    chunk_id VARCHAR(36) NOT NULL,
    vector_data BLOB NOT NULL,  -- 存储向量数据
    model_name VARCHAR(100) NOT NULL,
    dimensions INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (chunk_id) REFERENCES document_chunks(id) ON DELETE CASCADE,
    INDEX idx_embeddings_chunk_id (chunk_id),
    INDEX idx_embeddings_model (model_name)
);

-- 注：此表结构当前仅作为预留设计，将在向量化模块开发时正式创建和使用
```

## 正确性属性

*属性是应该在系统的所有有效执行中保持为真的特征或行为——本质上是关于系统应该做什么的正式陈述。属性作为人类可读规范和机器可验证正确性保证之间的桥梁。*

### 属性 1: PDF转换一致性
*对于任何*有效的PDF文档，使用MinerU转换应该产生有效的Markdown格式输出
**验证: 需求 1.1**

### 属性 2: 公式保留性
*对于任何*包含LaTeX公式的文档，转换后的Markdown应该保留所有公式的LaTeX表示
**验证: 需求 1.2**

### 属性 3: 表格格式转换
*对于任何*包含表格的文档，转换后应该生成有效的Markdown表格格式
**验证: 需求 1.3**

### 属性 4: 图像引用完整性
*对于任何*包含图像的文档，转换后的Markdown应该包含所有图像的正确引用
**验证: 需求 1.4**

### 属性 5: 结构信息保留
*对于任何*文档转换操作，输出的Markdown应该包含原始文档的结构信息
**验证: 需求 1.5**

### 属性 6: 分块连续性
*对于任何*文本分块操作，相邻块之间应该保持正确的重叠和连续性
**验证: 需求 3.3, 7.2**

### 属性 7: 块唯一标识符
*对于任何*分块操作，每个生成的文档块应该具有唯一的标识符
**验证: 需求 3.5**

### 属性 8: 存储完整性
*对于任何*文档块存储操作，存储的数据应该包含完整的内容、元数据和父文档引用
**验证: 需求 4.2**

### 属性 9: 级联删除一致性
*对于任何*文档删除操作，所有相关的文档块应该同时被删除
**验证: 需求 4.4**

### 属性 10: 更新操作原子性
*对于任何*文档更新操作，旧块的清理和新块的生成应该作为原子操作完成
**验证: 需求 4.3**

### 属性 11: 错误处理完整性
*对于任何*处理失败的情况，系统应该记录详细错误信息并返回用户友好的错误消息
**验证: 需求 5.1**

### 属性 12: 异步队列可靠性
*对于任何*文档处理请求，任务应该被正确加入异步处理队列
**验证: 需求 6.1**

### 属性 13: 质量验证完整性
*对于任何*转换操作，系统应该验证输出结果的完整性和正确性
**验证: 需求 7.1, 7.4**

### 属性 14: 集成触发一致性
*对于任何*预处理完成的文档，系统应该触发后续的向量化处理流程
**验证: 需求 8.1**

### 属性 15: 状态查询一致性
*对于任何*处理状态查询请求，返回的状态应该准确反映当前的处理进度和状态
**验证: 需求 10.1, 10.2**

### 属性 16: 按钮状态同步
*对于任何*文档的处理操作，前端按钮状态应该与后端处理状态保持同步
**验证: 需求 9.2, 9.4**

## 错误处理

### 1. 转换错误处理
- MinerU处理失败时的重试机制
- 不支持格式的优雅降级
- 内存不足时的资源管理

### 2. 分块错误处理
- 文本过长时的分段处理
- 编码问题的自动检测和修复
- 分块参数异常的验证

### 3. 存储错误处理
- 数据库连接失败的重试
- 事务回滚机制
- 数据完整性验证

### 4. 集成错误处理
- 外部服务不可用的降级策略
- API调用超时的处理
- 消息队列异常的恢复

## 前端集成设计

### 1. 文档列表UI组件

```typescript
// 文档处理状态接口
interface DocumentProcessingStatus {
  documentId: string;
  preprocessStatus: 'pending' | 'processing' | 'completed' | 'failed' | 'not_started';
  vectorizationStatus: 'pending' | 'processing' | 'completed' | 'failed' | 'not_started';
  progress: number;
  vectorizationProgress: number;
  error?: string;
  vectorizationError?: string;
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
  processingTime?: number;
}

// 文档操作按钮组件
interface DocumentActionButtonsProps {
  documentId: string;
  status: DocumentProcessingStatus;
  onPreprocess: (documentId: string) => void;
  onVectorize: (documentId: string) => void;
}
```

### 2. 状态轮询机制

```typescript
// 状态轮询服务
class ProcessingStatusService {
  private pollingIntervals: Map<string, NodeJS.Timeout> = new Map();
  
  startPolling(documentId: string, callback: (status: DocumentProcessingStatus) => void): void;
  stopPolling(documentId: string): void;
  getStatus(documentId: string): Promise<DocumentProcessingStatus>;
}
```

### 3. 实时状态更新

- 使用轮询机制每2秒检查处理状态
- 处理完成后自动停止轮询
- 支持WebSocket连接进行实时状态推送（可选扩展）
- 错误状态的用户友好显示和重试机制

## 测试策略

### 单元测试和属性测试的双重方法

系统将采用单元测试和基于属性的测试相结合的方法：
- 单元测试验证特定示例、边界情况和错误条件
- 属性测试验证应该在所有输入中保持的通用属性
- 两者结合提供全面覆盖：单元测试捕获具体错误，属性测试验证通用正确性

### 单元测试要求

单元测试通常涵盖：
- 演示正确行为的特定示例
- 组件之间的集成点
- 单元测试很有用，但避免写太多。属性测试的工作是处理大量输入的覆盖。

### 基于属性的测试要求

系统将使用Go的property-based testing库进行属性测试。每个属性测试将配置为运行最少100次迭代，因为属性测试过程是随机的。

每个基于属性的测试必须使用注释明确引用设计文档中的正确性属性，使用以下确切格式：'**Feature: document-preprocessing, Property {number}: {property_text}**'

每个正确性属性必须由单个基于属性的测试实现。

这些要求在测试策略部分明确说明。