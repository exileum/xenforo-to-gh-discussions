package mocks

import (
	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

type XenForoClient struct {
	TestConnectionFunc     func() error
	GetThreadsFunc         func(nodeID int) ([]xenforo.Thread, error)
	GetPostsFunc           func(threadID int) ([]xenforo.Post, error)
	GetAttachmentsFunc     func(threadID int) ([]xenforo.Attachment, error)
	DownloadAttachmentFunc func(url, filepath string) error
}

func (m *XenForoClient) TestConnection() error {
	if m.TestConnectionFunc != nil {
		return m.TestConnectionFunc()
	}
	return nil
}

func (m *XenForoClient) GetThreads(nodeID int) ([]xenforo.Thread, error) {
	if m.GetThreadsFunc != nil {
		return m.GetThreadsFunc(nodeID)
	}
	return []xenforo.Thread{
		{
			ThreadID:    1,
			Title:       "Test Thread",
			NodeID:      nodeID,
			Username:    "testuser",
			PostDate:    1642353000,
			FirstPostID: 1,
		},
	}, nil
}

func (m *XenForoClient) GetPosts(threadID int) ([]xenforo.Post, error) {
	if m.GetPostsFunc != nil {
		return m.GetPostsFunc(threadID)
	}
	return []xenforo.Post{
		{
			PostID:   1,
			ThreadID: threadID,
			Username: "testuser",
			PostDate: 1642353000,
			Message:  "This is a [b]test post[/b]",
		},
	}, nil
}

func (m *XenForoClient) GetAttachments(threadID int) ([]xenforo.Attachment, error) {
	if m.GetAttachmentsFunc != nil {
		return m.GetAttachmentsFunc(threadID)
	}
	return []xenforo.Attachment{
		{
			AttachmentID: 1,
			Filename:     "test.png",
			ViewURL:      "https://example.com/test.png",
		},
	}, nil
}

func (m *XenForoClient) DownloadAttachment(url, filepath string) error {
	if m.DownloadAttachmentFunc != nil {
		return m.DownloadAttachmentFunc(url, filepath)
	}
	return nil
}
