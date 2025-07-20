package xenforo

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

// Attachment represents a file attachment linked to a forum post.
// Contains download information and metadata for file migration.
type Attachment struct {
	AttachmentID int    `json:"attachment_id"` // Unique attachment identifier
	Filename     string `json:"filename"`      // Original filename
	DirectURL    string `json:"direct_url"`    // Download URL
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

type NodesResponse struct {
	Nodes []Node `json:"nodes"`
}
