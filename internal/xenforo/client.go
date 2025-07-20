package xenforo

import (
	"fmt"
	"math"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	baseURL    string
	apiKey     string
	apiUser    string
	maxRetries int
	client     *resty.Client
}

func NewClient(baseURL, apiKey, apiUser string, maxRetries int) *Client {
	restyClient := resty.New().
		SetTimeout(30*time.Second).
		SetRetryCount(0).
		SetRetryWaitTime(1*time.Second).
		SetRetryMaxWaitTime(10*time.Second).
		SetHeader("User-Agent", "XenForo-to-GH-Discussions/1.0").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json")

	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		apiUser:    apiUser,
		maxRetries: maxRetries,
		client:     restyClient,
	}
}

func (c *Client) retryableRequest(req func() (*resty.Response, error)) (*resty.Response, error) {
	for i := 0; i < c.maxRetries; i++ {
		resp, err := req()

		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != 429 {
			return resp, nil
		}

		if i < c.maxRetries-1 {
			delay := time.Duration(math.Pow(2, float64(i))) * time.Second
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("max retries (%d) exceeded", c.maxRetries)
}

// SetTimeout allows customizing the HTTP timeout after client creation
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) addHeaders(req *resty.Request) *resty.Request {
	return req.
		SetHeader("XF-Api-Key", c.apiKey).
		SetHeader("XF-Api-User", c.apiUser)
}
