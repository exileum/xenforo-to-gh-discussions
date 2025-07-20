package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	cfg := New()

	if cfg.XenForo.APIURL != "https://your-forum.com/api" {
		t.Error("Default XenForo API URL not set correctly")
	}

	if cfg.Migration.MaxRetries != 3 {
		t.Error("Default max retries not set correctly")
	}

	if cfg.Filesystem.AttachmentsDir != "./attachments" {
		t.Error("Default attachments directory not set correctly")
	}

	if cfg.Filesystem.AttachmentRateLimitDelay != 500*time.Millisecond {
		t.Error("Default attachment rate limit delay not set correctly")
	}
}

func TestConfigEnvironmentVariables(t *testing.T) {
	// Set environment variables
	if err := os.Setenv("XENFORO_API_URL", "https://test-forum.com/api"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("MAX_RETRIES", "5"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("ATTACHMENT_RATE_LIMIT_DELAY", "1s"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Unsetenv("XENFORO_API_URL")
		_ = os.Unsetenv("MAX_RETRIES")
		_ = os.Unsetenv("ATTACHMENT_RATE_LIMIT_DELAY")
	}()

	cfg := New()

	if cfg.XenForo.APIURL != "https://test-forum.com/api" {
		t.Error("Environment variable for XenForo API URL not used")
	}

	if cfg.Migration.MaxRetries != 5 {
		t.Error("Environment variable for max retries not used")
	}

	if cfg.Filesystem.AttachmentRateLimitDelay != 1*time.Second {
		t.Error("Environment variable for attachment rate limit delay not used")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Config)
		shouldErr bool
	}{
		{
			name: "Valid config",
			setup: func(cfg *Config) {
				cfg.XenForo.APIURL = "https://forum.example.com/api"
				cfg.XenForo.APIKey = "valid_key"
				cfg.XenForo.APIUser = "1"
				cfg.XenForo.NodeID = 1
				cfg.GitHub.Token = "valid_token"
				cfg.GitHub.Repository = "owner/repo"
				cfg.GitHub.Categories = map[int]string{1: "DIC_kwDOtest123"}
				cfg.GitHub.XenForoNodeID = 1
				cfg.GitHub.GitHubCategoryID = "DIC_kwDOtest123"
			},
			shouldErr: false,
		},
		{
			name: "Invalid XenForo URL",
			setup: func(cfg *Config) {
				cfg.XenForo.APIURL = "https://your-forum.com/api"
			},
			shouldErr: true,
		},
		{
			name: "Invalid GitHub repository format",
			setup: func(cfg *Config) {
				cfg.XenForo.APIURL = "https://forum.example.com/api"
				cfg.XenForo.APIKey = "valid_key"
				cfg.XenForo.APIUser = "1"
				cfg.XenForo.NodeID = 1
				cfg.GitHub.Token = "valid_token"
				cfg.GitHub.Repository = "invalid_format"
			},
			shouldErr: true,
		},
		{
			name: "No category mappings",
			setup: func(cfg *Config) {
				cfg.XenForo.APIURL = "https://forum.example.com/api"
				cfg.XenForo.APIKey = "valid_key"
				cfg.XenForo.APIUser = "1"
				cfg.XenForo.NodeID = 1
				cfg.GitHub.Token = "valid_token"
				cfg.GitHub.Repository = "owner/repo"
				cfg.GitHub.Categories = map[int]string{}
				cfg.GitHub.XenForoNodeID = 0
				cfg.GitHub.GitHubCategoryID = ""
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()
			tt.setup(cfg)

			err := cfg.Validate()
			if tt.shouldErr && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no validation error but got: %v", err)
			}
		})
	}
}
