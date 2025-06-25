package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

// TestRetryableRequest tests the retry logic with exponential backoff
func TestRetryableRequest(t *testing.T) {
	tests := []struct {
		name          string
		responses     []int         // Status codes to return
		expectedCalls int           // Expected number of API calls
		expectError   bool          // Whether we expect an error
		minDuration   time.Duration // Minimum expected duration (for backoff)
	}{
		{
			name:          "Successful on first try",
			responses:     []int{200},
			expectedCalls: 1,
			expectError:   false,
			minDuration:   0,
		},
		{
			name:          "Rate limited then success",
			responses:     []int{429, 200},
			expectedCalls: 2,
			expectError:   false,
			minDuration:   1 * time.Second, // First retry after 1 second
		},
		{
			name:          "Multiple rate limits then success",
			responses:     []int{429, 429, 200},
			expectedCalls: 3,
			expectError:   false,
			minDuration:   3 * time.Second, // 1s + 2s backoff
		},
		{
			name:          "Max retries exceeded",
			responses:     []int{429, 429, 429, 429}, // More than MaxRetries
			expectedCalls: MaxRetries,
			expectError:   true,
			minDuration:   3 * time.Second, // 1s + 2s backoff (no 3rd retry)
		},
		{
			name:          "Non-429 error",
			responses:     []int{500},
			expectedCalls: 1,
			expectError:   false, // Returns response even with error status
			minDuration:   0,
		},
		{
			name:          "404 Not Found",
			responses:     []int{404},
			expectedCalls: 1,
			expectError:   false,
			minDuration:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			responseIndex := 0

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if responseIndex < len(tt.responses) {
					w.WriteHeader(tt.responses[responseIndex])
					responseIndex++
				} else {
					w.WriteHeader(200)
				}
				w.Write([]byte(`{"status": "ok"}`))
			}))
			defer server.Close()

			// Create test client
			testClient := resty.New()

			start := time.Now()

			// Make request with retry logic
			resp, err := retryableRequest(func() (*resty.Response, error) {
				return testClient.R().Get(server.URL)
			})

			duration := time.Since(start)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check number of calls
			if callCount != tt.expectedCalls {
				t.Errorf("Expected %d calls, got %d", tt.expectedCalls, callCount)
			}

			// Check response
			if !tt.expectError && resp != nil {
				lastStatusIndex := len(tt.responses) - 1
				if lastStatusIndex >= 0 && responseIndex > 0 {
					expectedStatus := tt.responses[responseIndex-1]
					if resp.StatusCode() != expectedStatus {
						t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode())
					}
				}
			}

			// Check minimum duration for backoff
			if duration < tt.minDuration {
				t.Errorf("Expected minimum duration %v, but completed in %v", tt.minDuration, duration)
			}
		})
	}
}

// TestPreflightChecks tests the pre-flight validation
func TestPreflightChecks(t *testing.T) {
	// Save original values
	originalDryRun := dryRun
	originalClient := client
	originalAPIKey := XenForoAPIKey
	defer func() {
		dryRun = originalDryRun
		client = originalClient
		XenForoAPIKey = originalAPIKey
	}()

	tests := []struct {
		name        string
		setupFunc   func(*httptest.Server)
		expectError bool
		errorMsg    string
	}{
		{
			name: "Dry run mode - should pass",
			setupFunc: func(server *httptest.Server) {
				dryRun = true
			},
			expectError: false,
		},
		{
			name: "XenForo API authentication failed",
			setupFunc: func(server *httptest.Server) {
				dryRun = false
				// Create client without API key to trigger 401
				client = resty.New()
				// Don't set XF-Api-Key header to trigger authentication failure
				XenForoAPIKey = ""
				// Set githubClient to nil to avoid nil pointer dereference
				githubClient = nil
			},
			expectError: true,
			errorMsg:    "authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check XenForo API headers
				if r.Header.Get("XF-Api-Key") == "" {
					w.WriteHeader(401)
					return
				}
				w.WriteHeader(200)
				w.Write([]byte(`{"api": {"version": "1.0"}}`))
			}))
			defer server.Close()

			// Override API URL for testing
			oldURL := XenForoAPIURL
			defer func() {
				XenForoAPIURL = oldURL
			}()
			XenForoAPIURL = server.URL

			tt.setupFunc(server)

			err := runPreflightChecks()

			if tt.expectError && err == nil {
				t.Errorf("Expected error containing '%s' but got none", tt.errorMsg)
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
			}
		})
	}
}

// MockXenForoAPI creates a mock XenForo API server for testing
func MockXenForoAPI(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authentication
		if r.Header.Get("XF-Api-Key") != "test-key" {
			w.WriteHeader(401)
			w.Write([]byte(`{"errors": [{"code": "unauthorized"}]}`))
			return
		}

		switch r.URL.Path {
		case "/threads":
			page := r.URL.Query().Get("page")
			if page == "" || page == "1" {
				response := map[string]interface{}{
					"threads": []XenForoThread{
						{ThreadID: 1, Title: "Test Thread 1", NodeID: 1, Username: "user1", PostDate: time.Now().Unix()},
						{ThreadID: 2, Title: "Test Thread 2", NodeID: 1, Username: "user2", PostDate: time.Now().Unix()},
					},
					"pagination": map[string]int{
						"current_page": 1,
						"total_pages":  1,
					},
				}
				json.NewEncoder(w).Encode(response)
			}

		case "/threads/1/posts":
			response := map[string]interface{}{
				"posts": []XenForoPost{
					{PostID: 1, ThreadID: 1, Username: "user1", PostDate: time.Now().Unix(), Message: "[b]First post[/b]"},
					{PostID: 2, ThreadID: 1, Username: "user2", PostDate: time.Now().Unix(), Message: "Reply with [i]italic[/i]"},
				},
				"pagination": map[string]int{
					"current_page": 1,
					"total_pages":  1,
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/threads/1/attachments":
			response := map[string]interface{}{
				"attachments": []XenForoAttachment{
					{AttachmentID: 1, Filename: "test.png", ViewURL: fmt.Sprintf("%s/attachments/1", r.Host)},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/attachments/1":
			// Serve a mock image
			w.Header().Set("Content-Type", "image/png")
			w.Write([]byte("PNG_DATA"))

		default:
			w.WriteHeader(404)
		}
	}))
}


// TestAPIErrorHandling tests various API error scenarios
func TestAPIErrorHandling(t *testing.T) {
	// Save original values
	originalClient := client
	originalAPIURL := XenForoAPIURL
	defer func() {
		client = originalClient
		XenForoAPIURL = originalAPIURL
	}()

	tests := []struct {
		name          string
		endpoint      string
		statusCode    int
		responseBody  string
		expectError   bool
		errorContains string
	}{
		{
			name:          "XenForo 404",
			endpoint:      "/threads",
			statusCode:    404,
			responseBody:  `{"errors": [{"code": "not_found"}]}`,
			expectError:   true,
			errorContains: "XenForo API error",
		},
		{
			name:          "XenForo 500",
			endpoint:      "/threads",
			statusCode:    500,
			responseBody:  `{"errors": [{"code": "server_error"}]}`,
			expectError:   true,
			errorContains: "XenForo API error",
		},
		{
			name:          "Invalid JSON response",
			endpoint:      "/threads",
			statusCode:    200,
			responseBody:  `{invalid json`,
			expectError:   true,
			errorContains: "",
		},
		{
			name:         "Empty response",
			endpoint:     "/threads",
			statusCode:   200,
			responseBody: `{"threads": [], "pagination": {"current_page": 1, "total_pages": 1}}`,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Override API URL to point to mock server
			XenForoAPIURL = server.URL
			client = resty.New()

			// Test based on endpoint
			var err error
			switch tt.endpoint {
			case "/threads":
				_, err = getXenForoThreads(1)
			case "/posts":
				_, err = getXenForoPosts(1)
			case "/attachments":
				_, err = getXenForoAttachments(1)
			}

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if tt.expectError && tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
			}
		})
	}
}
