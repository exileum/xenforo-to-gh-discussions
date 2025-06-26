package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
)

type RepositoryInfo struct {
	ID                    string
	HasDiscussionsEnabled bool
	DiscussionCategories  []Category
}

type Category struct {
	ID   string
	Name string
}

func (c *Client) GetRepositoryInfo(repo string) (*RepositoryInfo, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format - expected 'owner/repo'")
	}

	var query struct {
		Repository struct {
			ID                    string
			HasDiscussionsEnabled bool
			DiscussionCategories  struct {
				Nodes []struct {
					ID   string
					Name string
				}
			} `graphql:"discussionCategories(first: 100)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(parts[0]),
		"name":  githubv4.String(parts[1]),
	}

	err := c.client.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, fmt.Errorf("GitHub API query failed: %w", err)
	}

	if !query.Repository.HasDiscussionsEnabled {
		return nil, fmt.Errorf("GitHub Discussions is not enabled for repository %s", repo)
	}

	categories := make([]Category, len(query.Repository.DiscussionCategories.Nodes))
	for i, cat := range query.Repository.DiscussionCategories.Nodes {
		categories[i] = Category{
			ID:   cat.ID,
			Name: cat.Name,
		}
	}

	info := &RepositoryInfo{
		ID:                    query.Repository.ID,
		HasDiscussionsEnabled: query.Repository.HasDiscussionsEnabled,
		DiscussionCategories:  categories,
	}

	c.repositoryID = info.ID
	c.repositoryName = repo

	return info, nil
}

func (c *Client) ValidateCategoryMappings(categories map[int]string) error {
	// Ensure we have a repository name stored
	if strings.TrimSpace(c.repositoryName) == "" {
		return fmt.Errorf("repository name not set - call GetRepositoryInfo first")
	}

	info, err := c.GetRepositoryInfo(c.repositoryName)
	if err != nil {
		return fmt.Errorf("failed to validate category mappings: %w", err)
	}

	validCategories := make(map[string]bool)
	for _, cat := range info.DiscussionCategories {
		validCategories[cat.ID] = true
	}

	for nodeID, categoryID := range categories {
		if !validCategories[categoryID] {
			return fmt.Errorf("invalid category ID '%s' for node %d", categoryID, nodeID)
		}
	}

	return nil
}
