package service

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"gorm.io/gorm"
	"ai-knowledge-app/internal/models"
)

type DocumentProcessor struct {
	db *gorm.DB
}

func NewDocumentProcessor(db *gorm.DB) *DocumentProcessor {
	return &DocumentProcessor{db: db}
}

func (dp *DocumentProcessor) CreateDocument(doc *models.Document) error {
	return dp.db.Create(doc).Error
}

func (dp *DocumentProcessor) GetDocument(id uint) (*models.Document, error) {
	var doc models.Document
	err := dp.db.First(&doc, id).Error
	return &doc, err
}

func (dp *DocumentProcessor) GetDocumentChunks(docID uint) ([]models.DocumentChunk, error) {
	var chunks []models.DocumentChunk
	err := dp.db.Where("document_id = ?", docID).Find(&chunks).Error
	return chunks, err
}

func (dp *DocumentProcessor) ProcessDocument(docID uint) error {
	var doc models.Document
	if err := dp.db.First(&doc, docID).Error; err != nil {
		return err
	}

	if err := dp.parseDocument(&doc); err != nil {
		doc.Status = "failed"
		doc.Error = err.Error()
		dp.db.Save(&doc)
		return err
	}

	if err := dp.cleanText(&doc); err != nil {
		doc.Status = "failed"
		doc.Error = err.Error()
		dp.db.Save(&doc)
		return err
	}

	if err := dp.chunkText(&doc); err != nil {
		doc.Status = "failed"
		doc.Error = err.Error()
		dp.db.Save(&doc)
		return err
	}

	doc.Status = "completed"
	return dp.db.Save(&doc).Error
}

func (dp *DocumentProcessor) parseDocument(doc *models.Document) error {
	doc.Status = "parsing"
	dp.db.Save(doc)

	content, err := os.ReadFile(doc.FilePath)
	if err != nil {
		return err
	}

	switch strings.ToLower(doc.FileType) {
	case "txt", "html":
		doc.RawText = string(content)
	default:
		return fmt.Errorf("unsupported file type: %s", doc.FileType)
	}

	return dp.db.Save(doc).Error
}

func (dp *DocumentProcessor) cleanText(doc *models.Document) error {
	doc.Status = "cleaning"
	dp.db.Save(doc)

	text := doc.RawText
	
	// 去除HTML标签
	text = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(text, "")
	// 去除页眉页脚
	text = regexp.MustCompile(`(?i)(第\s*\d+\s*页|page\s*\d+)`).ReplaceAllString(text, "")
	// 去除多余空白
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	// 去除特殊符号
	text = regexp.MustCompile(`[^\w\s\u4e00-\u9fff.,!?;:()""''【】（）。，！？；：]`).ReplaceAllString(text, "")
	
	doc.CleanedText = strings.TrimSpace(text)
	return dp.db.Save(doc).Error
}

func (dp *DocumentProcessor) chunkText(doc *models.Document) error {
	doc.Status = "chunking"
	dp.db.Save(doc)

	text := doc.CleanedText
	chunkSize := 500
	overlap := 50

	var chunks []models.DocumentChunk
	for i := 0; i < len(text); i += chunkSize - overlap {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}
		
		chunks = append(chunks, models.DocumentChunk{
			DocumentID: doc.ID,
			ChunkIndex: len(chunks),
			Content:    text[i:end],
		})
		
		if end == len(text) {
			break
		}
	}

	if err := dp.db.Create(&chunks).Error; err != nil {
		return err
	}

	doc.ChunkCount = len(chunks)
	return dp.db.Save(doc).Error
}
