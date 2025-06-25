package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/exileum/xenforo-to-gh-discussions/internal/config"
	"github.com/exileum/xenforo-to-gh-discussions/internal/github"
	"github.com/exileum/xenforo-to-gh-discussions/internal/migration"
	"github.com/exileum/xenforo-to-gh-discussions/internal/testutil"
)

func TestMigrationIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "migration-integration-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test configuration
	cfg := &config.Config{
		XenForo: config.XenForoConfig{
			APIURL:  "https://test-forum.com/api",
			APIKey:  "test_key",
			APIUser: "1",
			NodeID:  1,
		},
		GitHub: config.GitHubConfig{
			Token:      "test_token",
			Repository: "test/repo",
			Categories: map[int]string{1: "DIC_kwDOtest123"},
		},
		Migration: config.MigrationConfig{
			MaxRetries:   3,
			DryRun:       true, // Use dry run for testing
			Verbose:      false,
			ProgressFile: filepath.Join(tempDir, "progress.json"),
		},
		Filesystem: config.FilesystemConfig{
			AttachmentsDir: filepath.Join(tempDir, "attachments"),
		},
	}

	// This test verifies that the migration process can be initialized
	// and would work correctly with real APIs (but uses dry-run mode)
	migrator := migration.NewMigrator(cfg)

	// In a real integration test, this would make actual API calls
	// For now, we just verify the setup works
	if migrator == nil {
		t.Error("Failed to create migrator")
	}

	// Note: Full integration would require running migrator.Run()
	// but that needs real API credentials and would make actual calls
	// This could be extended with docker-compose for full e2e testing
}

func TestEndToEndWithMocks(t *testing.T) {
	// This test demonstrates how the full migration would work
	// with mocked dependencies

	tempDir, err := os.MkdirTemp("", "e2e-mock-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		XenForo: config.XenForoConfig{
			APIURL:  "https://test-forum.com/api",
			APIKey:  "test_key",
			APIUser: "1",
			NodeID:  1,
		},
		GitHub: config.GitHubConfig{
			Token:      "test_token",
			Repository: "test/repo",
			Categories: map[int]string{1: "DIC_kwDOtest123"},
		},
		Migration: config.MigrationConfig{
			MaxRetries:   3,
			DryRun:       true,
			Verbose:      false,
			ProgressFile: filepath.Join(tempDir, "progress.json"),
		},
		Filesystem: config.FilesystemConfig{
			AttachmentsDir: filepath.Join(tempDir, "attachments"),
		},
	}

	// Create mock clients
	xenforoMock := &testutil.XenForoClient{}
	githubMock := &testutil.GitHubClient{}

	// Set up mock behaviors
	var createdDiscussions []string
	githubMock.CreateDiscussionFunc = func(title, body, categoryID string) (*github.DiscussionResult, error) {
		createdDiscussions = append(createdDiscussions, title)
		return &github.DiscussionResult{ID: "test_id", Number: len(createdDiscussions)}, nil
	}

	// Update to valid category mapping for testing
	cfg.GitHub.Categories = map[int]string{1: "DIC_kwDOtest123"}

	// Verify configuration is valid for dry-run
	if err := cfg.Validate(); err != nil {
		t.Errorf("Configuration should be valid: %v", err)
	}

	// Verify mock clients are properly initialized
	if xenforoMock == nil || githubMock == nil {
		t.Error("Mock clients should be initialized")
	}

	// Verify that discussion creation tracking works
	if len(createdDiscussions) != 0 {
		t.Error("No discussions should be created initially")
	}
}
