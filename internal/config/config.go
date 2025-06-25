package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	XenForo    XenForoConfig
	GitHub     GitHubConfig
	Migration  MigrationConfig
	Filesystem FilesystemConfig
}

type XenForoConfig struct {
	APIURL  string
	APIKey  string
	APIUser string
	NodeID  int
}

type GitHubConfig struct {
	Token      string
	Repository string
	Categories map[int]string
}

type MigrationConfig struct {
	MaxRetries   int
	DryRun       bool
	Verbose      bool
	ResumeFrom   int
	ProgressFile string
}

type FilesystemConfig struct {
	AttachmentsDir           string
	AttachmentRateLimitDelay time.Duration
}

func New() *Config {
	return &Config{
		XenForo: XenForoConfig{
			APIURL:  getEnvOrDefault("XENFORO_API_URL", "https://your-forum.com/api"),
			APIKey:  getEnvOrDefault("XENFORO_API_KEY", "your_xenforo_api_key"),
			APIUser: getEnvOrDefault("XENFORO_API_USER", "1"),
			NodeID:  getEnvIntOrDefault("XENFORO_NODE_ID", 1),
		},
		GitHub: GitHubConfig{
			Token:      getEnvOrDefault("GITHUB_TOKEN", "your_github_token"),
			Repository: getEnvOrDefault("GITHUB_REPO", "your_username/your_repo"),
			Categories: map[int]string{
				1: "DIC_kwDOxxxxxxxx", // Default mapping
			},
		},
		Migration: MigrationConfig{
			MaxRetries:   getEnvIntOrDefault("MAX_RETRIES", 3),
			ProgressFile: getEnvOrDefault("PROGRESS_FILE", "migration_progress.json"),
		},
		Filesystem: FilesystemConfig{
			AttachmentsDir:           getEnvOrDefault("ATTACHMENTS_DIR", "./attachments"),
			AttachmentRateLimitDelay: getEnvDurationOrDefault("ATTACHMENT_RATE_LIMIT_DELAY", 500*time.Millisecond),
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
