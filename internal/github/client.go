package github

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Client struct {
	client               *githubv4.Client
	repositoryID         string
	repositoryName       string
	rateLimitDelay       time.Duration
	maxRetries           int
	retryBackoffMultiple int
	operationCount       int64
	rateLimitHits        int64
}

type RateLimitError struct {
	ResetTime time.Time
	Remaining int
	Message   string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("GitHub API rate limit exceeded: %s (remaining: %d, resets at %s)",
		e.Message, e.Remaining, e.ResetTime.Format(time.RFC3339))
}

func NewClient(token string, rateLimitDelay time.Duration, maxRetries, retryBackoffMultiple int) (*Client, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("GitHub token cannot be empty")
	}

	if len(token) < 20 {
		return nil, errors.New("GitHub token appears to be invalid (too short)")
	}

	if rateLimitDelay < 0 {
		return nil, errors.New("rate limit delay cannot be negative")
	}
	if maxRetries < 0 {
		return nil, errors.New("max retries cannot be negative")
	}
	if retryBackoffMultiple < 1 {
		return nil, errors.New("retry backoff multiple must be at least 1")
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	httpClient := oauth2.NewClient(context.Background(), src)
	if httpClient == nil {
		return nil, errors.New("failed to create OAuth2 HTTP client")
	}

	graphqlClient := githubv4.NewClient(httpClient)
	if graphqlClient == nil {
		return nil, errors.New("failed to create GitHub GraphQL client")
	}

	client := &Client{
		client:               graphqlClient,
		rateLimitDelay:       rateLimitDelay,
		maxRetries:           maxRetries,
		retryBackoffMultiple: retryBackoffMultiple,
	}

	client.logRateLimitStatus()

	return client, nil
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

func (c *Client) parseRateLimitFromError(err error) (*RateLimitError, bool) {
	if err == nil {
		return nil, false
	}

	errStr := err.Error()

	if !strings.Contains(strings.ToLower(errStr), "rate limit") &&
		!strings.Contains(strings.ToLower(errStr), "api rate limit exceeded") &&
		!strings.Contains(strings.ToLower(errStr), "secondary rate limit") &&
		!strings.Contains(strings.ToLower(errStr), "abuse detection") {
		return nil, false
	}

	resetTime := time.Now().Add(1 * time.Hour)

	if strings.Contains(errStr, "please retry your request after") {
		resetTime = time.Now().Add(60 * time.Minute)
	} else if strings.Contains(strings.ToLower(errStr), "secondary rate limit") {
		resetTime = time.Now().Add(10 * time.Minute)
	}

	rateLimitErr := &RateLimitError{
		Message:   errStr,
		Remaining: 0,
		ResetTime: resetTime,
	}

	return rateLimitErr, true
}

func (c *Client) logRateLimitStatus() {
	log.Printf("GitHub API: Using rate limit delay: %v, max retries: %d, backoff multiplier: %dx",
		c.rateLimitDelay, c.maxRetries, c.retryBackoffMultiple)
}

// executeWithRetry executes a function with rate limit handling, exponential backoff, and context support
func (c *Client) executeWithRetry(ctx context.Context, operation func() error) error {
	const maxBackoffDuration = 5 * time.Minute

	var lastErr error
	atomic.AddInt64(&c.operationCount, 1)

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		if attempt > 0 {
			backoffDuration := time.Duration(attempt*c.retryBackoffMultiple) * time.Second
			if backoffDuration > maxBackoffDuration {
				backoffDuration = maxBackoffDuration
			}
			log.Printf("GitHub API retry attempt %d/%d, waiting %v... (total ops: %d, rate limit hits: %d)",
				attempt, c.maxRetries, backoffDuration, atomic.LoadInt64(&c.operationCount), atomic.LoadInt64(&c.rateLimitHits))

			select {
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled during backoff: %w", ctx.Err())
			case <-time.After(backoffDuration):
			}
		} else if c.rateLimitDelay > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled during rate limit delay: %w", ctx.Err())
			case <-time.After(c.rateLimitDelay):
			}
		}

		err := operation()
		if err == nil {
			if attempt > 0 {
				log.Printf("GitHub API operation succeeded after %d retries (total ops: %d)", attempt, atomic.LoadInt64(&c.operationCount))
			}
			return nil
		}

		lastErr = err

		if rateLimitErr, isRateLimit := c.parseRateLimitFromError(err); isRateLimit {
			atomic.AddInt64(&c.rateLimitHits, 1)
			log.Printf("GitHub API rate limit detected (#%d): %s", atomic.LoadInt64(&c.rateLimitHits), rateLimitErr.Error())

			if attempt >= c.maxRetries {
				log.Printf("Maximum retries (%d) exceeded for GitHub API rate limit (total rate limit hits: %d)", c.maxRetries, atomic.LoadInt64(&c.rateLimitHits))
				return rateLimitErr
			}

			waitTime := time.Until(rateLimitErr.ResetTime)
			if waitTime > 0 && waitTime < 2*time.Hour {
				log.Printf("Waiting %v for GitHub API rate limit to reset... (hit #%d)", waitTime, atomic.LoadInt64(&c.rateLimitHits))

				select {
				case <-ctx.Done():
					return fmt.Errorf("operation cancelled during rate limit wait: %w", ctx.Err())
				case <-time.After(waitTime):
				}
			}

			continue
		}

		if !c.isRetryableError(err) {
			log.Printf("GitHub API operation failed with non-retryable error: %v", err)
			return err
		}

		if attempt >= c.maxRetries {
			log.Printf("Maximum retries (%d) exceeded for GitHub API operation (total ops: %d)", c.maxRetries, atomic.LoadInt64(&c.operationCount))
			break
		}

		log.Printf("GitHub API operation failed (attempt %d/%d): %v", attempt+1, c.maxRetries+1, err)
	}

	return fmt.Errorf("GitHub API operation failed after %d retries: %w", c.maxRetries, lastErr)
}

// isRetryableError determines if an error is transient and should trigger a retry
func (c *Client) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	retryablePatterns := []string{
		"connection reset",
		"connection refused",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"no such host",
		"server error",
		"internal server error",
		"bad gateway",
		"service unavailable",
		"gateway timeout",
		"502", "503", "504",
		"unexpected eof",
		"broken pipe",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	nonRetryablePatterns := []string{
		"unauthorized",
		"forbidden",
		"not found",
		"bad request",
		"invalid",
		"401", "403", "404", "400",
	}

	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	return true
}

// GetStats returns operation statistics for monitoring
func (c *Client) GetStats() (operationCount, rateLimitHits int64) {
	return atomic.LoadInt64(&c.operationCount), atomic.LoadInt64(&c.rateLimitHits)
}
