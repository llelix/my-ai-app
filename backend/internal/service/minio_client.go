package service

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"strings"
	"syscall"
	"time"

	"ai-knowledge-app/internal/config"
	"ai-knowledge-app/pkg/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/minio/minio-go/v7"
	miniocreds "github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

// RetryConfig defines retry behavior for MinIO operations
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"connection refused",
			"connection reset",
			"timeout",
			"temporary failure",
			"network is unreachable",
			"no route to host",
			"service unavailable",
			"internal server error",
			"bad gateway",
			"gateway timeout",
		},
	}
}

// MinIOClient wraps the MinIO client with configuration and retry logic
type MinIOClient struct {
	client      *minio.Client
	s3Client    *s3.Client
	config      *config.S3Config
	retryConfig *RetryConfig
	logger      *logrus.Logger
}

// NewMinIOClient creates a new MinIO client instance with retry capabilities
func NewMinIOClient(cfg *config.S3Config) (*MinIOClient, error) {
	// Get logger instance
	log := logger.GetLogger()
	if log == nil {
		log = logrus.New()
	}

	log.WithFields(logrus.Fields{
		"endpoint": cfg.Endpoint,
		"bucket":   cfg.Bucket,
		"region":   cfg.Region,
		"use_ssl":  cfg.UseSSL,
	}).Info("Initializing MinIO client")

	// Initialize MinIO client
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  miniocreds.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		log.WithError(err).Error("Failed to create MinIO client")
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Initialize AWS S3 client for multipart uploads
	awsConfig := aws.Config{
		Region:      cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               fmt.Sprintf("http%s://%s", map[bool]string{true: "s", false: ""}[cfg.UseSSL], cfg.Endpoint),
				HostnameImmutable: true,
			}, nil
		}),
	}

	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true // Required for MinIO
	})

	client := &MinIOClient{
		client:      minioClient,
		s3Client:    s3Client,
		config:      cfg,
		retryConfig: DefaultRetryConfig(),
		logger:      log,
	}

	// Test connection and create bucket if needed with retry
	if err := client.initializeBucketWithRetry(); err != nil {
		log.WithError(err).Error("Failed to initialize bucket after retries")
		return nil, fmt.Errorf("failed to initialize bucket: %w", err)
	}

	log.Info("MinIO client initialized successfully")
	return client, nil
}

// isRetryableError checks if an error is retryable based on error message
func (m *MinIOClient) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	
	// Check for network-related errors
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}

	// Check for syscall errors
	if opErr, ok := err.(*net.OpError); ok {
		if opErr.Op == "dial" || opErr.Op == "read" || opErr.Op == "write" {
			return true
		}
		if sysErr, ok := opErr.Err.(*net.DNSError); ok && sysErr.Temporary() {
			return true
		}
		if sysErr, ok := opErr.Err.(syscall.Errno); ok {
			switch sysErr {
			case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.ETIMEDOUT, syscall.EHOSTUNREACH, syscall.ENETUNREACH:
				return true
			}
		}
	}

	// Check for specific error messages
	for _, retryableErr := range m.retryConfig.RetryableErrors {
		if strings.Contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// calculateBackoffDelay calculates the delay for the next retry attempt
func (m *MinIOClient) calculateBackoffDelay(attempt int) time.Duration {
	delay := time.Duration(float64(m.retryConfig.InitialDelay) * math.Pow(m.retryConfig.BackoffFactor, float64(attempt)))
	if delay > m.retryConfig.MaxDelay {
		delay = m.retryConfig.MaxDelay
	}
	return delay
}

// retryOperation executes an operation with retry logic
func (m *MinIOClient) retryOperation(operation func() error, operationName string) error {
	var lastErr error
	
	for attempt := 0; attempt <= m.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := m.calculateBackoffDelay(attempt - 1)
			m.logger.WithFields(logrus.Fields{
				"operation": operationName,
				"attempt":   attempt,
				"delay":     delay,
				"error":     lastErr,
			}).Warn("Retrying MinIO operation after failure")
			time.Sleep(delay)
		}

		lastErr = operation()
		if lastErr == nil {
			if attempt > 0 {
				m.logger.WithFields(logrus.Fields{
					"operation": operationName,
					"attempt":   attempt,
				}).Info("MinIO operation succeeded after retry")
			}
			return nil
		}

		if !m.isRetryableError(lastErr) {
			m.logger.WithFields(logrus.Fields{
				"operation": operationName,
				"error":     lastErr,
			}).Error("MinIO operation failed with non-retryable error")
			break
		}

		m.logger.WithFields(logrus.Fields{
			"operation": operationName,
			"attempt":   attempt,
			"error":     lastErr,
		}).Debug("MinIO operation failed, will retry")
	}

	m.logger.WithFields(logrus.Fields{
		"operation":    operationName,
		"max_retries":  m.retryConfig.MaxRetries,
		"final_error":  lastErr,
	}).Error("MinIO operation failed after all retry attempts")

	return fmt.Errorf("operation %s failed after %d retries: %w", operationName, m.retryConfig.MaxRetries, lastErr)
}

// initializeBucketWithRetry tests connection and creates bucket if it doesn't exist with retry logic
func (m *MinIOClient) initializeBucketWithRetry() error {
	return m.retryOperation(func() error {
		return m.initializeBucket()
	}, "initialize_bucket")
}

// initializeBucket tests connection and creates bucket if it doesn't exist
func (m *MinIOClient) initializeBucket() error {
	ctx := context.Background()

	// Test connection by checking if bucket exists
	exists, err := m.client.BucketExists(ctx, m.config.Bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	// Create bucket if it doesn't exist
	if !exists {
		m.logger.WithField("bucket", m.config.Bucket).Info("Creating MinIO bucket")
		err = m.client.MakeBucket(ctx, m.config.Bucket, minio.MakeBucketOptions{
			Region: m.config.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", m.config.Bucket, err)
		}
		m.logger.WithField("bucket", m.config.Bucket).Info("Successfully created MinIO bucket")
	} else {
		m.logger.WithField("bucket", m.config.Bucket).Debug("MinIO bucket already exists")
	}

	return nil
}

// TestConnection tests the MinIO connection with retry logic
func (m *MinIOClient) TestConnection() error {
	return m.retryOperation(func() error {
		ctx := context.Background()
		
		// Test connection by listing buckets
		_, err := m.client.ListBuckets(ctx)
		if err != nil {
			return fmt.Errorf("connection test failed: %w", err)
		}
		
		return nil
	}, "test_connection")
}

// GetClient returns the underlying MinIO client
func (m *MinIOClient) GetClient() *minio.Client {
	return m.client
}

// GetBucketName returns the configured bucket name
func (m *MinIOClient) GetBucketName() string {
	return m.config.Bucket
}

// GetS3Client returns the AWS S3 client for multipart uploads
func (m *MinIOClient) GetS3Client() *s3.Client {
	return m.s3Client
}

// PutObjectWithRetry uploads an object to MinIO with retry logic
func (m *MinIOClient) PutObjectWithRetry(ctx context.Context, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	var result minio.UploadInfo
	var err error
	
	err = m.retryOperation(func() error {
		result, err = m.client.PutObject(ctx, m.config.Bucket, objectName, reader, objectSize, opts)
		return err
	}, fmt.Sprintf("put_object_%s", objectName))
	
	return result, err
}

// GetObjectWithRetry retrieves an object from MinIO with retry logic
func (m *MinIOClient) GetObjectWithRetry(ctx context.Context, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	var result *minio.Object
	var err error
	
	err = m.retryOperation(func() error {
		result, err = m.client.GetObject(ctx, m.config.Bucket, objectName, opts)
		return err
	}, fmt.Sprintf("get_object_%s", objectName))
	
	return result, err
}

// StatObjectWithRetry gets object metadata from MinIO with retry logic
func (m *MinIOClient) StatObjectWithRetry(ctx context.Context, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	var result minio.ObjectInfo
	var err error
	
	err = m.retryOperation(func() error {
		result, err = m.client.StatObject(ctx, m.config.Bucket, objectName, opts)
		return err
	}, fmt.Sprintf("stat_object_%s", objectName))
	
	return result, err
}

// RemoveObjectWithRetry removes an object from MinIO with retry logic
func (m *MinIOClient) RemoveObjectWithRetry(ctx context.Context, objectName string, opts minio.RemoveObjectOptions) error {
	return m.retryOperation(func() error {
		return m.client.RemoveObject(ctx, m.config.Bucket, objectName, opts)
	}, fmt.Sprintf("remove_object_%s", objectName))
}

// ListObjectsWithRetry lists objects in MinIO with retry logic
func (m *MinIOClient) ListObjectsWithRetry(ctx context.Context, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	// Note: ListObjects returns a channel, so we can't easily retry individual operations
	// Instead, we'll add logging for when the operation starts
	m.logger.WithFields(logrus.Fields{
		"bucket": m.config.Bucket,
		"prefix": opts.Prefix,
	}).Debug("Starting MinIO list objects operation")
	
	return m.client.ListObjects(ctx, m.config.Bucket, opts)
}

// S3 multipart upload operations with retry logic

// CreateMultipartUploadWithRetry creates a multipart upload with retry logic
func (m *MinIOClient) CreateMultipartUploadWithRetry(ctx context.Context, input *s3.CreateMultipartUploadInput) (*s3.CreateMultipartUploadOutput, error) {
	var result *s3.CreateMultipartUploadOutput
	var err error
	
	err = m.retryOperation(func() error {
		result, err = m.s3Client.CreateMultipartUpload(ctx, input)
		return err
	}, fmt.Sprintf("create_multipart_upload_%s", *input.Key))
	
	return result, err
}

// UploadPartWithRetry uploads a part for multipart upload with retry logic
func (m *MinIOClient) UploadPartWithRetry(ctx context.Context, input *s3.UploadPartInput) (*s3.UploadPartOutput, error) {
	var result *s3.UploadPartOutput
	var err error
	
	err = m.retryOperation(func() error {
		result, err = m.s3Client.UploadPart(ctx, input)
		return err
	}, fmt.Sprintf("upload_part_%s_part_%d", *input.Key, *input.PartNumber))
	
	return result, err
}

// CompleteMultipartUploadWithRetry completes a multipart upload with retry logic
func (m *MinIOClient) CompleteMultipartUploadWithRetry(ctx context.Context, input *s3.CompleteMultipartUploadInput) (*s3.CompleteMultipartUploadOutput, error) {
	var result *s3.CompleteMultipartUploadOutput
	var err error
	
	err = m.retryOperation(func() error {
		result, err = m.s3Client.CompleteMultipartUpload(ctx, input)
		return err
	}, fmt.Sprintf("complete_multipart_upload_%s", *input.Key))
	
	return result, err
}

// AbortMultipartUploadWithRetry aborts a multipart upload with retry logic
func (m *MinIOClient) AbortMultipartUploadWithRetry(ctx context.Context, input *s3.AbortMultipartUploadInput) (*s3.AbortMultipartUploadOutput, error) {
	var result *s3.AbortMultipartUploadOutput
	var err error
	
	err = m.retryOperation(func() error {
		result, err = m.s3Client.AbortMultipartUpload(ctx, input)
		return err
	}, fmt.Sprintf("abort_multipart_upload_%s", *input.Key))
	
	return result, err
}

// ListPartsWithRetry lists parts of a multipart upload with retry logic
func (m *MinIOClient) ListPartsWithRetry(ctx context.Context, input *s3.ListPartsInput) (*s3.ListPartsOutput, error) {
	var result *s3.ListPartsOutput
	var err error
	
	err = m.retryOperation(func() error {
		result, err = m.s3Client.ListParts(ctx, input)
		return err
	}, fmt.Sprintf("list_parts_%s", *input.Key))
	
	return result, err
}

// IsHealthy checks if the MinIO service is available and healthy
func (m *MinIOClient) IsHealthy() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Try to list buckets as a health check
	_, err := m.client.ListBuckets(ctx)
	if err != nil {
		m.logger.WithError(err).Error("MinIO health check failed")
		return fmt.Errorf("MinIO service is not healthy: %w", err)
	}
	
	return nil
}

// IsServiceAvailable checks if MinIO service is available without retries
func (m *MinIOClient) IsServiceAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := m.client.ListBuckets(ctx)
	return err == nil
}

// GetRetryConfig returns the current retry configuration
func (m *MinIOClient) GetRetryConfig() *RetryConfig {
	return m.retryConfig
}

// SetRetryConfig updates the retry configuration
func (m *MinIOClient) SetRetryConfig(config *RetryConfig) {
	m.retryConfig = config
	m.logger.WithFields(logrus.Fields{
		"max_retries":      config.MaxRetries,
		"initial_delay":    config.InitialDelay,
		"max_delay":        config.MaxDelay,
		"backoff_factor":   config.BackoffFactor,
	}).Info("Updated MinIO retry configuration")
}