package service

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"mime/multipart"
	"testing"

	"ai-knowledge-app/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto migrate the schema
	db.AutoMigrate(&models.Document{}, &models.UploadSession{})
	return db
}

func createTestFileHeader(filename, content string) *multipart.FileHeader {
	// Create a buffer to write our multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Create a form file field
	fw, _ := w.CreateFormFile("file", filename)
	fw.Write([]byte(content))
	w.Close()

	// Parse the multipart form
	r := multipart.NewReader(&b, w.Boundary())
	form, _ := r.ReadForm(32 << 20) // 32MB max memory

	return form.File["file"][0]
}

func TestFileDeduplication(t *testing.T) {
	db := setupTestDB()
	service := NewDocumentService(db)

	// Test content
	content := "This is test content for deduplication"
	hash := sha256.New()
	hash.Write([]byte(content))
	expectedHash := fmt.Sprintf("%x", hash.Sum(nil))

	// Create first file
	file1 := createTestFileHeader("test1.txt", content)
	doc1, err := service.Upload(file1)
	if err != nil {
		t.Fatalf("Failed to upload first file: %v", err)
	}

	// Verify first document was created
	if doc1.FileHash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, doc1.FileHash)
	}
	if doc1.RefCount != 1 {
		t.Errorf("Expected ref_count 1, got %d", doc1.RefCount)
	}

	// Create second file with same content but different name
	file2 := createTestFileHeader("test2.txt", content)
	doc2, err := service.Upload(file2)
	if err != nil {
		t.Fatalf("Failed to upload second file: %v", err)
	}

	// Verify second document references the same file
	if doc2.FileHash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, doc2.FileHash)
	}
	if doc2.FilePath != doc1.FilePath {
		t.Errorf("Expected same file path, got different paths: %s vs %s", doc1.FilePath, doc2.FilePath)
	}
	if doc2.RefCount != 1 {
		t.Errorf("Expected ref_count 1, got %d", doc2.RefCount)
	}

	// Verify original document's ref_count was incremented
	var updatedDoc1 models.Document
	db.First(&updatedDoc1, doc1.ID)
	if updatedDoc1.RefCount != 2 {
		t.Errorf("Expected original document ref_count 2, got %d", updatedDoc1.RefCount)
	}

	// Test deduplication stats
	stats, err := service.GetDeduplicationStats()
	if err != nil {
		t.Fatalf("Failed to get deduplication stats: %v", err)
	}

	totalDocs := stats["total_documents"].(int64)
	uniqueFiles := stats["unique_files"].(int64)

	if totalDocs != 2 {
		t.Errorf("Expected 2 total documents, got %d", totalDocs)
	}
	if uniqueFiles != 1 {
		t.Errorf("Expected 1 unique file, got %d", uniqueFiles)
	}
}

func TestReferenceCountedDeletion(t *testing.T) {
	db := setupTestDB()
	service := NewDocumentService(db)

	// Test content
	content := "This is test content for deletion"
	
	// Create first file
	file1 := createTestFileHeader("delete1.txt", content)
	doc1, err := service.Upload(file1)
	if err != nil {
		t.Fatalf("Failed to upload first file: %v", err)
	}

	// Create second file with same content
	file2 := createTestFileHeader("delete2.txt", content)
	doc2, err := service.Upload(file2)
	if err != nil {
		t.Fatalf("Failed to upload second file: %v", err)
	}

	// Verify both documents exist
	var count int64
	db.Model(&models.Document{}).Where("file_hash = ?", doc1.FileHash).Count(&count)
	if count != 2 {
		t.Errorf("Expected 2 documents with same hash, got %d", count)
	}

	// Delete first document
	err = service.Delete(doc1.ID)
	if err != nil {
		t.Fatalf("Failed to delete first document: %v", err)
	}

	// Verify first document is deleted but second still exists
	db.Model(&models.Document{}).Where("file_hash = ?", doc1.FileHash).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 document remaining after first deletion, got %d", count)
	}

	// Verify second document still exists
	var remainingDoc models.Document
	err = db.First(&remainingDoc, doc2.ID).Error
	if err != nil {
		t.Errorf("Second document should still exist: %v", err)
	}

	// Delete second document
	err = service.Delete(doc2.ID)
	if err != nil {
		t.Fatalf("Failed to delete second document: %v", err)
	}

	// Verify no documents with that hash remain
	db.Model(&models.Document{}).Where("file_hash = ?", doc1.FileHash).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 documents after all deletions, got %d", count)
	}
}

func TestCheckFileDeduplication(t *testing.T) {
	db := setupTestDB()
	service := NewDocumentService(db)

	// Test content
	content := "This is test content for check file"
	hash := sha256.New()
	hash.Write([]byte(content))
	expectedHash := fmt.Sprintf("%x", hash.Sum(nil))
	size := int64(len(content))

	// Initially, file should not exist
	doc, exists := service.CheckFile(expectedHash, size)
	if exists {
		t.Error("File should not exist initially")
	}
	if doc != nil {
		t.Error("Document should be nil when file doesn't exist")
	}

	// Create a file
	file := createTestFileHeader("check.txt", content)
	createdDoc, err := service.Upload(file)
	if err != nil {
		t.Fatalf("Failed to upload file: %v", err)
	}

	// Now file should exist
	doc, exists = service.CheckFile(expectedHash, size)
	if !exists {
		t.Error("File should exist after upload")
	}
	if doc == nil {
		t.Error("Document should not be nil when file exists")
	}
	if doc.ID != createdDoc.ID {
		t.Errorf("Expected document ID %d, got %d", createdDoc.ID, doc.ID)
	}
}