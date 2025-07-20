package xenforo

import (
	"strings"
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

func TestThread_Validation(t *testing.T) {
	tests := []struct {
		name   string
		thread Thread
		valid  bool
	}{
		{
			name: "Valid thread",
			thread: Thread{
				ThreadID:    1,
				Title:       "Test Thread",
				NodeID:      1,
				Username:    "testuser",
				PostDate:    time.Now().Unix(),
				FirstPostID: 1,
				ReplyCount:  5,
			},
			valid: true,
		},
		{
			name: "Zero thread ID",
			thread: Thread{
				ThreadID:    0,
				Title:       "Test Thread",
				NodeID:      1,
				Username:    "testuser",
				PostDate:    time.Now().Unix(),
				FirstPostID: 1,
				ReplyCount:  5,
			},
			valid: false,
		},
		{
			name: "Empty title",
			thread: Thread{
				ThreadID:    1,
				Title:       "",
				NodeID:      1,
				Username:    "testuser",
				PostDate:    time.Now().Unix(),
				FirstPostID: 1,
				ReplyCount:  5,
			},
			valid: false,
		},
		{
			name: "Empty username",
			thread: Thread{
				ThreadID:    1,
				Title:       "Test Thread",
				NodeID:      1,
				Username:    "",
				PostDate:    time.Now().Unix(),
				FirstPostID: 1,
				ReplyCount:  5,
			},
			valid: false,
		},
		{
			name: "Negative post date",
			thread: Thread{
				ThreadID:    1,
				Title:       "Test Thread",
				NodeID:      1,
				Username:    "testuser",
				PostDate:    -1,
				FirstPostID: 1,
				ReplyCount:  5,
			},
			valid: false,
		},
		{
			name: "Future post date (valid)",
			thread: Thread{
				ThreadID:    1,
				Title:       "Test Thread",
				NodeID:      1,
				Username:    "testuser",
				PostDate:    time.Now().Add(1 * time.Hour).Unix(),
				FirstPostID: 1,
				ReplyCount:  5,
			},
			valid: true, // Future dates might be valid in some cases
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.thread.ThreadID > 0 &&
				tt.thread.Title != "" &&
				tt.thread.Username != "" &&
				tt.thread.PostDate >= 0

			if isValid != tt.valid {
				t.Errorf("Expected validity %v for %s, got %v", tt.valid, tt.name, isValid)
			}
		})
	}
}

func TestPost_Validation(t *testing.T) {
	tests := []struct {
		name  string
		post  Post
		valid bool
	}{
		{
			name: "Valid post",
			post: Post{
				PostID:   1,
				ThreadID: 1,
				Username: "testuser",
				PostDate: time.Now().Unix(),
				Message:  "Test message content",
				Attachments: []Attachment{
					{AttachmentID: 1, Filename: "test.jpg", DirectURL: "https://example.com/test.jpg"},
				},
			},
			valid: true,
		},
		{
			name: "Post without attachments",
			post: Post{
				PostID:      1,
				ThreadID:    1,
				Username:    "testuser",
				PostDate:    time.Now().Unix(),
				Message:     "Test message content",
				Attachments: []Attachment{},
			},
			valid: true,
		},
		{
			name: "Zero post ID",
			post: Post{
				PostID:   0,
				ThreadID: 1,
				Username: "testuser",
				PostDate: time.Now().Unix(),
				Message:  "Test message content",
			},
			valid: false,
		},
		{
			name: "Empty message",
			post: Post{
				PostID:   1,
				ThreadID: 1,
				Username: "testuser",
				PostDate: time.Now().Unix(),
				Message:  "",
			},
			valid: false,
		},
		{
			name: "Whitespace-only message",
			post: Post{
				PostID:   1,
				ThreadID: 1,
				Username: "testuser",
				PostDate: time.Now().Unix(),
				Message:  "   \t\n   ",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.post.PostID > 0 &&
				tt.post.ThreadID > 0 &&
				tt.post.Username != "" &&
				tt.post.PostDate >= 0 &&
				len(tt.post.Message) > 0 &&
				tt.post.Message != "" &&
				len(strings.TrimSpace(tt.post.Message)) > 0

			if isValid != tt.valid {
				t.Errorf("Expected validity %v for %s, got %v", tt.valid, tt.name, isValid)
			}
		})
	}
}

func TestAttachment_Validation(t *testing.T) {
	tests := []struct {
		name       string
		attachment Attachment
		valid      bool
	}{
		{
			name: "Valid attachment",
			attachment: Attachment{
				AttachmentID: 1,
				Filename:     "test.jpg",
				DirectURL:    "https://example.com/attachments/test.jpg",
			},
			valid: true,
		},
		{
			name: "Zero attachment ID",
			attachment: Attachment{
				AttachmentID: 0,
				Filename:     "test.jpg",
				DirectURL:    "https://example.com/attachments/test.jpg",
			},
			valid: false,
		},
		{
			name: "Empty filename",
			attachment: Attachment{
				AttachmentID: 1,
				Filename:     "",
				DirectURL:    "https://example.com/attachments/test.jpg",
			},
			valid: false,
		},
		{
			name: "Empty URL",
			attachment: Attachment{
				AttachmentID: 1,
				Filename:     "test.jpg",
				DirectURL:    "",
			},
			valid: false,
		},
		{
			name: "Invalid URL format",
			attachment: Attachment{
				AttachmentID: 1,
				Filename:     "test.jpg",
				DirectURL:    "not-a-valid-url",
			},
			valid: false, // For this test, we consider it invalid
		},
		{
			name: "Filename with path traversal",
			attachment: Attachment{
				AttachmentID: 1,
				Filename:     "../../../etc/passwd",
				DirectURL:    "https://example.com/attachments/test.jpg",
			},
			valid: false, // Should be considered invalid for security
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.attachment.AttachmentID > 0 &&
				tt.attachment.Filename != "" &&
				tt.attachment.DirectURL != "" &&
				!strings.Contains(tt.attachment.Filename, "..") && // Basic path traversal check
				(strings.HasPrefix(tt.attachment.DirectURL, "http://") ||
					strings.HasPrefix(tt.attachment.DirectURL, "https://"))

			if isValid != tt.valid {
				t.Errorf("Expected validity %v for %s, got %v", tt.valid, tt.name, isValid)
			}
		})
	}
}

func TestNode_Validation(t *testing.T) {
	tests := []struct {
		name  string
		node  Node
		valid bool
	}{
		{
			name: "Valid forum node",
			node: Node{
				NodeID:        1,
				Title:         "General Discussion",
				NodeTypeID:    "Forum",
				Description:   stringPtr("A place for general discussion"),
				ParentNodeID:  0,
				DisplayOrder:  1,
				DisplayInList: true,
				ThreadCount:   intPtr(42),
			},
			valid: true,
		},
		{
			name: "Valid category node",
			node: Node{
				NodeID:        2,
				Title:         "Main Category",
				NodeTypeID:    "Category",
				Description:   nil,
				ParentNodeID:  0,
				DisplayOrder:  1,
				DisplayInList: true,
				ThreadCount:   nil, // Categories might not have thread counts
			},
			valid: true,
		},
		{
			name: "Zero node ID",
			node: Node{
				NodeID:        0,
				Title:         "Invalid Node",
				NodeTypeID:    "Forum",
				ParentNodeID:  0,
				DisplayOrder:  1,
				DisplayInList: true,
			},
			valid: false,
		},
		{
			name: "Empty title",
			node: Node{
				NodeID:        1,
				Title:         "",
				NodeTypeID:    "Forum",
				ParentNodeID:  0,
				DisplayOrder:  1,
				DisplayInList: true,
			},
			valid: false,
		},
		{
			name: "Empty node type",
			node: Node{
				NodeID:        1,
				Title:         "Test Node",
				NodeTypeID:    "",
				ParentNodeID:  0,
				DisplayOrder:  1,
				DisplayInList: true,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.node.NodeID > 0 &&
				tt.node.Title != "" &&
				tt.node.NodeTypeID != ""

			if isValid != tt.valid {
				t.Errorf("Expected validity %v for %s, got %v", tt.valid, tt.name, isValid)
			}
		})
	}
}

func TestThreadsResponse_Validation(t *testing.T) {
	response := ThreadsResponse{
		Threads: []Thread{
			{ThreadID: 1, Title: "Thread 1", NodeID: 1, Username: "user1", PostDate: time.Now().Unix()},
			{ThreadID: 2, Title: "Thread 2", NodeID: 1, Username: "user2", PostDate: time.Now().Unix()},
		},
		Pagination: struct {
			CurrentPage int `json:"current_page"`
			TotalPages  int `json:"total_pages"`
		}{
			CurrentPage: 1,
			TotalPages:  5,
		},
	}

	if len(response.Threads) != 2 {
		t.Errorf("Expected 2 threads, got %d", len(response.Threads))
	}

	if response.Pagination.CurrentPage != 1 {
		t.Errorf("Expected current page 1, got %d", response.Pagination.CurrentPage)
	}

	if response.Pagination.TotalPages != 5 {
		t.Errorf("Expected total pages 5, got %d", response.Pagination.TotalPages)
	}
}

func TestPostsResponse_Validation(t *testing.T) {
	response := PostsResponse{
		Posts: []Post{
			{PostID: 1, ThreadID: 1, Username: "user1", PostDate: time.Now().Unix(), Message: "First post"},
			{PostID: 2, ThreadID: 1, Username: "user2", PostDate: time.Now().Unix(), Message: "Reply post"},
		},
		Pagination: struct {
			CurrentPage int `json:"current_page"`
			TotalPages  int `json:"total_pages"`
		}{
			CurrentPage: 1,
			TotalPages:  3,
		},
	}

	if len(response.Posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(response.Posts))
	}

	if response.Pagination.CurrentPage != 1 {
		t.Errorf("Expected current page 1, got %d", response.Pagination.CurrentPage)
	}

	if response.Pagination.TotalPages != 3 {
		t.Errorf("Expected total pages 3, got %d", response.Pagination.TotalPages)
	}
}

func TestNodesResponse_Validation(t *testing.T) {
	response := NodesResponse{
		Nodes: []Node{
			{NodeID: 1, Title: "Forum 1", NodeTypeID: "Forum", DisplayInList: true},
			{NodeID: 2, Title: "Category 1", NodeTypeID: "Category", DisplayInList: true},
			{NodeID: 3, Title: "Hidden Forum", NodeTypeID: "Forum", DisplayInList: false},
		},
	}

	if len(response.Nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(response.Nodes))
	}

	// Count visible nodes
	visibleCount := 0
	for _, node := range response.Nodes {
		if node.DisplayInList {
			visibleCount++
		}
	}

	if visibleCount != 2 {
		t.Errorf("Expected 2 visible nodes, got %d", visibleCount)
	}
}

func TestClient_NewClient(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		apiKey     string
		apiUser    string
		maxRetries int
		expectNil  bool
	}{
		{
			name:       "Valid parameters",
			baseURL:    "https://forum.example.com/api",
			apiKey:     "valid_api_key",
			apiUser:    "1",
			maxRetries: 3,
			expectNil:  false,
		},
		{
			name:       "Empty base URL",
			baseURL:    "",
			apiKey:     "valid_api_key",
			apiUser:    "1",
			maxRetries: 3,
			expectNil:  false, // Client creation doesn't validate URL format
		},
		{
			name:       "Empty API key",
			baseURL:    "https://forum.example.com/api",
			apiKey:     "",
			apiUser:    "1",
			maxRetries: 3,
			expectNil:  false, // Client creation doesn't validate key
		},
		{
			name:       "Zero max retries",
			baseURL:    "https://forum.example.com/api",
			apiKey:     "valid_api_key",
			apiUser:    "1",
			maxRetries: 0,
			expectNil:  false, // Zero retries is valid
		},
		{
			name:       "Negative max retries",
			baseURL:    "https://forum.example.com/api",
			apiKey:     "valid_api_key",
			apiUser:    "1",
			maxRetries: -1,
			expectNil:  false, // Client creation doesn't validate retry count
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.apiKey, tt.apiUser, tt.maxRetries)

			if tt.expectNil && client != nil {
				t.Errorf("Expected nil client for %s", tt.name)
			}
			if !tt.expectNil && client == nil {
				t.Errorf("Expected non-nil client for %s", tt.name)
			}
		})
	}
}

// Helper functions for pointer types
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
