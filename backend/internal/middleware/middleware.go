package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"ai-knowledge-app/pkg/logger"
	"ai-knowledge-app/pkg/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = utils.GenerateID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	})
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 跳过健康检查日志
		if path == "/health" {
			return
		}

		// 计算延迟
		latency := time.Since(startTime)
		clientIP := utils.GetClientIP(c)
		method := c.Request.Method
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()

		// 获取请求ID
		requestID, _ := c.Get("request_id")

		// 构建日志消息
		if raw != "" {
			path = path + "?" + raw
		}

		// 记录访问日志
		entry := logger.GetLogger().WithFields(logrus.Fields{
			"request_id":   requestID,
			"client_ip":    clientIP,
			"method":       method,
			"path":         path,
			"status_code":  statusCode,
			"latency":      latency.String(),
			"user_agent":   userAgent,
			"content_type": c.GetHeader("Content-Type"),
		})

		if len(c.Errors) > 0 {
			// 记录错误
			entry.Error(c.Errors.String())
		} else {
			// 根据状态码设置日志级别
			switch {
			case statusCode >= 500:
				entry.Error("Server Error")
			case statusCode >= 400:
				entry.Warn("Client Error")
			case statusCode >= 300:
				entry.Info("Redirection")
			default:
				entry.Info("Success")
			}
		}
	}
}

// CORS 跨域中间件
func CORS(origins []string, methods []string, headers []string) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = origins
	config.AllowMethods = methods
	config.AllowHeaders = headers
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	return cors.New(config)
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取请求信息
				requestID, _ := c.Get("request_id")
				clientIP := utils.GetClientIP(c)
				method := c.Request.Method
				path := c.Request.URL.Path

				// 记录错误日志
				logger.WithError(fmt.Errorf("panic recovered: %v", err)).
					WithFields(logrus.Fields{
						"request_id": requestID,
						"client_ip":  clientIP,
						"method":     method,
						"path":       path,
					}).Error("Server panic")

				// 返回错误响应
				utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error")
				c.Abort()
			}
		}()

		c.Next()
	}
}

// RateLimiter 简单的速率限制中间件
// 注意：这是一个基本实现，生产环境建议使用Redis等分布式存储
type RateLimiter struct {
	visitors map[string]*visitor
	mu       *sync.RWMutex
	rate     rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		mu:       &sync.RWMutex{},
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}

	// 定期清理过期访问者
	go rl.cleanupVisitors()

	return rl
}

// AllowIP 检查IP是否允许访问
func (rl *RateLimiter) AllowIP(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{limiter, time.Now()}
		return limiter.Allow()
	}

	v.lastSeen = time.Now()
	return v.limiter.Allow()
}

// cleanupVisitors 清理过期访问者
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := utils.GetClientIP(c)
		if !rl.AllowIP(ip) {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateRequest 请求验证中间件
func ValidateRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查Content-Type（对于POST/PUT请求）
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && !strings.Contains(contentType, "application/json") &&
				!strings.Contains(contentType, "multipart/form-data") {
				utils.ErrorResponse(c, http.StatusUnsupportedMediaType, "Unsupported Media Type")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// SecurityHeaders 安全头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全相关的HTTP头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// HTTPS相关（在HTTPS环境下）
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// Timeout 超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置超时上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			defer close(done)
			c.Next()
		}()

		select {
		case <-done:
			// 正常完成
			return
		case <-ctx.Done():
			// 超时
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"code":    http.StatusRequestTimeout,
				"message": "Request timeout",
			})
		}
	}
}