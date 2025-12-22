package monitoring

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck 健康检查结果
type HealthCheck struct {
	Status    HealthStatus     `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Checks    map[string]Check `json:"checks"`
	Duration  time.Duration    `json:"duration"`
}

// Check 单项检查结果
type Check struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message"`
	Error   string       `json:"error,omitempty"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	db *gorm.DB
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(db *gorm.DB) *HealthChecker {
	return &HealthChecker{
		db: db,
	}
}

// CheckHealth 执行健康检查
func (h *HealthChecker) CheckHealth(ctx context.Context) *HealthCheck {
	start := time.Now()

	checks := make(map[string]Check)

	// 检查数据库连接
	checks["database"] = h.checkDatabase(ctx)

	// 检查磁盘空间（可选）
	checks["disk_space"] = h.checkDiskSpace()

	// 检查内存使用（可选）
	checks["memory"] = h.checkMemory()

	// 确定整体状态
	overallStatus := h.determineOverallStatus(checks)

	return &HealthCheck{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    checks,
		Duration:  time.Since(start),
	}
}

// checkDatabase 检查数据库连接
func (h *HealthChecker) checkDatabase(ctx context.Context) Check {
	if h.db == nil {
		return Check{
			Status:  HealthStatusUnhealthy,
			Message: "Database connection not initialized",
		}
	}

	sqlDB, err := h.db.DB()
	if err != nil {
		return Check{
			Status:  HealthStatusUnhealthy,
			Message: "Failed to get underlying database connection",
			Error:   err.Error(),
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return Check{
			Status:  HealthStatusUnhealthy,
			Message: "Database ping failed",
			Error:   err.Error(),
		}
	}

	return Check{
		Status:  HealthStatusHealthy,
		Message: "Database connection is healthy",
	}
}

// checkDiskSpace 检查磁盘空间
func (h *HealthChecker) checkDiskSpace() Check {
	// 简化实现，实际应用中可以检查具体的磁盘使用情况
	return Check{
		Status:  HealthStatusHealthy,
		Message: "Disk space check not implemented",
	}
}

// checkMemory 检查内存使用
func (h *HealthChecker) checkMemory() Check {
	// 简化实现，实际应用中可以检查内存使用情况
	return Check{
		Status:  HealthStatusHealthy,
		Message: "Memory check not implemented",
	}
}

// determineOverallStatus 确定整体健康状态
func (h *HealthChecker) determineOverallStatus(checks map[string]Check) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false

	for _, check := range checks {
		switch check.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}
