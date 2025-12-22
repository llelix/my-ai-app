package models

import (
	"time"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// Knowledge 知识条目模型
type Knowledge struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	Title         string          `json:"title" gorm:"not null;size:255"`
	Content       string          `json:"content" gorm:"type:text"`
	Summary       string          `json:"summary" gorm:"type:text"`
	IsPublished   bool            `json:"is_published" gorm:"default:true"`
	ViewCount     int             `json:"view_count" gorm:"default:0"`
	ContentVector pgvector.Vector `json:"-" gorm:"type:vector(1536)"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `json:"-" gorm:"index"`
}

// Tag 标签模型
type Tag struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	Name       string         `json:"name" gorm:"not null;size:50;unique"`
	Color      string         `json:"color" gorm:"size:7"`
	UsageCount int            `json:"usage_count" gorm:"default:0"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// QueryHistory AI查询历史模型
type QueryHistory struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Query        string         `json:"query" gorm:"not null;type:text"`
	Response     string         `json:"response" gorm:"type:text"`
	KnowledgeID  *uint          `json:"knowledge_id" gorm:"index"`
	Model        string         `json:"model" gorm:"size:50"`
	Tokens       int            `json:"tokens" gorm:"default:0"`
	Duration     int            `json:"duration" gorm:"default:0"` // 毫秒
	IsSuccess    bool           `json:"is_success" gorm:"default:true"`
	ErrorMessage string         `json:"error_message" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Knowledge *Knowledge `json:"knowledge,omitempty" gorm:"foreignKey:KnowledgeID"`
}

// KnowledgeTag 知识标签关联表
type KnowledgeTag struct {
	KnowledgeID uint      `json:"knowledge_id" gorm:"primaryKey"`
	TagID       uint      `json:"tag_id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 设置表名
func (Knowledge) TableName() string {
	return "knowledges"
}

func (Tag) TableName() string {
	return "tags"
}

func (QueryHistory) TableName() string {
	return "query_histories"
}

func (KnowledgeTag) TableName() string {
	return "knowledge_tags"
}
