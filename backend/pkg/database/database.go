package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"ai-knowledge-app/internal/config"
	"ai-knowledge-app/internal/models"

	"github.com/pgvector/pgvector-go"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.DatabaseConfig) error {
	var db *gorm.DB
	var err error

	// 配置GORM日志
	logLevel := logger.Silent
	switch cfg.Type {
	case "debug":
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	// 根据数据库类型建立连接
	switch cfg.Type {
	case "sqlite":
		db, err = initSQLiteDB(cfg, gormConfig)
	case "postgres":
		db, err = initPostgresDB(cfg, gormConfig)
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	log.Println("Database connected successfully")
	return nil
}

// initSQLiteDB 初始化SQLite数据库
func initSQLiteDB(cfg *config.DatabaseConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	// 确保数据库目录存在
	dbDir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// 连接SQLite数据库
	db, err := gorm.Open(sqlite.Open(cfg.Path), gormConfig)
	if err != nil {
		return nil, err
	}

	// SQLite特殊配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 启用外键约束
	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	// 设置WAL模式提高并发性能
	_, err = sqlDB.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		return nil, err
	}

	return db, nil
}

// initPostgresDB 初始化PostgreSQL数据库
func initPostgresDB(cfg *config.DatabaseConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port)

	var db *gorm.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err == nil {
			return db, nil
		}
		log.Printf("Failed to connect to database, retrying in 5 seconds... (%d/5)", i+1)
		time.Sleep(5 * time.Second)
	}

	return nil, err
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 定义需要迁移的模型
	models := []interface{}{
		&models.Category{},
		&models.Tag{},
		&models.Knowledge{},
		&models.KnowledgeTag{},
		&models.QueryHistory{},
		&models.Document{},
	}

	// 执行迁移
	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	log.Println("Database migration completed successfully")
	return nil
}

// SeedDatabase 初始化种子数据
func SeedDatabase() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 检查是否已有数据
	var count int64
	if err := DB.Model(&models.Category{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有数据，跳过种子数据创建
	if count > 0 {
		log.Println("Database already has data, skipping seed")
		return nil
	}

	// 创建默认分类
	categories := []models.Category{
		{
			Name:        "技术",
			Description: "编程、开发相关技术知识",
			Color:       "#1890ff",
			Icon:        "code",
			SortOrder:   1,
		},
		{
			Name:        "产品",
			Description: "产品设计、产品管理相关知识",
			Color:       "#52c41a",
			Icon:        "product",
			SortOrder:   2,
		},
		{
			Name:        "设计",
			Description: "UI/UX设计相关知识",
			Color:       "#fa8c16",
			Icon:        "design",
			SortOrder:   3,
		},
		{
			Name:        "其他",
			Description: "其他类型的知识",
			Color:       "#722ed1",
			Icon:        "more",
			SortOrder:   4,
		},
	}

	for _, category := range categories {
		if err := DB.Create(&category).Error; err != nil {
			return fmt.Errorf("failed to create category %s: %w", category.Name, err)
		}
	}

	// 创建默认标签
	tags := []models.Tag{
		{Name: "Go", Color: "#00a8e6"},
		{Name: "React", Color: "#61dafb"},
		{Name: "TypeScript", Color: "#3178c6"},
		{Name: "API设计", Color: "#ff6b6b"},
		{Name: "数据库", Color: "#4ecdc4"},
		{Name: "前端", Color: "#ff6b9d"},
		{Name: "后端", Color: "#c44569"},
		{Name: "算法", Color: "#f8b500"},
	}

	for _, tag := range tags {
		if err := DB.Create(&tag).Error; err != nil {
			return fmt.Errorf("failed to create tag %s: %w", tag.Name, err)
		}
	}

	// 创建示例知识条目
	sampleKnowledge := []models.Knowledge{
		{
			Title:         "Go语言基础",
			Content:       "Go是Google开发的编程语言，具有简洁的语法、高效的并发性能和丰富的标准库。",
			ContentVector: &pgvector.Vector{},
			Summary:       "Go语言是一种现代化的编程语言，特别适合构建高性能的网络服务。",
			CategoryID:    1, // 技术分类
			Metadata: models.Metadata{
				Author:     "浮浮酱",
				Source:     "官方文档",
				Language:   "zh",
				Difficulty: "easy",
				Keywords:   "Go,golang,编程语言",
			},
			IsPublished: true,
		},
		{
			Title:         "React Hooks简介",
			Content:       "React Hooks让你在不编写class的情况下使用state以及其他的React特性。",
			ContentVector: &pgvector.Vector{},
			Summary:       "Hooks是React 16.8推出的新特性，简化了组件逻辑的复用。",
			CategoryID:    1, // 技术分类
			Metadata: models.Metadata{
				Author:     "浮浮酱",
				Source:     "React文档",
				Language:   "zh",
				Difficulty: "medium",
				Keywords:   "React,Hooks,前端框架",
			},
			IsPublished: true,
		},
	}

	for _, knowledge := range sampleKnowledge {
		if err := DB.Create(&knowledge).Error; err != nil {
			return fmt.Errorf("failed to create knowledge %s: %w", knowledge.Title, err)
		}

		// 为示例知识添加标签
		var goTag, reactTag models.Tag
		DB.Where("name = ?", "Go").First(&goTag)
		DB.Where("name = ?", "React").First(&reactTag)

		if knowledge.Title == "Go语言基础" && goTag.ID != 0 {
			DB.Model(&knowledge).Association("Tags").Append(&goTag)
		}
		if knowledge.Title == "React Hooks简介" && reactTag.ID != 0 {
			DB.Model(&knowledge).Association("Tags").Append(&reactTag)
		}
	}

	log.Println("Seed data created successfully")
	return nil
}

// GetDatabase 获取数据库实例
func GetDatabase() *gorm.DB {
	return DB
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}