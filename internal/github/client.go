package github

import (
	"context"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Client struct {
	client       *githubv4.Client
	repositoryID string
}

func NewClient(token string) (*Client, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	return &Client{
		client: githubv4.NewClient(httpClient),
	}, nil
}

func (c *Client) SetRepositoryID(id string) {
	c.repositoryID = id
}

func (c *Client) GetRepositoryID() string {
	return c.repositoryID
}
