package github

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
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
			token:     "test_token_1234567890_fake_github_pat",
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
			token:     "test_minimal_token_12345", // 24 chars, above minimum
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.token, 1*time.Second, 3, 2)

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
	client, err := NewClient("test_github_token_for_testing_only", 1*time.Second, 3, 2)
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

func TestNewClientParameterValidation(t *testing.T) {
	tests := []struct {
		name                 string
		token                string
		rateLimitDelay       time.Duration
		maxRetries           int
		retryBackoffMultiple int
		shouldErr            bool
		errMsg               string
	}{
		{
			name:                 "Valid parameters",
			token:                "test_token_1234567890_fake_github_pat",
			rateLimitDelay:       1 * time.Second,
			maxRetries:           3,
			retryBackoffMultiple: 2,
			shouldErr:            false,
		},
		{
			name:                 "Negative rate limit delay",
			token:                "test_token_1234567890_fake_github_pat",
			rateLimitDelay:       -1 * time.Second,
			maxRetries:           3,
			retryBackoffMultiple: 2,
			shouldErr:            true,
			errMsg:               "rate limit delay cannot be negative",
		},
		{
			name:                 "Negative max retries",
			token:                "test_token_1234567890_fake_github_pat",
			rateLimitDelay:       1 * time.Second,
			maxRetries:           -1,
			retryBackoffMultiple: 2,
			shouldErr:            true,
			errMsg:               "max retries cannot be negative",
		},
		{
			name:                 "Zero retry backoff multiple",
			token:                "test_token_1234567890_fake_github_pat",
			rateLimitDelay:       1 * time.Second,
			maxRetries:           3,
			retryBackoffMultiple: 0,
			shouldErr:            true,
			errMsg:               "retry backoff multiple must be at least 1",
		},
		{
			name:                 "Zero values (valid)",
			token:                "test_token_1234567890_fake_github_pat",
			rateLimitDelay:       0,
			maxRetries:           0,
			retryBackoffMultiple: 1,
			shouldErr:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.token, tt.rateLimitDelay, tt.maxRetries, tt.retryBackoffMultiple)

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
					return
				}

				// Verify parameters were set correctly
				if client.rateLimitDelay != tt.rateLimitDelay {
					t.Errorf("Expected rate limit delay %v, got %v", tt.rateLimitDelay, client.rateLimitDelay)
				}
				if client.maxRetries != tt.maxRetries {
					t.Errorf("Expected max retries %d, got %d", tt.maxRetries, client.maxRetries)
				}
				if client.retryBackoffMultiple != tt.retryBackoffMultiple {
					t.Errorf("Expected retry backoff multiple %d, got %d", tt.retryBackoffMultiple, client.retryBackoffMultiple)
				}
			}
		})
	}
}

func TestClient_GetStats(t *testing.T) {
	client, err := NewClient("test_github_token_for_testing_only", 1*time.Second, 3, 2)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test initial stats
	opCount, rateLimitHits := client.GetStats()
	if opCount != 0 {
		t.Errorf("Expected 0 operations initially, got %d", opCount)
	}
	if rateLimitHits != 0 {
		t.Errorf("Expected 0 rate limit hits initially, got %d", rateLimitHits)
	}
}

func TestRateLimitError(t *testing.T) {
	resetTime := time.Now().Add(1 * time.Hour)
	rateLimitErr := &RateLimitError{
		ResetTime: resetTime,
		Remaining: 0,
		Message:   "API rate limit exceeded",
	}

	errorMsg := rateLimitErr.Error()
	expectedSubstrings := []string{
		"GitHub API rate limit exceeded",
		"API rate limit exceeded",
		"remaining: 0",
		resetTime.Format(time.RFC3339),
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(errorMsg, substr) {
			t.Errorf("Expected error message to contain %q, got %q", substr, errorMsg)
		}
	}
}

func TestClient_parseRateLimitFromError(t *testing.T) {
	client, err := NewClient("test_github_token_for_testing_only", 1*time.Second, 3, 2)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name            string
		inputError      error
		expectRateLimit bool
		expectedMessage string
	}{
		{
			name:            "Nil error",
			inputError:      nil,
			expectRateLimit: false,
		},
		{
			name:            "Rate limit error",
			inputError:      errors.New("API rate limit exceeded"),
			expectRateLimit: true,
			expectedMessage: "API rate limit exceeded",
		},
		{
			name:            "Secondary rate limit error",
			inputError:      errors.New("You have been rate limited for creating content"),
			expectRateLimit: true,
			expectedMessage: "You have been rate limited for creating content",
		},
		{
			name:            "Abuse detection error",
			inputError:      errors.New("Request blocked by abuse detection"),
			expectRateLimit: true,
			expectedMessage: "Request blocked by abuse detection",
		},
		{
			name:            "Generic error",
			inputError:      errors.New("Network connection failed"),
			expectRateLimit: false,
		},
		{
			name:            "Authentication error",
			inputError:      errors.New("Bad credentials"),
			expectRateLimit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rateLimitErr, isRateLimit := client.parseRateLimitFromError(tt.inputError)

			if isRateLimit != tt.expectRateLimit {
				t.Errorf("Expected rate limit detection %v, got %v", tt.expectRateLimit, isRateLimit)
			}

			if tt.expectRateLimit {
				if rateLimitErr == nil {
					t.Error("Expected RateLimitError but got nil")
					return
				}
				if rateLimitErr.Message != tt.expectedMessage {
					t.Errorf("Expected message %q, got %q", tt.expectedMessage, rateLimitErr.Message)
				}
				if rateLimitErr.ResetTime.Before(time.Now()) {
					t.Error("Expected reset time to be in the future")
				}
			} else {
				if rateLimitErr != nil {
					t.Errorf("Expected nil RateLimitError but got %v", rateLimitErr)
				}
			}
		})
	}
}

func TestClient_isRetryableError(t *testing.T) {
	client, err := NewClient("test_github_token_for_testing_only", 1*time.Second, 3, 2)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name     string
		error    error
		expected bool
	}{
		{
			name:     "Nil error",
			error:    nil,
			expected: false,
		},
		{
			name:     "Connection reset error",
			error:    errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "Timeout error",
			error:    errors.New("request timeout"),
			expected: true,
		},
		{
			name:     "Server error",
			error:    errors.New("internal server error"),
			expected: true,
		},
		{
			name:     "502 Bad Gateway",
			error:    errors.New("502 bad gateway"),
			expected: true,
		},
		{
			name:     "503 Service Unavailable",
			error:    errors.New("503 service unavailable"),
			expected: true,
		},
		{
			name:     "504 Gateway Timeout",
			error:    errors.New("504 gateway timeout"),
			expected: true,
		},
		{
			name:     "Unauthorized error",
			error:    errors.New("401 unauthorized"),
			expected: false,
		},
		{
			name:     "Forbidden error",
			error:    errors.New("403 forbidden"),
			expected: false,
		},
		{
			name:     "Not found error",
			error:    errors.New("404 not found"),
			expected: false,
		},
		{
			name:     "Bad request error",
			error:    errors.New("400 bad request"),
			expected: false,
		},
		{
			name:     "Invalid input error",
			error:    errors.New("invalid input provided"),
			expected: false,
		},
		{
			name:     "Unknown error (default retryable)",
			error:    errors.New("some random error message"),
			expected: true, // Unknown errors are considered retryable by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.isRetryableError(tt.error)
			if result != tt.expected {
				t.Errorf("Expected %v for error %q, got %v", tt.expected, tt.error, result)
			}
		})
	}
}

func TestClient_executeWithRetryContextCancellation(t *testing.T) {
	client, err := NewClient("test_github_token_for_testing_only", 100*time.Millisecond, 3, 2)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test context cancellation before operation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = client.executeWithRetry(ctx, func() error {
		return errors.New("test error")
	})

	if err == nil {
		t.Error("Expected error with cancelled context")
	}
	if !strings.Contains(err.Error(), "operation cancelled") {
		t.Errorf("Expected cancellation error, got: %v", err)
	}
}

func TestClient_executeWithRetrySuccess(t *testing.T) {
	client, err := NewClient("test_github_token_for_testing_only", 1*time.Millisecond, 3, 2)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	callCount := 0

	err = client.executeWithRetry(ctx, func() error {
		callCount++
		if callCount < 2 {
			return errors.New("temporary failure")
		}
		return nil // Success on second try
	})

	if err != nil {
		t.Errorf("Expected success after retry, got: %v", err)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestClient_executeWithRetryMaxRetries(t *testing.T) {
	client, err := NewClient("test_github_token_for_testing_only", 1*time.Millisecond, 2, 2)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	callCount := 0

	err = client.executeWithRetry(ctx, func() error {
		callCount++
		return errors.New("persistent failure")
	})

	if err == nil {
		t.Error("Expected error after max retries")
	}
	if !strings.Contains(err.Error(), "operation failed after") && !strings.Contains(err.Error(), "persistent failure") {
		t.Errorf("Expected max retries error, got: %v", err)
	}
	// Should be called 3 times: initial + 2 retries
	if callCount != 3 {
		t.Errorf("Expected 3 calls (1 initial + 2 retries), got %d", callCount)
	}
}
