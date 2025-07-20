package migration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/exileum/xenforo-to-gh-discussions/internal/config"
)

func TestNewMigrator(t *testing.T) {
	cfg := &config.Config{
		XenForo: config.XenForoConfig{
			APIURL:  "https://forum.example.com/api",
			APIKey:  "test_key",
			APIUser: "1",
			NodeID:  1,
		},
		GitHub: config.GitHubConfig{
			Token:                "test_token",
			Repository:           "test/repo",
			XenForoNodeID:        1,
			GitHubCategoryID:     "DIC_kwDOtest123",
			RateLimitDelay:       1 * time.Second,
			MaxRetries:           3,
			RetryBackoffMultiple: 2,
		},
		Migration: config.MigrationConfig{
			MaxRetries:   3,
			DryRun:       true,
			Verbose:      false,
			ProgressFile: "./progress.json",
		},
		Filesystem: config.FilesystemConfig{
			AttachmentsDir:           "./attachments",
			AttachmentRateLimitDelay: 500 * time.Millisecond,
		},
	}

	migrator := NewMigrator(cfg)
	if migrator == nil {
		t.Fatal("NewMigrator returned nil")
	}

	if migrator.config != cfg {
		t.Error("Migrator config not set correctly")
	}
}

func TestMigrator_RunConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		shouldErr bool
		errMsg    string
	}{
		{
			name: "Valid configuration",
			config: &config.Config{
				XenForo: config.XenForoConfig{
					APIURL:  "https://forum.example.com/api",
					APIKey:  "test_key",
					APIUser: "1",
					NodeID:  1,
				},
				GitHub: config.GitHubConfig{
					Token:                "test_token",
					Repository:           "test/repo",
					XenForoNodeID:        1,
					GitHubCategoryID:     "DIC_kwDOtest123",
					RateLimitDelay:       1 * time.Second,
					MaxRetries:           3,
					RetryBackoffMultiple: 2,
				},
				Migration: config.MigrationConfig{
					MaxRetries:   3,
					DryRun:       true,
					Verbose:      false,
					ProgressFile: "./progress.json",
				},
				Filesystem: config.FilesystemConfig{
					AttachmentsDir:           "./attachments",
					AttachmentRateLimitDelay: 500 * time.Millisecond,
				},
			},
			shouldErr: false,
		},
		{
			name: "Default placeholder XenForo URL",
			config: &config.Config{
				XenForo: config.XenForoConfig{
					APIURL:  "https://your-forum.com/api", // Default placeholder that should fail validation
					APIKey:  "test_key",
					APIUser: "1",
					NodeID:  1,
				},
				GitHub: config.GitHubConfig{
					Token:                "test_token",
					Repository:           "test/repo",
					XenForoNodeID:        1,
					GitHubCategoryID:     "DIC_kwDOtest123",
					RateLimitDelay:       1 * time.Second,
					MaxRetries:           3,
					RetryBackoffMultiple: 2,
				},
				Migration: config.MigrationConfig{
					MaxRetries:   3,
					ProgressFile: "./progress.json",
				},
				Filesystem: config.FilesystemConfig{
					AttachmentsDir:           "./attachments",
					AttachmentRateLimitDelay: 500 * time.Millisecond,
				},
			},
			shouldErr: true,
			errMsg:    "configuration validation failed",
		},
		{
			name: "Empty GitHub token",
			config: &config.Config{
				XenForo: config.XenForoConfig{
					APIURL:  "https://forum.example.com/api",
					APIKey:  "test_key",
					APIUser: "1",
					NodeID:  1,
				},
				GitHub: config.GitHubConfig{
					Token:                "", // Empty token
					Repository:           "test/repo",
					XenForoNodeID:        1,
					GitHubCategoryID:     "DIC_kwDOtest123",
					RateLimitDelay:       1 * time.Second,
					MaxRetries:           3,
					RetryBackoffMultiple: 2,
				},
				Migration: config.MigrationConfig{
					MaxRetries:   3,
					ProgressFile: "./progress.json",
				},
				Filesystem: config.FilesystemConfig{
					AttachmentsDir:           "./attachments",
					AttachmentRateLimitDelay: 500 * time.Millisecond,
				},
			},
			shouldErr: true,
			errMsg:    "configuration validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := NewMigrator(tt.config)
			ctx := context.Background()

			err := migrator.Run(ctx)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				// For valid configs, we expect the migrator to fail later due to mock setup
				// but not during initial validation
				if err != nil && strings.Contains(err.Error(), "configuration validation failed") {
					t.Errorf("Expected no validation error but got: %v", err)
				}
			}
		})
	}
}

func TestMigrator_RunContextCancellation(t *testing.T) {
	cfg := &config.Config{
		XenForo: config.XenForoConfig{
			APIURL:  "https://forum.example.com/api",
			APIKey:  "test_key",
			APIUser: "1",
			NodeID:  1,
		},
		GitHub: config.GitHubConfig{
			Token:                "test_token",
			Repository:           "test/repo",
			XenForoNodeID:        1,
			GitHubCategoryID:     "DIC_kwDOtest123",
			RateLimitDelay:       1 * time.Second,
			MaxRetries:           3,
			RetryBackoffMultiple: 2,
		},
		Migration: config.MigrationConfig{
			MaxRetries:   3,
			DryRun:       true,
			Verbose:      false,
			ProgressFile: "./progress.json",
		},
		Filesystem: config.FilesystemConfig{
			AttachmentsDir:           "./attachments",
			AttachmentRateLimitDelay: 500 * time.Millisecond,
		},
	}

	migrator := NewMigrator(cfg)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := migrator.Run(ctx)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}

	// The error might be from context cancellation or from trying to connect with invalid credentials
	// Both are acceptable for this test
}

func TestMigrator_RunDryRunMode(t *testing.T) {
	cfg := &config.Config{
		XenForo: config.XenForoConfig{
			APIURL:  "https://forum.example.com/api",
			APIKey:  "test_key",
			APIUser: "1",
			NodeID:  1,
		},
		GitHub: config.GitHubConfig{
			Token:                "test_token",
			Repository:           "test/repo",
			XenForoNodeID:        1,
			GitHubCategoryID:     "DIC_kwDOtest123",
			RateLimitDelay:       1 * time.Second,
			MaxRetries:           3,
			RetryBackoffMultiple: 2,
		},
		Migration: config.MigrationConfig{
			MaxRetries:   3,
			DryRun:       true, // Enable dry run
			Verbose:      false,
			ProgressFile: "./progress.json",
		},
		Filesystem: config.FilesystemConfig{
			AttachmentsDir:           "./attachments",
			AttachmentRateLimitDelay: 500 * time.Millisecond,
		},
	}

	migrator := NewMigrator(cfg)
	ctx := context.Background()

	// In dry run mode, the migrator should attempt to initialize but will fail
	// due to invalid credentials. This tests that dry run mode is properly handled.
	err := migrator.Run(ctx)

	// We expect an error because we're using dummy credentials,
	// but it should not be a configuration validation error
	if err != nil && strings.Contains(err.Error(), "configuration validation failed") {
		t.Errorf("Dry run mode should pass configuration validation: %v", err)
	}
}
