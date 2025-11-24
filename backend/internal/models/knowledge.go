package models

import (
	"time"
	"gorm.io/gorm"
)

// Knowledge 知识条目模型
type Knowledge struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"not null;size:255;index"`
	Content     string         `json:"content" gorm:"type:text"`
	Summary     string         `json:"summary" gorm:"type:text"`
	CategoryID  uint           `json:"category_id" gorm:"index"`
	Tags        []Tag          `json:"tags" gorm:"many2many:knowledge_tags;"`
	Metadata    Metadata       `json:"metadata" gorm:"embedded"`
	IsPublished bool           `json:"is_published" gorm:"default:true"`
	ViewCount   int            `json:"view_count" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Category    *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	QueryHistory []QueryHistory `json:"query_history,omitempty" gorm:"foreignKey:KnowledgeID"`
}

// Category 知识分类模型
type Category struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:100;uniqueIndex"`
	Description string         `json:"description" gorm:"type:text"`
	Color       string         `json:"color" gorm:"size:7"` // 十六进制颜色代码
	Icon        string         `json:"icon" gorm:"size:50"`
	ParentID    *uint          `json:"parent_id" gorm:"index"`
	SortOrder   int            `json:"sort_order" gorm:"default:0"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Parent   *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Knowledges []Knowledge `json:"knowledges,omitempty" gorm:"foreignKey:CategoryID"`
}

// Tag 标签模型
type Tag struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null;size:50;uniqueIndex"`
	Color     string         `json:"color" gorm:"size:7"`
	UsageCount int           `json:"usage_count" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Knowledges []Knowledge `json:"knowledges,omitempty" gorm:"many2many:knowledge_tags;"`
}

// Metadata 元数据嵌入
type Metadata struct {
	Author     string `json:"author" gorm:"size:100"`
	Source     string `json:"source" gorm:"size:255"`
	Language   string `json:"language" gorm:"default:'zh';size:10"`
	Difficulty string `json:"difficulty" gorm:"size:20"` // easy, medium, hard
	Keywords   string `json:"keywords" gorm:"type:text"`
	WordCount  int    `json:"word_count" gorm:"default:0"`
}

// QueryHistory AI查询历史模型
type QueryHistory struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Query       string         `json:"query" gorm:"not null;type:text"`
	Response    string         `json:"response" gorm:"type:text"`
	KnowledgeID *uint          `json:"knowledge_id" gorm:"index"`
	Model       string         `json:"model" gorm:"size:50"`
	Tokens      int            `json:"tokens" gorm:"default:0"`
	Duration    int            `json:"duration" gorm:"default:0"` // 毫秒
	IsSuccess   bool           `json:"is_success" gorm:"default:true"`
	ErrorMessage string        `json:"error_message" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Knowledge *Knowledge `json:"knowledge,omitempty" gorm:"foreignKey:KnowledgeID"`
}

// KnowledgeTag 知识标签关联表
type KnowledgeTag struct {
	KnowledgeID uint `json:"knowledge_id" gorm:"primaryKey"`
	TagID       uint `json:"tag_id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 设置表名
func (Knowledge) TableName() string {
	return "knowledges"
}

func (Category) TableName() string {
	return "categories"
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

// BeforeCreate GORM钩子：创建前
func (k *Knowledge) BeforeCreate(tx *gorm.DB) error {
	if k.Metadata.WordCount == 0 && k.Content != "" {
		// 简单的字数统计（可以根据需要优化）
		k.Metadata.WordCount = len([]rune(k.Content))
	}
	return nil
}

// BeforeUpdate GORM钩子：更新前
func (k *Knowledge) BeforeUpdate(tx *gorm.DB) error {
	if k.Content != "" {
		k.Metadata.WordCount = len([]rune(k.Content))
	}
	return nil
}

// BeforeCreate GORM钩子：标签创建前
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	t.UsageCount = 0
	return nil
}