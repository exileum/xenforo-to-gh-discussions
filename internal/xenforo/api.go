package xenforo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

func (c *Client) TestConnection() error {
	resp, err := c.retryableRequest(func() (*resty.Response, error) {
		return c.addHeaders(c.client.R()).Get(c.baseURL + "/")
	})

	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	if resp.StatusCode() == 401 {
		return fmt.Errorf("authentication failed - check API key and user ID")
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("API error: %s", resp.String())
	}

	return nil
}

func (c *Client) GetThreads(ctx context.Context, nodeID int) ([]Thread, error) {
	var threads []Thread
	page := 1

	for {
		// Check context cancellation before each iteration
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := c.retryableRequest(func() (*resty.Response, error) {
			return c.addHeaders(c.client.R()).
				SetContext(ctx).
				SetQueryParam("page", fmt.Sprintf("%d", page)).
				Get(fmt.Sprintf("%s/forums/%d/threads", c.baseURL, nodeID))
		})

		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("API error: %s", resp.String())
		}

		var result ThreadsResponse
		if err := json.Unmarshal(resp.Body(), &result); err != nil {
			return nil, err
		}

		threads = append(threads, result.Threads...)

		if result.Pagination.CurrentPage >= result.Pagination.TotalPages {
			break
		}

		page++

		// Check context before sleep
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}

	return threads, nil
}

func (c *Client) GetPosts(ctx context.Context, thread Thread) ([]Post, error) {
	var posts []Post

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Calculate total posts: reply_count + 1 (original post)
	totalPosts := thread.ReplyCount + 1

	// Start with first page to determine posts per page
	firstPageResp, err := c.retryableRequest(func() (*resty.Response, error) {
		return c.addHeaders(c.client.R()).
			SetContext(ctx).
			SetQueryParam("page", "1").
			Get(fmt.Sprintf("%s/threads/%d/posts", c.baseURL, thread.ThreadID))
	})

	if err != nil {
		return nil, err
	}

	if firstPageResp.StatusCode() != 200 {
		return nil, fmt.Errorf("API error: %s", firstPageResp.String())
	}

	var firstResult PostsResponse
	if err := json.Unmarshal(firstPageResp.Body(), &firstResult); err != nil {
		return nil, err
	}

	posts = append(posts, firstResult.Posts...)
	postsPerPage := len(firstResult.Posts)

	// If we got all posts on the first page, we're done
	if len(posts) >= totalPosts {
		return posts, nil
	}

	// Calculate how many more pages we need
	totalPages := (totalPosts + postsPerPage - 1) / postsPerPage // Ceiling division

	// Fetch remaining pages
	for page := 2; page <= totalPages; page++ {
		// Check context cancellation before each iteration
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := c.retryableRequest(func() (*resty.Response, error) {
			return c.addHeaders(c.client.R()).
				SetContext(ctx).
				SetQueryParam("page", fmt.Sprintf("%d", page)).
				Get(fmt.Sprintf("%s/threads/%d/posts", c.baseURL, thread.ThreadID))
		})

		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("API error: %s", resp.String())
		}

		var result PostsResponse
		if err := json.Unmarshal(resp.Body(), &result); err != nil {
			return nil, err
		}

		posts = append(posts, result.Posts...)

		// Break if we got fewer posts than expected (last page)
		if len(result.Posts) < postsPerPage {
			break
		}

		// Check context before sleep
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}

	return posts, nil
}

func (c *Client) DownloadAttachment(url, filepath string) error {
	resp, err := c.retryableRequest(func() (*resty.Response, error) {
		return c.addHeaders(c.client.R()).
			SetOutput(filepath).
			Get(url)
	})

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("download failed: status %d", resp.StatusCode())
	}

	return nil
}

// GetDryRunStats returns statistics for a node by fetching actual data
func (c *Client) GetDryRunStats(ctx context.Context, nodeID int) (threadCount, postCount, attachmentCount, userCount int, err error) {
	// Get all threads from the node using our working GetThreads method
	threads, err := c.GetThreads(ctx, nodeID)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to get threads: %w", err)
	}

	// Count threads directly
	threadCount = len(threads)

	// Calculate accurate post count using reply_count from each thread
	postCount = 0
	users := make(map[string]bool)
	for _, thread := range threads {
		// Each thread has reply_count replies + 1 original post
		postCount += thread.ReplyCount + 1
		users[thread.Username] = true
	}

	// For attachments, estimate 10% of posts have attachments on average
	// This is a conservative estimate since we'd need to fetch all posts to get exact count
	attachmentCount = postCount / 10

	// Count unique users from thread authors
	userCount = len(users)

	return threadCount, postCount, attachmentCount, userCount, nil
}

// GetNodes fetches available forum nodes/categories from XenForo
func (c *Client) GetNodes() ([]Node, error) {
	resp, err := c.retryableRequest(func() (*resty.Response, error) {
		return c.addHeaders(c.client.R()).Get(c.baseURL + "/nodes")
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	var result NodesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse nodes response: %w", err)
	}

	return result.Nodes, nil
}
