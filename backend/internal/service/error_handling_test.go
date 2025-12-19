package service

import (
	"errors"
	"net"
	"strings"
	"syscall"
	"testing"
	"time"

	"ai-knowledge-app/internal/config"
	"github.com/sirupsen/logrus"
)

// Helper function to check if a string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

func TestRetryLogic(t *testing.T) {
	// Create a MinIO client with custom retry config for faster testing
	cfg := &config.S3Config{
		Endpoint:        "localhost:9999", // Non-existent endpoint
		AccessKeyID:     "test",
		SecretAccessKey: "test",
		UseSSL:          false,
		Bucket:          "test-bucket",
		Region:          "us-east-1",
	}

	// Disable logging for cleaner test output
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	client := &MinIOClient{
		config: cfg,
		retryConfig: &RetryConfig{
			MaxRetries:    2,
			InitialDelay:  10 * time.Millisecond,
			MaxDelay:      100 * time.Millisecond,
			BackoffFactor: 2.0,
			RetryableErrors: []string{
				"connection refused",
				"connection reset",
				"timeout",
			},
		},
		logger: logger,
	}

	t.Run("TestRetryableErrorDetection", func(t *testing.T) {
		// Test network errors
		netErr := &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}
		if !client.isRetryableError(netErr) {
			t.Error("Expected connection refused error to be retryable")
		}

		// Test timeout errors
		timeoutErr := &net.OpError{Op: "read", Err: &net.DNSError{IsTimeout: true}}
		if !client.isRetryableError(timeoutErr) {
			t.Error("Expected timeout error to be retryable")
		}

		// Test non-retryable errors
		authErr := errors.New("authentication failed")
		if client.isRetryableError(authErr) {
			t.Error("Expected authentication error to be non-retryable")
		}
	})

	t.Run("TestBackoffCalculation", func(t *testing.T) {
		// Test exponential backoff
		delay1 := client.calculateBackoffDelay(0)
		delay2 := client.calculateBackoffDelay(1)
		delay3 := client.calculateBackoffDelay(2)

		if delay1 != 10*time.Millisecond {
			t.Errorf("Expected first delay to be 10ms, got %v", delay1)
		}

		if delay2 != 20*time.Millisecond {
			t.Errorf("Expected second delay to be 20ms, got %v", delay2)
		}

		if delay3 != 40*time.Millisecond {
			t.Errorf("Expected third delay to be 40ms, got %v", delay3)
		}

		// Test max delay cap
		longDelay := client.calculateBackoffDelay(10)
		if longDelay != client.retryConfig.MaxDelay {
			t.Errorf("Expected delay to be capped at %v, got %v", client.retryConfig.MaxDelay, longDelay)
		}
	})

	t.Run("TestRetryOperation", func(t *testing.T) {
		// Test successful operation (no retries needed)
		attempts := 0
		err := client.retryOperation(func() error {
			attempts++
			return nil
		}, "test_success")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}

		// Test retryable error that eventually succeeds
		attempts = 0
		err = client.retryOperation(func() error {
			attempts++
			if attempts < 3 {
				return &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}
			}
			return nil
		}, "test_retry_success")

		if err != nil {
			t.Errorf("Expected no error after retries, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}

		// Test non-retryable error
		attempts = 0
		err = client.retryOperation(func() error {
			attempts++
			return errors.New("authentication failed")
		}, "test_non_retryable")

		if err == nil {
			t.Error("Expected error for non-retryable failure")
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempts)
		}

		// Test retryable error that always fails
		attempts = 0
		err = client.retryOperation(func() error {
			attempts++
			return &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}
		}, "test_retry_failure")

		if err == nil {
			t.Error("Expected error after max retries")
		}
		expectedAttempts := client.retryConfig.MaxRetries + 1
		if attempts != expectedAttempts {
			t.Errorf("Expected %d attempts, got %d", expectedAttempts, attempts)
		}
	})
}

func TestMinIOServiceAvailability(t *testing.T) {
	// Test with non-existent MinIO service
	cfg := &config.S3Config{
		Endpoint:        "localhost:9999",
		AccessKeyID:     "test",
		SecretAccessKey: "test",
		UseSSL:          false,
		Bucket:          "test-bucket",
		Region:          "us-east-1",
	}

	t.Run("TestServiceAvailabilityCheck", func(t *testing.T) {
		// Try to create a client - this should fail due to unavailable service
		// but we'll catch the error and test the availability check
		logger := logrus.New()
		logger.SetLevel(logrus.FatalLevel)
		
		// We can't test IsServiceAvailable without a properly initialized client
		// So we'll test the creation failure instead
		_, err := NewMinIOClient(cfg)
		if err == nil {
			t.Error("Expected MinIO client creation to fail when service is unavailable")
		}
		
		// The error should indicate connection issues
		if err != nil && !containsAny(err.Error(), []string{"connection refused", "failed to initialize"}) {
			t.Errorf("Expected connection-related error, got: %v", err)
		}
	})
}

func TestDocumentServiceMinIOIntegration(t *testing.T) {
	// Test DocumentService methods for MinIO availability
	service := NewDocumentService(nil)

	t.Run("TestMinIONotConfigured", func(t *testing.T) {
		if service.IsMinIOAvailable() {
			t.Error("Expected MinIO to be unavailable when not configured")
		}

		err := service.CheckMinIOHealth()
		if err == nil {
			t.Error("Expected health check to fail when MinIO not configured")
		}
		
		// Verify error message
		if !strings.Contains(err.Error(), "not configured") {
			t.Errorf("Expected 'not configured' error, got: %v", err)
		}
	})
}