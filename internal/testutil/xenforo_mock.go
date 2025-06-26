package testutil

import (
	"errors"

	"github.com/exileum/xenforo-to-gh-discussions/internal/xenforo"
)

type XenForoClient struct {
	TestConnectionFunc     func() error
	GetThreadsFunc         func(nodeID int) ([]xenforo.Thread, error)
	GetPostsFunc           func(thread xenforo.Thread) ([]xenforo.Post, error)
	DownloadAttachmentFunc func(url, filepath string) error
}

func (m *XenForoClient) TestConnection() error {
	if m.TestConnectionFunc != nil {
		return m.TestConnectionFunc()
	}
	return errors.New("TestConnectionFunc not set - test must explicitly set mock behavior")
}

func (m *XenForoClient) GetThreads(nodeID int) ([]xenforo.Thread, error) {
	if m.GetThreadsFunc != nil {
		return m.GetThreadsFunc(nodeID)
	}
	return nil, errors.New("GetThreadsFunc not set - test must explicitly set mock behavior")
}

func (m *XenForoClient) GetPosts(thread xenforo.Thread) ([]xenforo.Post, error) {
	if m.GetPostsFunc != nil {
		return m.GetPostsFunc(thread)
	}
	return nil, errors.New("GetPostsFunc not set - test must explicitly set mock behavior")
}

func (m *XenForoClient) DownloadAttachment(url, filepath string) error {
	if m.DownloadAttachmentFunc != nil {
		return m.DownloadAttachmentFunc(url, filepath)
	}
	return errors.New("DownloadAttachmentFunc not set - test must explicitly set mock behavior")
}
