package testutil

import (
	"github.com/exileum/xenforo-to-gh-discussions/internal/github"
)

type GitHubClient struct {
	CreateDiscussionFunc  func(title, body, categoryID string) (*github.DiscussionResult, error)
	AddCommentFunc        func(discussionID, body string) error
	GetRepositoryInfoFunc func(repo string) (*github.RepositoryInfo, error)
}

func (m *GitHubClient) CreateDiscussion(title, body, categoryID string) (*github.DiscussionResult, error) {
	if m.CreateDiscussionFunc != nil {
		return m.CreateDiscussionFunc(title, body, categoryID)
	}
	return &github.DiscussionResult{ID: "test_id", Number: 1}, nil
}

func (m *GitHubClient) AddComment(discussionID, body string) error {
	if m.AddCommentFunc != nil {
		return m.AddCommentFunc(discussionID, body)
	}
	return nil
}

func (m *GitHubClient) GetRepositoryInfo(repo string) (*github.RepositoryInfo, error) {
	if m.GetRepositoryInfoFunc != nil {
		return m.GetRepositoryInfoFunc(repo)
	}
	return &github.RepositoryInfo{
		ID:                 "test_repo_id",
		DiscussionsEnabled: true,
		DiscussionCategories: []github.Category{
			{ID: "DIC_kwDOtest123", Name: "General"},
		},
	}, nil
}

func (m *GitHubClient) SetRepositoryID(id string) {
	// Mock implementation
}

func (m *GitHubClient) GetRepositoryID() string {
	return "test_repo_id"
}
