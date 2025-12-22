package preprocessing

import (
	"ai-knowledge-app/internal/preprocessing/core"
	"ai-knowledge-app/internal/preprocessing/repository"

	"gorm.io/gorm"
)

// Service 预处理服务主入口
type Service struct {
	*ServiceImpl
}

// ServiceImpl 文档预处理服务实现
type ServiceImpl struct {
	chunkRepo      core.DocumentChunkRepository
	statusRepo     core.ProcessingStatusRepository
	cascadeManager *repository.CascadeDeleteManager
	db             *gorm.DB
}

// NewService 创建新的文档预处理服务
func NewService(db *gorm.DB) *Service {
	service := &ServiceImpl{
		chunkRepo:      repository.NewDocumentChunkRepository(db),
		statusRepo:     repository.NewProcessingStatusRepository(db),
		cascadeManager: repository.NewCascadeDeleteManager(db),
		db:             db,
	}

	return &Service{ServiceImpl: service}
}

// GetSupportedFormats 获取支持的文档格式
func (s *ServiceImpl) GetSupportedFormats() []string {
	return []string{"pdf", "docx", "doc", "txt", "md"}
}
