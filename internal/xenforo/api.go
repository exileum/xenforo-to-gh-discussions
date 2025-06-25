package xenforo

import (
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

func (c *Client) GetThreads(nodeID int) ([]Thread, error) {
	var threads []Thread
	page := 1

	for {
		resp, err := c.retryableRequest(func() (*resty.Response, error) {
			return c.addHeaders(c.client.R()).
				SetQueryParams(map[string]string{
					"page":    fmt.Sprintf("%d", page),
					"node_id": fmt.Sprintf("%d", nodeID),
				}).
				Get(c.baseURL + "/threads")
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
		time.Sleep(1 * time.Second)
	}

	return threads, nil
}

func (c *Client) GetPosts(threadID int) ([]Post, error) {
	var posts []Post
	page := 1

	for {
		resp, err := c.retryableRequest(func() (*resty.Response, error) {
			return c.addHeaders(c.client.R()).
				SetQueryParam("page", fmt.Sprintf("%d", page)).
				Get(fmt.Sprintf("%s/threads/%d/posts", c.baseURL, threadID))
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

		if result.Pagination.CurrentPage >= result.Pagination.TotalPages {
			break
		}

		page++
		time.Sleep(1 * time.Second)
	}

	return posts, nil
}

func (c *Client) GetAttachments(threadID int) ([]Attachment, error) {
	resp, err := c.retryableRequest(func() (*resty.Response, error) {
		return c.addHeaders(c.client.R()).
			Get(fmt.Sprintf("%s/threads/%d/attachments", c.baseURL, threadID))
	})

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API error (attachments): %s", resp.String())
	}

	var result AttachmentsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return result.Attachments, nil
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

// GetDryRunStats returns statistics for a node without fetching all data
func (c *Client) GetDryRunStats(nodeID int) (threadCount, postCount, attachmentCount, userCount int, err error) {
	// Get first page of threads to get total count from pagination
	resp, err := c.retryableRequest(func() (*resty.Response, error) {
		return c.addHeaders(c.client.R()).
			SetQueryParams(map[string]string{
				"page":    "1",
				"node_id": fmt.Sprintf("%d", nodeID),
			}).
			Get(c.baseURL + "/threads")
	})

	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to get thread statistics: %w", err)
	}

	if resp.StatusCode() != 200 {
		return 0, 0, 0, 0, fmt.Errorf("API error: %s", resp.String())
	}

	var result ThreadsResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to parse threads response: %w", err)
	}

	// Since pagination doesn't include total, we need to estimate based on TotalPages
	threadCount = result.Pagination.TotalPages * len(result.Threads)
	if result.Pagination.CurrentPage == result.Pagination.TotalPages {
		// On the last page, we have exact count
		threadCount = (result.Pagination.TotalPages-1)*len(result.Threads) + len(result.Threads)
	}

	// For posts, we need to make an estimate
	// Since we can't get post count from thread data, estimate 5 posts per thread on average
	postCount = threadCount * 5

	// For attachments, estimate 20% of posts have attachments
	attachmentCount = postCount / 5

	// For users, collect unique usernames from first page
	users := make(map[string]bool)
	for _, thread := range result.Threads {
		users[thread.Username] = true
	}

	// Estimate total users based on sample
	if len(result.Threads) > 0 {
		uniqueUsersOnPage := len(users)
		estimatedUsersPerPage := float64(uniqueUsersOnPage) * 0.8 // 80% unique rate estimate
		userCount = int(float64(result.Pagination.TotalPages) * estimatedUsersPerPage)
	}

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
