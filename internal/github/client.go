package github

import (
	"context"
	"errors"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Client struct {
	client         *githubv4.Client
	repositoryID   string
	repositoryName string
}

func NewClient(token string) (*Client, error) {
	// Validate token parameter
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("GitHub token cannot be empty")
	}

	// Basic token format validation (GitHub tokens are typically 40+ characters)
	if len(token) < 20 {
		return nil, errors.New("GitHub token appears to be invalid (too short)")
	}

	// Create OAuth2 token source
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	// Create HTTP client with OAuth2 token
	httpClient := oauth2.NewClient(context.Background(), src)
	if httpClient == nil {
		return nil, errors.New("failed to create OAuth2 HTTP client")
	}

	// Create GitHub GraphQL client
	graphqlClient := githubv4.NewClient(httpClient)
	if graphqlClient == nil {
		return nil, errors.New("failed to create GitHub GraphQL client")
	}

	return &Client{
		client: graphqlClient,
	}, nil
}

func (c *Client) SetRepositoryID(id string) {
	c.repositoryID = id
}

func (c *Client) GetRepositoryID() string {
	return c.repositoryID
}

func (c *Client) SetRepositoryName(name string) {
	c.repositoryName = name
}

func (c *Client) GetRepositoryName() string {
	return c.repositoryName
}
