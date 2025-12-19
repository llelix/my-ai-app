# 设计文档

## 概述

将AI知识库应用的文件存储从本地文件系统迁移到MinIO对象存储。保持现有API不变，替换底层存储实现。

核心目标：
- 替换本地文件存储为MinIO
- 保持API兼容性
- 支持现有的分块上传功能
- 提供数据迁移工具

## 架构

简单的替换策略：
1. 保持现有的DocumentService接口不变
2. 将文件存储从本地文件系统改为MinIO
3. 使用MinIO的multipart upload替代现有的分块上传

```
DocumentService (不变)
    ↓
MinIO客户端 (新增)
    ↓  
MinIO服务器 (Docker)
```

## 组件和接口

### MinIO配置

```go
type MinIOConfig struct {
    Endpoint        string `mapstructure:"endpoint"`
    AccessKeyID     string `mapstructure:"access_key_id"`
    SecretAccessKey string `mapstructure:"secret_access_key"`
    UseSSL          bool   `mapstructure:"use_ssl"`
    Bucket          string `mapstructure:"bucket"`
}
```

### DocumentService修改

保持现有方法签名不变，内部实现改为使用MinIO：

```go
// 现有方法保持不变
func (s *DocumentService) Upload(file *multipart.FileHeader) (*models.Document, error)
func (s *DocumentService) InitUpload(fileName string, fileSize int64, fileHash string) (*models.UploadSession, error)
func (s *DocumentService) UploadChunk(sessionID string, chunkIndex int, data []byte) error
func (s *DocumentService) CompleteUpload(sessionID string) (*models.Document, error)
```

## 数据模型

### Document模型修改

只添加必要的MinIO字段：

```go
type Document struct {
    // 现有字段保持不变
    ID           uint      `json:"id" gorm:"primaryKey"`
    Name         string    `json:"name"`
    OriginalName string    `json:"original_name"`
    FilePath     string    `json:"file_path"`     // 改为存储MinIO对象键
    FileSize     int64     `json:"file_size"`
    FileHash     string    `json:"file_hash"`
    // ... 其他现有字段
}
```

### UploadSession模型修改

添加MinIO multipart upload ID：

```go
type UploadSession struct {
    // 现有字段保持不变
    ID           string    `json:"id" gorm:"primaryKey"`
    FileName     string    `json:"file_name"`
    FileSize     int64     `json:"file_size"`
    FileHash     string    `json:"file_hash"`
    
    // 新增MinIO字段
    UploadID     string    `json:"upload_id"`     // MinIO multipart upload ID
    
    // 现有字段保持不变
    ChunkSize    int64     `json:"chunk_size"`
    TotalChunks  int       `json:"total_chunks"`
    // ...
}
```

## 错误处理

使用现有的错误处理机制，MinIO错误直接返回给调用者。对于网络错误，使用简单的重试：

```go
func retryOnError(operation func() error, maxRetries int) error {
    var err error
    for i := 0; i <= maxRetries; i++ {
        err = operation()
        if err == nil {
            return nil
        }
        if i < maxRetries {
            time.Sleep(time.Second * time.Duration(i+1))
        }
    }
    return err
}
```

## 测试策略

### 后端测试
- **单元测试**: MinIO客户端连接、文件上传下载、分块上传流程
- **集成测试**: 使用Docker MinIO进行端到端测试、API兼容性测试

### 前端测试
- **单元测试**: 文件上传组件、进度显示组件、错误处理
- **集成测试**: 完整的文件上传流程、分块上传进度显示
- **用户界面测试**: 确保MinIO迁移后用户体验保持一致

## 正确性属性

*属性是应该在系统的所有有效执行中保持为真的特征或行为——本质上是关于系统应该做什么的正式陈述。属性作为人类可读规范和机器可验证正确性保证之间的桥梁。*

**属性 1: 文件上传往返一致性**
*对于任何* 成功上传的文件，从MinIO下载的文件内容应该与原始上传的文件内容完全相同
**验证: 需求 2.3, 3.1**

**属性 2: 分块上传完整性**
*对于任何* 分块上传会话，当所有分块上传完成后，合并的文件应该与原始文件内容相同
**验证: 需求 7.1, 7.2, 7.3**

**属性 3: 文件去重正确性**
*对于任何* 具有相同哈希值的文件，系统应该在MinIO中仅存储一个副本
**验证: 需求 2.5, 8.1, 8.2, 8.5**

**属性 4: API兼容性**
*对于任何* 现有的API调用，MinIO实现应该返回与本地存储实现相同格式的响应
**验证: 需求 2.1**

**属性 5: 前端功能正确性**
*对于任何* 前端文件操作（上传、下载、进度显示）也要功能正常
**验证: 需求 2.1, 2.4**