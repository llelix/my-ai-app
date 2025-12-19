package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"ai-knowledge-app/internal/config"
	"ai-knowledge-app/internal/models"
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
		&models.DocumentChunk{},
		&models.UploadSession{},
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