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
