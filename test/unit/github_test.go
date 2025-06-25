package unit

import (
	"strings"
	"testing"

	"github.com/exileum/xenforo-to-gh-discussions/internal/github"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "Valid token",
			token:     "ghp_1234567890abcdef1234567890abcdef12345678",
			shouldErr: false,
		},
		{
			name:      "Empty token",
			token:     "",
			shouldErr: true,
			errMsg:    "GitHub token cannot be empty",
		},
		{
			name:      "Whitespace only token",
			token:     "   \t\n   ",
			shouldErr: true,
			errMsg:    "GitHub token cannot be empty",
		},
		{
			name:      "Token too short",
			token:     "short",
			shouldErr: true,
			errMsg:    "GitHub token appears to be invalid (too short)",
		},
		{
			name:      "Valid but minimal length token",
			token:     "ghp_12345678901234567890", // 24 chars, above minimum
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := github.NewClient(tt.token)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
				if client != nil {
					t.Error("Expected nil client when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if client == nil {
					t.Error("Expected valid client but got nil")
				}
			}
		})
	}
}

func TestClientRepositoryID(t *testing.T) {
	client, err := github.NewClient("ghp_1234567890abcdef1234567890abcdef12345678")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test initial state
	if client.GetRepositoryID() != "" {
		t.Error("New client should have empty repository ID")
	}

	// Test setting repository ID
	testRepoID := "R_kgDOtest123"
	client.SetRepositoryID(testRepoID)

	if client.GetRepositoryID() != testRepoID {
		t.Errorf("Expected repository ID %q, got %q", testRepoID, client.GetRepositoryID())
	}
}
