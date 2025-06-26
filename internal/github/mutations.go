package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
)

type DiscussionResult struct {
	ID     string
	Number int
}

func (c *Client) CreateDiscussion(title, body, categoryID string) (*DiscussionResult, error) {
	// Input validation
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("discussion title cannot be empty")
	}
	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("discussion body cannot be empty")
	}
	if strings.TrimSpace(categoryID) == "" {
		return nil, fmt.Errorf("categoryID cannot be empty")
	}

	var result *DiscussionResult

	err := c.executeWithRetry(func() error {
		var mutation struct {
			CreateDiscussion struct {
				Discussion struct {
					ID     string
					Number int
				}
			} `graphql:"createDiscussion(input: $input)"`
		}

		input := githubv4.CreateDiscussionInput{
			RepositoryID: githubv4.ID(c.repositoryID),
			Title:        githubv4.String(title),
			Body:         githubv4.String(body),
			CategoryID:   githubv4.ID(categoryID),
		}

		err := c.client.Mutate(context.Background(), &mutation, input, nil)
		if err != nil {
			return fmt.Errorf("failed to create discussion %q in category %q: %w", title, categoryID, err)
		}

		result = &DiscussionResult{
			ID:     mutation.CreateDiscussion.Discussion.ID,
			Number: mutation.CreateDiscussion.Discussion.Number,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) AddComment(discussionID, body string) error {
	// Input validation
	if strings.TrimSpace(discussionID) == "" {
		return fmt.Errorf("discussionID cannot be empty")
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("comment body cannot be empty")
	}

	return c.executeWithRetry(func() error {
		var mutation struct {
			AddDiscussionComment struct {
				Comment struct {
					ID githubv4.ID
				}
			} `graphql:"addDiscussionComment(input: $input)"`
		}

		input := githubv4.AddDiscussionCommentInput{
			DiscussionID: githubv4.ID(discussionID),
			Body:         githubv4.String(body),
		}

		err := c.client.Mutate(context.Background(), &mutation, input, nil)
		if err != nil {
			return fmt.Errorf("failed to add comment to discussion %q: %w", discussionID, err)
		}

		return nil
	})
}
