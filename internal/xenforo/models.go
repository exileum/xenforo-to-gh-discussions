package xenforo

type Thread struct {
	ThreadID    int    `json:"thread_id"`
	Title       string `json:"title"`
	NodeID      int    `json:"node_id"`
	Username    string `json:"username"`
	PostDate    int64  `json:"post_date"`
	FirstPostID int    `json:"first_post_id"`
}

type Post struct {
	PostID   int    `json:"post_id"`
	ThreadID int    `json:"thread_id"`
	Username string `json:"username"`
	PostDate int64  `json:"post_date"`
	Message  string `json:"message"`
}

type Attachment struct {
	AttachmentID int    `json:"attachment_id"`
	Filename     string `json:"filename"`
	ViewURL      string `json:"view_url"`
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

type AttachmentsResponse struct {
	Attachments []Attachment `json:"attachments"`
}

type Node struct {
	NodeID        int     `json:"node_id"`
	Title         string  `json:"title"`
	NodeTypeID    string  `json:"node_type_id"`
	Description   *string `json:"description,omitempty"`
	ParentNodeID  int     `json:"parent_node_id"`
	DisplayOrder  int     `json:"display_order"`
	DisplayInList bool    `json:"display_in_list"`
	ThreadCount   *int    `json:"discussion_count,omitempty"` // For forum nodes
}

type NodesResponse struct {
	Nodes []Node `json:"nodes"`
}
