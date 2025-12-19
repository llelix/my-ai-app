package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	AI       AIConfig       `mapstructure:"ai"`
	Log      LogConfig      `mapstructure:"log"`
	CORS     CORSConfig     `mapstructure:"cors"`
	S3       S3Config       `mapstructure:"s3"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	Path     string `mapstructure:"path"`
}

// AIConfig AI服务配置
type AIConfig struct {
	Provider  string `mapstructure:"provider"`
	OpenAI    OpenAIConfig `mapstructure:"openai"`
	Claude    ClaudeConfig `mapstructure:"claude"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

// ClaudeConfig Claude配置
type ClaudeConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

// S3Config S3兼容对象存储配置
type S3Config struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	Bucket          string `mapstructure:"bucket"`
	Region          string `mapstructure:"region"`
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证S3配置
	if err := c.S3.Validate(); err != nil {
		return fmt.Errorf("S3 configuration error: %w", err)
	}
	return nil
}

// Validate 验证S3配置
func (s *S3Config) Validate() error {
	if s.Endpoint == "" {
		return fmt.Errorf("S3 endpoint is required")
	}
	if s.AccessKeyID == "" {
		return fmt.Errorf("S3 access key ID is required")
	}
	if s.SecretAccessKey == "" {
		return fmt.Errorf("S3 secret access key is required")
	}
	if s.Bucket == "" {
		return fmt.Errorf("S3 bucket name is required")
	}
	if s.Region == "" {
		return fmt.Errorf("S3 region is required")
	}
	return nil
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

	// 设置默认值
	setDefaults()

	// 绑定环境变量
	bindEnvVars()

	// 环境变量自动覆盖
	viper.AutomaticEnv()

	// 读取环境变量文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果.env文件不存在，使用默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// bindEnvVars 绑定环境变量到配置键
func bindEnvVars() {
	// S3 environment variable bindings
	viper.BindEnv("s3.endpoint", "S3_ENDPOINT")
	viper.BindEnv("s3.access_key_id", "S3_ACCESS_KEY_ID")
	viper.BindEnv("s3.secret_access_key", "S3_SECRET_ACCESS_KEY")
	viper.BindEnv("s3.use_ssl", "S3_USE_SSL")
	viper.BindEnv("s3.bucket", "S3_BUCKET")
	viper.BindEnv("s3.region", "S3_REGION")
}

// setDefaults 设置默认配置值
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")

	// Database defaults
	viper.SetDefault("database.type", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "ai_knowledge_pw")
	viper.SetDefault("database.dbname", "ai_knowledge_db")
	viper.SetDefault("database.path", "./data/app.db")

	// AI defaults
	viper.SetDefault("ai.provider", "openai")
	viper.SetDefault("ai.openai.base_url", "https://api.openai.com/v1")
	viper.SetDefault("ai.openai.model", "gpt-3.5-turbo")

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:5173"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Content-Type", "Authorization"})

	// S3 defaults (for MinIO local development)
	viper.SetDefault("s3.endpoint", "localhost:9000")
	viper.SetDefault("s3.access_key_id", "minioadmin")
	viper.SetDefault("s3.secret_access_key", "minioadmin123")
	viper.SetDefault("s3.use_ssl", false)
	viper.SetDefault("s3.bucket", "ai-knowledge-files")
	viper.SetDefault("s3.region", "us-east-1")
}