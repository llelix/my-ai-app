package logger

import (
	"os"
	"path/filepath"

	"ai-knowledge-app/internal/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志实例
var Logger *logrus.Logger

// InitLogger 初始化日志系统
func InitLogger(cfg *config.LogConfig) error {
	Logger = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// 设置日志格式
	switch cfg.Format {
	case "json":
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 设置输出
	Logger.SetOutput(os.Stdout)

	// 创建日志目录
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 添加文件日志钩子
	fileHook := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "app.log"),
		MaxSize:    100, // MB
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	}

	Logger.AddHook(&FileHook{
		Logger: fileHook,
		Level:  level,
	})

	Logger.Info("Logger initialized successfully")
	return nil
}

// FileHook 文件日志钩子
type FileHook struct {
	Logger *lumberjack.Logger
	Level  logrus.Level
}

// Levels 实现logrus.Hook接口
func (hook *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels[:hook.Level+1]
}

// Fire 实现logrus.Hook接口
func (hook *FileHook) Fire(entry *logrus.Entry) error {
	line, err := entry.Logger.Formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = hook.Logger.Write(line)
	return err
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	return Logger
}

// WithRequestID 为日志添加请求ID
func WithRequestID(requestID string) *logrus.Entry {
	return Logger.WithField("request_id", requestID)
}

// WithUserID 为日志添加用户ID
func WithUserID(userID uint) *logrus.Entry {
	return Logger.WithField("user_id", userID)
}

// WithError 为日志添加错误信息
func WithError(err error) *logrus.Entry {
	return Logger.WithError(err)
}

// WithFields 为日志添加多个字段
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}