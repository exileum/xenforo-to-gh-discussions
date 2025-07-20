package xenforo

import (
	"strings"
)

// Thread represents a XenForo forum thread with metadata.
// Contains thread identification, authoring information, and reply statistics.
type Thread struct {
	ThreadID    int    `json:"thread_id"`     // Unique thread identifier
	Title       string `json:"title"`         // Thread title
	NodeID      int    `json:"node_id"`       // Parent forum/category ID
	Username    string `json:"username"`      // Thread author username
	PostDate    int64  `json:"post_date"`     // Creation timestamp (Unix)
	FirstPostID int    `json:"first_post_id"` // ID of the opening post
	ReplyCount  int    `json:"reply_count"`   // Number of replies
}

// IsValid validates the Thread struct and returns true if all required fields are valid.
func (t *Thread) IsValid() bool {
	return t.ThreadID > 0 &&
		len(strings.TrimSpace(t.Title)) > 0 &&
		len(strings.TrimSpace(t.Username)) > 0 &&
		t.PostDate >= 0
}

// Post represents an individual forum post within a thread.
// Includes content, authoring information, and optional file attachments.
type Post struct {
	PostID      int          `json:"post_id"`               // Unique post identifier
	ThreadID    int          `json:"thread_id"`             // Parent thread ID
	Username    string       `json:"username"`              // Post author username
	PostDate    int64        `json:"post_date"`             // Creation timestamp (Unix)
	Message     string       `json:"message"`               // Post content (BB-code formatted)
	Attachments []Attachment `json:"Attachments,omitempty"` // File attachments
}

// IsValid validates the Post struct and returns true if all required fields are valid.
func (p *Post) IsValid() bool {
	return p.PostID > 0 &&
		p.ThreadID > 0 &&
		len(strings.TrimSpace(p.Username)) > 0 &&
		p.PostDate >= 0 &&
		len(strings.TrimSpace(p.Message)) > 0
}

// Attachment represents a file attachment linked to a forum post.
// Contains download information and metadata for file migration.
type Attachment struct {
	AttachmentID int    `json:"attachment_id"` // Unique attachment identifier
	Filename     string `json:"filename"`      // Original filename
	DirectURL    string `json:"direct_url"`    // Download URL
}

// IsValid validates the Attachment struct and returns true if all required fields are valid.
// Includes security checks for path traversal attempts in filenames.
func (a *Attachment) IsValid() bool {
	return a.AttachmentID > 0 &&
		a.Filename != "" &&
		a.DirectURL != "" &&
		!strings.Contains(a.Filename, "..") && // Basic path traversal check
		!strings.Contains(a.Filename, "\\") && // Windows path traversal check
		(strings.HasPrefix(a.DirectURL, "http://") ||
			strings.HasPrefix(a.DirectURL, "https://"))
}

type ThreadsResponse struct {
	Threads    []Thread `json:"threads"`
	Pagination struct {
		CurrentPage int `json:"current_page"`
		TotalPages  int `json:"total_pages"`
	} `json:"pagination"`
}

type PostsResponse struct {
	Posts      []Post `json:"posts"`
	Pagination struct {
		CurrentPage int `json:"current_page"`
		TotalPages  int `json:"total_pages"`
	} `json:"pagination"`
}

// Node represents a XenForo forum node (category or forum).
// Contains hierarchical structure information and content statistics.
type Node struct {
	NodeID        int     `json:"node_id"`                    // Unique node identifier
	Title         string  `json:"title"`                      // Node display name
	NodeTypeID    string  `json:"node_type_id"`               // Node type (e.g., "Forum", "Category")
	Description   *string `json:"description,omitempty"`      // Optional description
	ParentNodeID  int     `json:"parent_node_id"`             // Parent node ID (0 for root)
	DisplayOrder  int     `json:"display_order"`              // Sort order
	DisplayInList bool    `json:"display_in_list"`            // Visibility flag
	ThreadCount   *int    `json:"discussion_count,omitempty"` // Thread count for forums
}

// IsValid validates the Node struct and returns true if all required fields are valid.
func (n *Node) IsValid() bool {
	return n.NodeID > 0 &&
		len(strings.TrimSpace(n.Title)) > 0 &&
		len(strings.TrimSpace(n.NodeTypeID)) > 0
}

type NodesResponse struct {
	Nodes []Node `json:"nodes"`
}
