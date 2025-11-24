package config

import (
	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	AI       AIConfig       `mapstructure:"ai"`
	Log      LogConfig      `mapstructure:"log"`
	CORS     CORSConfig     `mapstructure:"cors"`
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

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

	// 设置默认值
	setDefaults()

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

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")

	// Database defaults
	viper.SetDefault("database.type", "sqlite")
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
}