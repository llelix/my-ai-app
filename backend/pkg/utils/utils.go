package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一API响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationRequest 分页请求结构
type PaginationRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Search   string `form:"search"`
	Sort     string `form:"sort"`
	Order    string `form:"order,default=desc" binding:"oneof=asc desc"`
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// ValidationError 验证错误响应
func ValidationError(c *gin.Context, errors interface{}) {
	c.JSON(422, Response{
		Code:    422,
		Message: "validation failed",
		Data:    errors,
	})
}

// GetOffset 计算分页偏移量
func GetOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// CalculateTotalPages 计算总页数
func CalculateTotalPages(total int64, pageSize int) int {
	if total <= 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

// GenerateID 生成随机ID
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// CleanText 清理文本
func CleanText(text string) string {
	// 移除多余的空白字符
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	return text
}

// TruncateText 截断文本
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

// ExtractKeywords 提取关键词
func ExtractKeywords(text string) []string {
	// 简单的关键词提取，可以根据需要改进
	words := strings.Fields(text)
	var keywords []string
	seen := make(map[string]bool)

	for _, word := range words {
		word = strings.ToLower(strings.Trim(word, ",.!?;:\"'()[]{}"))
		if len(word) > 2 && !seen[word] {
			keywords = append(keywords, word)
			seen[word] = true
		}
	}

	return keywords
}

// SaveUploadedFile 保存上传的文件
func SaveUploadedFile(file *multipart.FileHeader, dstDir string) (string, error) {
	// 确保目标目录存在
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return "", err
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), GenerateID()[:8], ext)
	dst := filepath.Join(dstDir, filename)

	// 保存文件
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, src); err != nil {
		return "", err
	}

	return filename, nil
}

// IsValidURL 验证URL格式
func IsValidURL(url string) bool {
	regex := regexp.MustCompile(`^(https?|ftp):\/\/[^\s/$.?#].[^\s]*$`)
	return regex.MatchString(url)
}

// EscapeHTML 转义HTML字符
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#x27;")
	return s
}

// ToJSON 转换为JSON字符串
func ToJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// FromJSON 从JSON字符串解析
func FromJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

// ContainsString 检查字符串切片是否包含指定字符串
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// RemoveDuplicateStrings 移除字符串切片中的重复项
func RemoveDuplicateStrings(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// TimeFormat 时间格式化常量
const (
	TimeFormatYYYYMMDD     = "2006-01-02"
	TimeFormatYYYYMMDDHHMM = "2006-01-02 15:04"
	TimeFormatYYYYMMDDHHMMSS = "2006-01-02 15:04:05"
	TimeFormatRFC3339      = time.RFC3339
)

// FormatTime 格式化时间
func FormatTime(t time.Time, layout string) string {
	return t.Format(layout)
}

// ParseTimeString 解析时间字符串
func ParseTimeString(s, layout string) (time.Time, error) {
	return time.Parse(layout, s)
}

// GetEnv 获取环境变量，支持默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevelopment 判断是否为开发环境
func IsDevelopment() bool {
	return gin.Mode() == gin.DebugMode
}

// GetClientIP 获取客户端IP
func GetClientIP(c *gin.Context) string {
	// 优先从X-Forwarded-For获取
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return ip
	}
	// 其次从X-Real-IP获取
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	// 最后使用RemoteAddr
	return c.ClientIP()
}