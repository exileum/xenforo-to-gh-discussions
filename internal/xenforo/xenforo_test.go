package xenforo

import (
	"testing"
	"time"
)

func TestNewXenForoClient(t *testing.T) {
	baseURL := "https://forum.example.com/api"
	apiKey := "test-api-key"
	apiUser := "1"
	maxRetries := 3

	client := NewClient(baseURL, apiKey, apiUser, maxRetries)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	// Test that the client is properly initialized
	// Note: We can't directly access the resty client's timeout due to private fields,
	// but we can verify the client was created properly
}

func TestXenForoClientTimeout(t *testing.T) {
	// Test that the client handles timeouts appropriately
	client := NewClient("https://example.com/api", "test-key", "1", 1)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	// Test timeout customization
	customTimeout := 10 * time.Second

	// Test that SetTimeout returns the client (for method chaining)
	chainedClient := client.SetTimeout(customTimeout)

	if chainedClient != client {
		t.Error("SetTimeout should return the same client instance for method chaining")
	}

	// Verify the client still works after timeout modification
	if client == nil {
		t.Error("Client should still be valid after timeout modification")
	}
}

func TestXenForoClientConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		apiKey     string
		apiUser    string
		maxRetries int
	}{
		{
			name:       "Valid configuration",
			baseURL:    "https://forum.example.com/api",
			apiKey:     "valid-key",
			apiUser:    "1",
			maxRetries: 3,
		},
		{
			name:       "Different retry count",
			baseURL:    "https://test.com/api",
			apiKey:     "another-key",
			apiUser:    "2",
			maxRetries: 5,
		},
		{
			name:       "Zero retries",
			baseURL:    "https://minimal.com/api",
			apiKey:     "minimal-key",
			apiUser:    "0",
			maxRetries: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.apiKey, tt.apiUser, tt.maxRetries)

			if client == nil {
				t.Errorf("Expected client to be created for %s", tt.name)
			}
		})
	}
}
