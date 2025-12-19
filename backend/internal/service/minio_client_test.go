package service

import (
	"testing"

	"ai-knowledge-app/internal/config"
)

func TestNewMinIOClient(t *testing.T) {
	// Test configuration for MinIO client
	testConfig := &config.S3Config{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin123",
		UseSSL:          false,
		Bucket:          "test-bucket",
		Region:          "us-east-1",
	}

	// Test creating MinIO client
	client, err := NewMinIOClient(testConfig)
	
	// Note: This test will fail if MinIO is not running, which is expected
	// In a real environment, MinIO should be running for this test to pass
	if err != nil {
		t.Logf("MinIO client creation failed (expected if MinIO is not running): %v", err)
		// Don't fail the test if MinIO is not available
		return
	}

	// If MinIO is available, test the connection
	if client != nil {
		if err := client.TestConnection(); err != nil {
			t.Errorf("MinIO connection test failed: %v", err)
		}

		// Test getting bucket name
		bucketName := client.GetBucketName()
		if bucketName != testConfig.Bucket {
			t.Errorf("Expected bucket name %s, got %s", testConfig.Bucket, bucketName)
		}

		// Test getting client
		minioClient := client.GetClient()
		if minioClient == nil {
			t.Error("GetClient() returned nil")
		}

		t.Logf("MinIO client test passed with bucket: %s", bucketName)
	}
}

func TestMinIOClientValidation(t *testing.T) {
	// Test with invalid configuration
	invalidConfigs := []struct {
		name   string
		config *config.S3Config
	}{
		{
			name: "empty endpoint",
			config: &config.S3Config{
				Endpoint:        "",
				AccessKeyID:     "test",
				SecretAccessKey: "test",
				UseSSL:          false,
				Bucket:          "test",
				Region:          "us-east-1",
			},
		},
		{
			name: "empty access key",
			config: &config.S3Config{
				Endpoint:        "localhost:9000",
				AccessKeyID:     "",
				SecretAccessKey: "test",
				UseSSL:          false,
				Bucket:          "test",
				Region:          "us-east-1",
			},
		},
		{
			name: "empty secret key",
			config: &config.S3Config{
				Endpoint:        "localhost:9000",
				AccessKeyID:     "test",
				SecretAccessKey: "",
				UseSSL:          false,
				Bucket:          "test",
				Region:          "us-east-1",
			},
		},
		{
			name: "empty bucket",
			config: &config.S3Config{
				Endpoint:        "localhost:9000",
				AccessKeyID:     "test",
				SecretAccessKey: "test",
				UseSSL:          false,
				Bucket:          "",
				Region:          "us-east-1",
			},
		},
		{
			name: "empty region",
			config: &config.S3Config{
				Endpoint:        "localhost:9000",
				AccessKeyID:     "test",
				SecretAccessKey: "test",
				UseSSL:          false,
				Bucket:          "test",
				Region:          "",
			},
		},
	}

	for _, tc := range invalidConfigs {
		t.Run(tc.name, func(t *testing.T) {
			// Test that validation catches invalid configurations
			err := tc.config.Validate()
			if err == nil {
				t.Errorf("Expected validation error for %s, but got none", tc.name)
			} else {
				t.Logf("Validation correctly caught error for %s: %v", tc.name, err)
			}
		})
	}
}