package github

import (
	"context"

	"github.com/shurcooL/githubv4"
)

type DiscussionResult struct {
	ID     string
	Number int
}

func (c *Client) CreateDiscussion(title, body, categoryID string) (*DiscussionResult, error) {
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
		return nil, err
	}

	return &DiscussionResult{
		ID:     mutation.CreateDiscussion.Discussion.ID,
		Number: mutation.CreateDiscussion.Discussion.Number,
	}, nil
}

func (c *Client) AddComment(discussionID, body string) error {
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

	return c.client.Mutate(context.Background(), &mutation, input, nil)
}
