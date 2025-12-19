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
	Provider string       `mapstructure:"provider"`
	OpenAI   OpenAIConfig `mapstructure:"openai"`
	Claude   ClaudeConfig `mapstructure:"claude"`
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
	// 优先尝试加载YAML配置文件
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

	// 绑定环境变量
	bindEnvVars()

	// 环境变量自动覆盖
	viper.AutomaticEnv()

	// 尝试读取YAML配置文件
	err := viper.ReadInConfig()
	if err != nil {
		// 如果YAML文件不存在，尝试读取.env文件作为后备
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.SetConfigName(".env")
			viper.SetConfigType("env")
			if envErr := viper.ReadInConfig(); envErr != nil {
				// 如果两个文件都不存在，返回错误
				if _, ok := envErr.(viper.ConfigFileNotFoundError); ok {
					return nil, fmt.Errorf("no configuration file found (config.yml or .env)")
				}
				return nil, envErr
			}
		} else {
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
	// Server environment variable bindings
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.mode", "GIN_MODE")

	// Database environment variable bindings
	viper.BindEnv("database.type", "DB_TYPE")
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME")
	viper.BindEnv("database.path", "DB_PATH")

	// AI environment variable bindings
	viper.BindEnv("ai.provider", "AI_PROVIDER")
	viper.BindEnv("ai.openai.api_key", "OPENAI_API_KEY")
	viper.BindEnv("ai.openai.base_url", "OPENAI_BASE_URL")
	viper.BindEnv("ai.openai.model", "OPENAI_MODEL")
	viper.BindEnv("ai.claude.api_key", "CLAUDE_API_KEY")
	viper.BindEnv("ai.claude.base_url", "CLAUDE_BASE_URL")
	viper.BindEnv("ai.claude.model", "CLAUDE_MODEL")

	// Log environment variable bindings
	viper.BindEnv("log.level", "LOG_LEVEL")
	viper.BindEnv("log.format", "LOG_FORMAT")

	// CORS environment variable bindings
	viper.BindEnv("cors.allowed_origins", "CORS_ALLOWED_ORIGINS")
	viper.BindEnv("cors.allowed_methods", "CORS_ALLOWED_METHODS")
	viper.BindEnv("cors.allowed_headers", "CORS_ALLOWED_HEADERS")

	// S3 environment variable bindings
	viper.BindEnv("s3.endpoint", "S3_ENDPOINT")
	viper.BindEnv("s3.access_key_id", "S3_ACCESS_KEY_ID")
	viper.BindEnv("s3.secret_access_key", "S3_SECRET_ACCESS_KEY")
	viper.BindEnv("s3.use_ssl", "S3_USE_SSL")
	viper.BindEnv("s3.bucket", "S3_BUCKET")
	viper.BindEnv("s3.region", "S3_REGION")
}
