package preprocessing

import (
	"ai-knowledge-app/internal/preprocessing/core"
)

// Config 预处理模块配置
type Config struct {
	// MinerU配置
	MinerU MinerUConfig `json:"mineru" yaml:"mineru"`

	// 文本分块配置
	TextChunking ChunkingConfig `json:"text_chunking" yaml:"text_chunking"`

	// 队列配置
	Queue QueueConfig `json:"queue" yaml:"queue"`

	// 向量化配置（预留）
	Vectorization core.VectorizationConfig `json:"vectorization" yaml:"vectorization"`

	// 质量验证配置
	Quality QualityConfig `json:"quality" yaml:"quality"`
}

// MinerUConfig MinerU处理器配置
type MinerUConfig struct {
	// 默认转换选项
	DefaultOptions core.ConversionOptions `json:"default_options" yaml:"default_options"`

	// 超时设置
	TimeoutSeconds int `json:"timeout_seconds" yaml:"timeout_seconds"`

	// 临时目录
	TempDir string `json:"temp_dir" yaml:"temp_dir"`

	// 最大文件大小（字节）
	MaxFileSize int64 `json:"max_file_size" yaml:"max_file_size"`

	// 支持的格式
	SupportedFormats []string `json:"supported_formats" yaml:"supported_formats"`
}

// ChunkingConfig 文本分块配置
type ChunkingConfig struct {
	// 默认分块选项
	DefaultOptions core.ChunkingOptions `json:"default_options" yaml:"default_options"`

	// 最大块数量
	MaxChunks int `json:"max_chunks" yaml:"max_chunks"`

	// 最小块大小
	MinChunkSize int `json:"min_chunk_size" yaml:"min_chunk_size"`

	// 最大块大小
	MaxChunkSize int `json:"max_chunk_size" yaml:"max_chunk_size"`
}

// QueueConfig 队列配置
type QueueConfig struct {
	// 工作协程数量
	WorkerCount int `json:"worker_count" yaml:"worker_count"`

	// 队列容量
	QueueSize int `json:"queue_size" yaml:"queue_size"`

	// 最大重试次数
	MaxRetries int `json:"max_retries" yaml:"max_retries"`

	// 重试延迟（秒）
	RetryDelaySeconds int `json:"retry_delay_seconds" yaml:"retry_delay_seconds"`

	// 任务超时（秒）
	TaskTimeoutSeconds int `json:"task_timeout_seconds" yaml:"task_timeout_seconds"`
}

// QualityConfig 质量验证配置
type QualityConfig struct {
	// 是否启用质量验证
	Enabled bool `json:"enabled" yaml:"enabled"`

	// 最小质量分数
	MinQualityScore float64 `json:"min_quality_score" yaml:"min_quality_score"`

	// 是否严格模式
	StrictMode bool `json:"strict_mode" yaml:"strict_mode"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MinerU: MinerUConfig{
			DefaultOptions: core.ConversionOptions{
				Language:      "zh",
				Backend:       "pipeline",
				ParseMethod:   "auto",
				FormulaEnable: true,
				TableEnable:   true,
				ExtractImages: false,
			},
			TimeoutSeconds:   300, // 5分钟
			TempDir:          "/tmp/preprocessing",
			MaxFileSize:      100 * 1024 * 1024, // 100MB
			SupportedFormats: []string{"pdf", "docx", "doc", "txt", "md"},
		},
		TextChunking: ChunkingConfig{
			DefaultOptions: core.ChunkingOptions{
				ChunkSize:    1000,
				ChunkOverlap: 200,
				Separators:   []string{"\n\n", "\n", ".", "!", "?"},
			},
			MaxChunks:    1000,
			MinChunkSize: 100,
			MaxChunkSize: 4000,
		},
		Queue: QueueConfig{
			WorkerCount:        3,
			QueueSize:          100,
			MaxRetries:         3,
			RetryDelaySeconds:  30,
			TaskTimeoutSeconds: 600, // 10分钟
		},
		Vectorization: core.VectorizationConfig{
			Enabled:    false, // 预留功能，默认关闭
			Model:      "text-embedding-ada-002",
			BatchSize:  100,
			Dimensions: 1536,
		},
		Quality: QualityConfig{
			Enabled:         true,
			MinQualityScore: 0.7,
			StrictMode:      false,
		},
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Queue.WorkerCount <= 0 {
		return core.NewValidationError("queue.worker_count", c.Queue.WorkerCount, "worker count must be positive")
	}

	if c.Queue.QueueSize <= 0 {
		return core.NewValidationError("queue.queue_size", c.Queue.QueueSize, "queue size must be positive")
	}

	if c.TextChunking.DefaultOptions.ChunkSize <= 0 {
		return core.NewValidationError("text_chunking.chunk_size", c.TextChunking.DefaultOptions.ChunkSize, "chunk size must be positive")
	}

	if c.TextChunking.DefaultOptions.ChunkOverlap < 0 {
		return core.NewValidationError("text_chunking.chunk_overlap", c.TextChunking.DefaultOptions.ChunkOverlap, "chunk overlap must be non-negative")
	}

	if c.TextChunking.DefaultOptions.ChunkOverlap >= c.TextChunking.DefaultOptions.ChunkSize {
		return core.NewValidationError("text_chunking.chunk_overlap", c.TextChunking.DefaultOptions.ChunkOverlap, "chunk overlap must be less than chunk size")
	}

	if c.MinerU.MaxFileSize <= 0 {
		return core.NewValidationError("mineru.max_file_size", c.MinerU.MaxFileSize, "max file size must be positive")
	}

	if c.Quality.Enabled && (c.Quality.MinQualityScore < 0 || c.Quality.MinQualityScore > 1) {
		return core.NewValidationError("quality.min_quality_score", c.Quality.MinQualityScore, "quality score must be between 0 and 1")
	}

	return nil
}
