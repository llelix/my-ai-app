package models

import "time"

type ProcessingStatus string

const (
	StatusParsing   ProcessingStatus = "parsing"
	StatusCleaning  ProcessingStatus = "cleaning"
	StatusChunking  ProcessingStatus = "chunking"
	StatusCompleted ProcessingStatus = "completed"
	StatusFailed    ProcessingStatus = "failed"
)

type Document struct {
	ID           uint             `json:"id" gorm:"primaryKey"`
	Name         string           `json:"name"`
	OriginalName string           `json:"original_name"`
	FileName     string           `json:"file_name"`
	FileType     string           `json:"file_type"`
	FilePath     string           `json:"file_path"` // Stores S3 object key for S3-compatible storage
	FileSize     int64            `json:"file_size"`
	FileHash     string           `json:"file_hash"`
	MimeType     string           `json:"mime_type"`
	Extension    string           `json:"extension"`
	Description  string           `json:"description"`
	Status       string           `json:"status" gorm:"default:'completed'"`
	RawText      string           `json:"raw_text" gorm:"type:text"`
	CleanedText  string           `json:"cleaned_text" gorm:"type:text"`
	ChunkCount   int              `json:"chunk_count"`
	Error        string           `json:"error,omitempty"`
	
	// Reference counting for deduplication
	RefCount     int              `json:"ref_count" gorm:"default:1"`
	
	// Relationships
	Chunks       []DocumentChunk  `json:"chunks,omitempty" gorm:"foreignKey:DocumentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type DocumentChunk struct {
	ID         uint     `json:"id" gorm:"primaryKey"`
	DocumentID uint     `json:"document_id" gorm:"not null;index"`
	Document   Document `json:"document" gorm:"foreignKey:DocumentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ChunkIndex int      `json:"chunk_index"`
	Content    string   `json:"content" gorm:"type:text"`
}

type UploadSession struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	FileHash     string    `json:"file_hash"`
	ChunkSize    int64     `json:"chunk_size"`
	TotalChunks  int       `json:"total_chunks"`
	UploadedSize int64     `json:"uploaded_size"`
	TempDir      string    `json:"temp_dir"`
	
	// MinIO multipart upload ID for S3-compatible storage
	UploadID     string    `json:"upload_id" gorm:"column:upload_id"`
	
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
