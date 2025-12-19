package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-knowledge-app/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type DocumentService struct {
	db          *gorm.DB
	uploadDir   string
	tempDir     string
	minioClient *MinIOClient
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

// SetMinIOClient sets the MinIO client for S3-compatible storage
func (s *DocumentService) SetMinIOClient(client *MinIOClient) {
	s.minioClient = client
}

// IsMinIOAvailable checks if MinIO service is available
func (s *DocumentService) IsMinIOAvailable() bool {
	if s.minioClient == nil {
		return false
	}
	return s.minioClient.IsServiceAvailable()
}

// CheckMinIOHealth performs a health check on MinIO service
func (s *DocumentService) CheckMinIOHealth() error {
	if s.minioClient == nil {
		return fmt.Errorf("MinIO client not configured")
	}
	return s.minioClient.IsHealthy()
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

// VerifyObjectIntegrity verifies that an object exists in storage and matches the expected hash
func (s *DocumentService) VerifyObjectIntegrity(filePath, expectedHash string) error {
	if s.minioClient != nil {
		// For MinIO, check if object exists and get its metadata
		ctx := context.Background()
		_, err := s.minioClient.StatObjectWithRetry(ctx, filePath, minio.StatObjectOptions{})
		if err != nil {
			return fmt.Errorf("object does not exist in MinIO: %w", err)
		}

		// Get the object to calculate its hash
		object, err := s.minioClient.GetObjectWithRetry(ctx, filePath, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to get object from MinIO: %w", err)
		}
		defer object.Close()

		// Calculate hash
		hash := sha256.New()
		if _, err := io.Copy(hash, object); err != nil {
			return fmt.Errorf("failed to calculate object hash: %w", err)
		}
		
		calculatedHash := fmt.Sprintf("%x", hash.Sum(nil))
		if calculatedHash != expectedHash {
			return fmt.Errorf("object hash mismatch: expected %s, got %s", expectedHash, calculatedHash)
		}

		return nil
	} else {
		// For local storage, check if file exists and verify hash
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("local file does not exist: %w", err)
		}
		defer file.Close()

		hash := sha256.New()
		if _, err := io.Copy(hash, file); err != nil {
			return fmt.Errorf("failed to calculate file hash: %w", err)
		}

		calculatedHash := fmt.Sprintf("%x", hash.Sum(nil))
		if calculatedHash != expectedHash {
			return fmt.Errorf("file hash mismatch: expected %s, got %s", expectedHash, calculatedHash)
		}

		return nil
	}
}

// CreateDuplicateReference creates a new document record that references an existing file
func (s *DocumentService) CreateDuplicateReference(originalDoc *models.Document, fileName, originalName string) (*models.Document, error) {
	// Verify that the original file still exists and has the correct hash
	if err := s.VerifyObjectIntegrity(originalDoc.FilePath, originalDoc.FileHash); err != nil {
		return nil, fmt.Errorf("original file integrity check failed: %w", err)
	}

	// Increment reference count of the original document
	if err := s.db.Model(originalDoc).UpdateColumn("ref_count", gorm.Expr("ref_count + ?", 1)).Error; err != nil {
		return nil, fmt.Errorf("failed to increment reference count: %w", err)
	}

	// Create new document record with same file path and hash
	ext := filepath.Ext(originalName)
	newDoc := &models.Document{
		Name:         strings.TrimSuffix(fileName, ext),
		OriginalName: originalName,
		FilePath:     originalDoc.FilePath, // Same S3 object key or local path
		FileSize:     originalDoc.FileSize,
		FileHash:     originalDoc.FileHash,
		MimeType:     originalDoc.MimeType,
		Extension:    ext,
		Status:       "completed",
		RefCount:     1, // This document also references the file
	}

	if err := s.db.Create(newDoc).Error; err != nil {
		// Rollback reference count increment on error
		s.db.Model(originalDoc).UpdateColumn("ref_count", gorm.Expr("ref_count - ?", 1))
		return nil, fmt.Errorf("failed to create duplicate reference: %w", err)
	}

	return newDoc, nil
}

// InitUpload 初始化上传会话
func (s *DocumentService) InitUpload(fileName string, fileSize int64, fileHash string) (*models.UploadSession, error) {
	// 检查是否可以秒传
	if doc, exists := s.CheckFile(fileHash, fileSize); exists {
		// Create a duplicate reference instead of returning an error
		duplicateDoc, err := s.CreateDuplicateReference(doc, fileName, fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to create duplicate reference: %w", err)
		}
		return nil, fmt.Errorf("file already exists, created duplicate reference: %d", duplicateDoc.ID)
	}

	chunkSize := int64(1048576) // 1MB
	totalChunks := int((fileSize + chunkSize - 1) / chunkSize)

	sessionID := uuid.New().String()
	tempDir := filepath.Join(s.tempDir, sessionID)
	var uploadID string
	
	if s.minioClient != nil {
		// For MinIO, use AWS S3 multipart upload
		objectKey := fmt.Sprintf("documents/%d_%s", time.Now().Unix(), fileName)
		tempDir = objectKey
		
		// Initialize S3 multipart upload
		ctx := context.Background()
		input := &s3.CreateMultipartUploadInput{
			Bucket: aws.String(s.minioClient.GetBucketName()),
			Key:    aws.String(objectKey),
		}
		
		result, err := s.minioClient.CreateMultipartUploadWithRetry(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize S3 multipart upload: %w", err)
		}
		uploadID = *result.UploadId
	} else {
		// Create temp directory for local storage
		os.MkdirAll(tempDir, 0755)
	}

	session := &models.UploadSession{
		ID:          sessionID,
		FileName:    fileName,
		FileSize:    fileSize,
		FileHash:    fileHash,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		TempDir:     tempDir,
		UploadID:    uploadID,
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
		// Clean up expired session
		if s.minioClient != nil && session.UploadID != "" {
			ctx := context.Background()
			input := &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(s.minioClient.GetBucketName()),
				Key:      aws.String(session.TempDir),
				UploadId: aws.String(session.UploadID),
			}
			s.minioClient.AbortMultipartUploadWithRetry(ctx, input)
		}
		s.db.Delete(&session)
		return fmt.Errorf("upload session expired")
	}

	if s.minioClient != nil {
		// For MinIO, use AWS S3 multipart upload part
		ctx := context.Background()
		reader := bytes.NewReader(data)
		
		// Part numbers in S3 start from 1, not 0
		partNumber := int32(chunkIndex + 1)
		
		input := &s3.UploadPartInput{
			Bucket:     aws.String(s.minioClient.GetBucketName()),
			Key:        aws.String(session.TempDir),
			UploadId:   aws.String(session.UploadID),
			PartNumber: &partNumber,
			Body:       reader,
		}
		
		_, err := s.minioClient.UploadPartWithRetry(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to upload chunk %d to S3: %w", chunkIndex, err)
		}
	} else {
		// Upload chunk to local storage
		chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk_%d", chunkIndex))
		if err := os.WriteFile(chunkPath, data, 0644); err != nil {
			return err
		}
	}
	
	return nil
}

// CompleteUpload 完成上传
func (s *DocumentService) CompleteUpload(sessionID string) (*models.Document, error) {
	var session models.UploadSession
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}

	ext := filepath.Ext(session.FileName)
	var finalPath string
	var calculatedHash string

	if s.minioClient != nil {
		// For MinIO: complete S3 multipart upload
		ctx := context.Background()
		finalPath = session.TempDir // This is the object key
		
		// First, list the uploaded parts to get their ETags
		listInput := &s3.ListPartsInput{
			Bucket:   aws.String(s.minioClient.GetBucketName()),
			Key:      aws.String(finalPath),
			UploadId: aws.String(session.UploadID),
		}
		
		listResult, err := s.minioClient.ListPartsWithRetry(ctx, listInput)
		if err != nil {
			// Abort the multipart upload on error
			abortInput := &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(s.minioClient.GetBucketName()),
				Key:      aws.String(finalPath),
				UploadId: aws.String(session.UploadID),
			}
			s.minioClient.AbortMultipartUploadWithRetry(ctx, abortInput)
			return nil, fmt.Errorf("failed to list parts for S3 multipart upload: %w", err)
		}
		
		// Build the list of completed parts with ETags
		var completedParts []types.CompletedPart
		for _, part := range listResult.Parts {
			completedParts = append(completedParts, types.CompletedPart{
				PartNumber: part.PartNumber,
				ETag:       part.ETag,
			})
		}
		
		// Complete the multipart upload
		completeInput := &s3.CompleteMultipartUploadInput{
			Bucket:   aws.String(s.minioClient.GetBucketName()),
			Key:      aws.String(finalPath),
			UploadId: aws.String(session.UploadID),
			MultipartUpload: &types.CompletedMultipartUpload{
				Parts: completedParts,
			},
		}
		
		_, err = s.minioClient.CompleteMultipartUploadWithRetry(ctx, completeInput)
		if err != nil {
			// Abort the multipart upload on error
			abortInput := &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(s.minioClient.GetBucketName()),
				Key:      aws.String(finalPath),
				UploadId: aws.String(session.UploadID),
			}
			s.minioClient.AbortMultipartUploadWithRetry(ctx, abortInput)
			return nil, fmt.Errorf("failed to complete S3 multipart upload: %w", err)
		}
		
		// For MinIO, we trust the hash provided during initialization
		calculatedHash = session.FileHash
	} else {
		// Local storage: merge chunks and verify hash
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), session.FileName)
		finalPath = filepath.Join(s.uploadDir, filename)

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
		calculatedHash = fmt.Sprintf("%x", hash.Sum(nil))

		if calculatedHash != session.FileHash {
			os.Remove(finalPath)
			return nil, fmt.Errorf("file hash mismatch")
		}
	}

	// 创建文档记录
	doc := &models.Document{
		Name:         strings.TrimSuffix(session.FileName, ext),
		OriginalName: session.FileName,
		FilePath:     finalPath,
		FileSize:     session.FileSize,
		FileHash:     calculatedHash,
		Extension:    ext,
		Status:       "completed",
	}

	if err := s.db.Create(doc).Error; err != nil {
		// Clean up on database error
		if s.minioClient != nil {
			ctx := context.Background()
			s.minioClient.RemoveObjectWithRetry(ctx, finalPath, minio.RemoveObjectOptions{})
		} else {
			os.Remove(finalPath)
		}
		return nil, err
	}

	// 清理临时文件和会话
	if s.minioClient == nil {
		os.RemoveAll(session.TempDir)
	}
	s.db.Delete(&session)

	return doc, nil
}

// GetUploadProgress 获取上传进度
func (s *DocumentService) GetUploadProgress(sessionID string) (*models.UploadSession, error) {
	var session models.UploadSession
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}

	uploadedSize := int64(0)
	
	if s.minioClient != nil {
		// For MinIO multipart upload, list uploaded parts using S3 API
		ctx := context.Background()
		
		if session.UploadID != "" {
			// List parts for the multipart upload
			input := &s3.ListPartsInput{
				Bucket:   aws.String(s.minioClient.GetBucketName()),
				Key:      aws.String(session.TempDir),
				UploadId: aws.String(session.UploadID),
			}
			
			result, err := s.minioClient.ListPartsWithRetry(ctx, input)
			if err != nil {
				// If we can't list parts, assume no progress
				uploadedSize = 0
			} else {
				// Sum up the sizes of uploaded parts
				for _, part := range result.Parts {
					if part.Size != nil {
						uploadedSize += *part.Size
					}
				}
			}
		}
	} else {
		// Local storage: calculate from chunk files
		for i := 0; i < session.TotalChunks; i++ {
			chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk_%d", i))
			if info, err := os.Stat(chunkPath); err == nil {
				uploadedSize += info.Size()
			}
		}
	}

	session.UploadedSize = uploadedSize
	s.db.Save(&session)

	return &session, nil
}

// AbortUpload 中止上传会话并清理资源
func (s *DocumentService) AbortUpload(sessionID string) error {
	var session models.UploadSession
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return err
	}

	if s.minioClient != nil {
		// Abort S3 multipart upload
		if session.UploadID != "" {
			ctx := context.Background()
			input := &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(s.minioClient.GetBucketName()),
				Key:      aws.String(session.TempDir),
				UploadId: aws.String(session.UploadID),
			}
			
			_, err := s.minioClient.AbortMultipartUploadWithRetry(ctx, input)
			if err != nil {
				// Log error but continue with cleanup
				fmt.Printf("Warning: failed to abort S3 multipart upload: %v\n", err)
			}
		}
	} else {
		// Clean up local temporary files
		if session.TempDir != "" {
			os.RemoveAll(session.TempDir)
		}
	}

	// Remove session from database
	return s.db.Delete(&session).Error
}

// CleanupExpiredSessions 清理过期的上传会话
func (s *DocumentService) CleanupExpiredSessions() error {
	var expiredSessions []models.UploadSession
	if err := s.db.Where("expires_at < ?", time.Now()).Find(&expiredSessions).Error; err != nil {
		return err
	}

	for _, session := range expiredSessions {
		if s.minioClient != nil {
			// Abort S3 multipart upload
			if session.UploadID != "" {
				ctx := context.Background()
				input := &s3.AbortMultipartUploadInput{
					Bucket:   aws.String(s.minioClient.GetBucketName()),
					Key:      aws.String(session.TempDir),
					UploadId: aws.String(session.UploadID),
				}
				
				_, err := s.minioClient.AbortMultipartUploadWithRetry(ctx, input)
				if err != nil {
					// Log error but continue with cleanup
					fmt.Printf("Warning: failed to abort expired S3 multipart upload %s: %v\n", session.ID, err)
				}
			}
		} else {
			// Clean up local temporary files
			if session.TempDir != "" {
				os.RemoveAll(session.TempDir)
			}
		}
	}

	// Remove expired sessions from database
	return s.db.Where("expires_at < ?", time.Now()).Delete(&models.UploadSession{}).Error
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
		// Create a duplicate reference instead of returning the original
		return s.CreateDuplicateReference(doc, file.Filename, file.Filename)
	}

	src.Seek(0, 0)
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	
	var filePath string
	
	// Use MinIO if available, otherwise fallback to local storage
	if s.minioClient != nil {
		// Generate S3 object key
		objectKey := fmt.Sprintf("documents/%s", filename)
		
		// Upload to MinIO with retry logic
		ctx := context.Background()
		_, err = s.minioClient.PutObjectWithRetry(ctx, objectKey, src, file.Size, minio.PutObjectOptions{
			ContentType: file.Header.Get("Content-Type"),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upload to MinIO: %w", err)
		}
		
		filePath = objectKey // Store S3 object key as file path
	} else {
		// Fallback to local storage
		filePath = filepath.Join(s.uploadDir, filename)
		dst, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			return nil, err
		}
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
		// Clean up uploaded file on database error
		if s.minioClient != nil {
			ctx := context.Background()
			s.minioClient.RemoveObjectWithRetry(ctx, filePath, minio.RemoveObjectOptions{})
		} else {
			os.Remove(filePath)
		}
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

// GetObject retrieves a file from storage (MinIO or local)
func (s *DocumentService) GetObject(filePath string) (io.ReadCloser, error) {
	if s.minioClient != nil {
		// Get object from MinIO
		ctx := context.Background()
		object, err := s.minioClient.GetObjectWithRetry(ctx, filePath, minio.GetObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get object from MinIO: %w", err)
		}
		return object, nil
	} else {
		// Get file from local storage
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open local file: %w", err)
		}
		return file, nil
	}
}

func (s *DocumentService) Delete(id uint) error {
	var doc models.Document
	if err := s.db.First(&doc, id).Error; err != nil {
		return err
	}

	// Start a transaction to ensure consistency
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete the document record
	if err := tx.Delete(&doc).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Check if there are other documents referencing the same file
	var remainingRefs int64
	if err := tx.Model(&models.Document{}).Where("file_hash = ? AND file_size = ? AND status = ?", 
		doc.FileHash, doc.FileSize, "completed").Count(&remainingRefs).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to count remaining references: %w", err)
	}

	// Only remove the physical file if no other documents reference it
	if remainingRefs == 0 {
		if s.minioClient != nil {
			// Remove object from MinIO
			ctx := context.Background()
			err := s.minioClient.RemoveObjectWithRetry(ctx, doc.FilePath, minio.RemoveObjectOptions{})
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to remove object from MinIO: %w", err)
			}
		} else {
			// Remove file from local storage
			if err := os.Remove(doc.FilePath); err != nil && !os.IsNotExist(err) {
				tx.Rollback()
				return fmt.Errorf("failed to remove local file: %w", err)
			}
		}
	}

	return tx.Commit().Error
}

func (s *DocumentService) UpdateDescription(id uint, description string) error {
	return s.db.Model(&models.Document{}).Where("id = ?", id).Update("description", description).Error
}

// CleanupOrphanedObjects removes objects from storage that have no database references
func (s *DocumentService) CleanupOrphanedObjects() error {
	if s.minioClient == nil {
		// For local storage, this is more complex and not implemented in this basic version
		return nil
	}

	ctx := context.Background()
	
	// List all objects in the bucket
	objectCh := s.minioClient.ListObjectsWithRetry(ctx, minio.ListObjectsOptions{
		Prefix:    "documents/",
		Recursive: true,
	})

	var orphanedObjects []string
	
	for object := range objectCh {
		if object.Err != nil {
			return fmt.Errorf("error listing objects: %w", object.Err)
		}

		// Check if any document references this object
		var count int64
		if err := s.db.Model(&models.Document{}).Where("file_path = ? AND status = ?", object.Key, "completed").Count(&count).Error; err != nil {
			return fmt.Errorf("error checking object references: %w", err)
		}

		if count == 0 {
			orphanedObjects = append(orphanedObjects, object.Key)
		}
	}

	// Remove orphaned objects
	for _, objectKey := range orphanedObjects {
		if err := s.minioClient.RemoveObjectWithRetry(ctx, objectKey, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("failed to remove orphaned object %s: %w", objectKey, err)
		}
	}

	return nil
}

// GetDeduplicationStats returns statistics about file deduplication
func (s *DocumentService) GetDeduplicationStats() (map[string]interface{}, error) {
	var totalDocs int64
	var uniqueFiles int64
	var totalSize int64
	var uniqueSize int64

	// Count total documents
	if err := s.db.Model(&models.Document{}).Where("status = ?", "completed").Count(&totalDocs).Error; err != nil {
		return nil, fmt.Errorf("failed to count total documents: %w", err)
	}

	// Count unique files (by hash and size)
	if err := s.db.Model(&models.Document{}).
		Select("COUNT(DISTINCT (file_hash || ':' || file_size))").
		Where("status = ?", "completed").
		Scan(&uniqueFiles).Error; err != nil {
		return nil, fmt.Errorf("failed to count unique files: %w", err)
	}

	// Calculate total size
	if err := s.db.Model(&models.Document{}).
		Select("SUM(file_size)").
		Where("status = ?", "completed").
		Scan(&totalSize).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total size: %w", err)
	}

	// Calculate unique size (sum of distinct file sizes by hash)
	if err := s.db.Raw(`
		SELECT SUM(file_size) FROM (
			SELECT DISTINCT file_hash, file_size 
			FROM documents 
			WHERE status = ?
		) AS unique_files
	`, "completed").Scan(&uniqueSize).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate unique size: %w", err)
	}

	spaceSaved := totalSize - uniqueSize
	deduplicationRatio := float64(0)
	if totalSize > 0 {
		deduplicationRatio = float64(spaceSaved) / float64(totalSize) * 100
	}

	return map[string]interface{}{
		"total_documents":      totalDocs,
		"unique_files":         uniqueFiles,
		"total_size_bytes":     totalSize,
		"unique_size_bytes":    uniqueSize,
		"space_saved_bytes":    spaceSaved,
		"deduplication_ratio":  deduplicationRatio,
	}, nil
}
