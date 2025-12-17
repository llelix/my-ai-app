package service

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-knowledge-app/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentService struct {
	db        *gorm.DB
	uploadDir string
	tempDir   string
}

func NewDocumentService(db *gorm.DB) *DocumentService {
	uploadDir := "uploads"
	tempDir := "temp"
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(tempDir, 0755)
	return &DocumentService{
		db:        db,
		uploadDir: uploadDir,
		tempDir:   tempDir,
	}
}

// CheckFile 检查文件是否已存在（秒传）
func (s *DocumentService) CheckFile(hash string, size int64) (*models.Document, bool) {
	var doc models.Document
	err := s.db.Where("file_hash = ? AND file_size = ? AND status = ?", hash, size, "completed").First(&doc).Error
	if err == nil {
		return &doc, true
	}
	return nil, false
}

// InitUpload 初始化上传会话
func (s *DocumentService) InitUpload(fileName string, fileSize int64, fileHash string) (*models.UploadSession, error) {
	// 检查是否可以秒传
	if doc, exists := s.CheckFile(fileHash, fileSize); exists {
		return nil, fmt.Errorf("file already exists: %d", doc.ID)
	}

	chunkSize := int64(1048576) // 1MB
	totalChunks := int((fileSize + chunkSize - 1) / chunkSize)

	sessionID := uuid.New().String()
	tempDir := filepath.Join(s.tempDir, sessionID)
	os.MkdirAll(tempDir, 0755)

	session := &models.UploadSession{
		ID:          sessionID,
		FileName:    fileName,
		FileSize:    fileSize,
		FileHash:    fileHash,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		TempDir:     tempDir,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	return session, s.db.Create(session).Error
}

// UploadChunk 上传分片
func (s *DocumentService) UploadChunk(sessionID string, chunkIndex int, data []byte) error {
	var session models.UploadSession
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return err
	}

	if time.Now().After(session.ExpiresAt) {
		return fmt.Errorf("upload session expired")
	}

	chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk_%d", chunkIndex))
	return os.WriteFile(chunkPath, data, 0644)
}

// CompleteUpload 完成上传
func (s *DocumentService) CompleteUpload(sessionID string) (*models.Document, error) {
	var session models.UploadSession
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}

	// 合并分片
	ext := filepath.Ext(session.FileName)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), session.FileName)
	finalPath := filepath.Join(s.uploadDir, filename)

	finalFile, err := os.Create(finalPath)
	if err != nil {
		return nil, err
	}
	defer finalFile.Close()

	// 按顺序合并分片
	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk_%d", i))
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return nil, err
		}
		finalFile.Write(chunkData)
	}

	// 验证文件哈希
	finalFile.Seek(0, 0)
	hash := sha256.New()
	io.Copy(hash, finalFile)
	calculatedHash := fmt.Sprintf("%x", hash.Sum(nil))

	if calculatedHash != session.FileHash {
		os.Remove(finalPath)
		return nil, fmt.Errorf("file hash mismatch")
	}

	// 创建文档记录
	doc := &models.Document{
		Name:         strings.TrimSuffix(session.FileName, ext),
		OriginalName: session.FileName,
		FilePath:     finalPath,
		FileSize:     session.FileSize,
		FileHash:     session.FileHash,
		Extension:    ext,
		Status:       "completed",
	}

	if err := s.db.Create(doc).Error; err != nil {
		os.Remove(finalPath)
		return nil, err
	}

	// 清理临时文件和会话
	os.RemoveAll(session.TempDir)
	s.db.Delete(&session)

	return doc, nil
}

// GetUploadProgress 获取上传进度
func (s *DocumentService) GetUploadProgress(sessionID string) (*models.UploadSession, error) {
	var session models.UploadSession
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}

	// 计算已上传的分片数量
	uploadedSize := int64(0)
	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk_%d", i))
		if info, err := os.Stat(chunkPath); err == nil {
			uploadedSize += info.Size()
		}
	}

	session.UploadedSize = uploadedSize
	s.db.Save(&session)

	return &session, nil
}

// Upload 传统上传方法（保持兼容性）
func (s *DocumentService) Upload(file *multipart.FileHeader) (*models.Document, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// 计算文件哈希
	hash := sha256.New()
	src.Seek(0, 0)
	io.Copy(hash, src)
	fileHash := fmt.Sprintf("%x", hash.Sum(nil))

	// 检查是否可以秒传
	if doc, exists := s.CheckFile(fileHash, file.Size); exists {
		return doc, nil
	}

	src.Seek(0, 0)
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	filePath := filepath.Join(s.uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, err
	}

	doc := &models.Document{
		Name:         strings.TrimSuffix(file.Filename, ext),
		OriginalName: file.Filename,
		FilePath:     filePath,
		FileSize:     file.Size,
		FileHash:     fileHash,
		MimeType:     file.Header.Get("Content-Type"),
		Extension:    ext,
		Status:       "completed",
	}

	if err := s.db.Create(doc).Error; err != nil {
		os.Remove(filePath)
		return nil, err
	}

	return doc, nil
}

func (s *DocumentService) List() ([]models.Document, error) {
	var docs []models.Document
	err := s.db.Find(&docs).Error
	return docs, err
}

func (s *DocumentService) GetByID(id uint) (*models.Document, error) {
	var doc models.Document
	err := s.db.First(&doc, id).Error
	return &doc, err
}

func (s *DocumentService) Delete(id uint) error {
	var doc models.Document
	if err := s.db.First(&doc, id).Error; err != nil {
		return err
	}

	if err := s.db.Delete(&doc).Error; err != nil {
		return err
	}

	os.Remove(doc.FilePath)
	return nil
}

func (s *DocumentService) UpdateDescription(id uint, description string) error {
	return s.db.Model(&models.Document{}).Where("id = ?", id).Update("description", description).Error
}
