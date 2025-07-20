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

	client := &Client{
		client:               graphqlClient,
		rateLimitDelay:       rateLimitDelay,
		maxRetries:           maxRetries,
		retryBackoffMultiple: retryBackoffMultiple,
	}

	// Log rate limit configuration
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

// parseRateLimitFromError extracts rate limit information from GitHub API errors
func (c *Client) parseRateLimitFromError(err error) (*RateLimitError, bool) {
	if err == nil {
		return nil, false
	}

	errStr := err.Error()

	// Check if this is a rate limit error
	if !strings.Contains(strings.ToLower(errStr), "rate limit") &&
		!strings.Contains(strings.ToLower(errStr), "api rate limit exceeded") &&
		!strings.Contains(strings.ToLower(errStr), "secondary rate limit") &&
		!strings.Contains(strings.ToLower(errStr), "abuse detection") {
		return nil, false
	}

	resetTime := time.Now().Add(1 * time.Hour) // Default to 1 hour if we can't parse

	// Try to parse reset time from common GitHub API error patterns
	if strings.Contains(errStr, "please retry your request after") {
		// GitHub usually provides a specific time to retry
		resetTime = time.Now().Add(60 * time.Minute) // Conservative default
	} else if strings.Contains(strings.ToLower(errStr), "secondary rate limit") {
		// Secondary rate limits typically reset faster
		resetTime = time.Now().Add(10 * time.Minute)
	}

	rateLimitErr := &RateLimitError{
		Message:   errStr,
		Remaining: 0,
		ResetTime: resetTime,
	}

	return rateLimitErr, true
}

// logRateLimitStatus logs current rate limit status if available
func (c *Client) logRateLimitStatus() {
	log.Printf("GitHub API: Using rate limit delay: %v, max retries: %d, backoff multiplier: %dx",
		c.rateLimitDelay, c.maxRetries, c.retryBackoffMultiple)
}

// executeWithRetry executes a function with rate limit handling and exponential backoff
func (c *Client) executeWithRetry(operation func() error) error {
	var lastErr error
	atomic.AddInt64(&c.operationCount, 1)

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Apply rate limit delay before each attempt (except the first)
		if attempt > 0 {
			backoffDuration := time.Duration(attempt*c.retryBackoffMultiple) * time.Second
			log.Printf("GitHub API retry attempt %d/%d, waiting %v... (total ops: %d, rate limit hits: %d)",
				attempt, c.maxRetries, backoffDuration, atomic.LoadInt64(&c.operationCount), atomic.LoadInt64(&c.rateLimitHits))
			time.Sleep(backoffDuration)
		} else if c.rateLimitDelay > 0 {
			// Always apply base rate limit delay
			time.Sleep(c.rateLimitDelay)
		}

		// Execute the operation
		err := operation()
		if err == nil {
			// Success!
			if attempt > 0 {
				log.Printf("GitHub API operation succeeded after %d retries (total ops: %d)", attempt, atomic.LoadInt64(&c.operationCount))
			}
			return nil
		}

		lastErr = err

		// Check if this is a rate limit error
		if rateLimitErr, isRateLimit := c.parseRateLimitFromError(err); isRateLimit {
			atomic.AddInt64(&c.rateLimitHits, 1)
			log.Printf("GitHub API rate limit detected (#%d): %s", atomic.LoadInt64(&c.rateLimitHits), rateLimitErr.Error())

			// If we've exhausted retries, return the rate limit error
			if attempt >= c.maxRetries {
				log.Printf("Maximum retries (%d) exceeded for GitHub API rate limit (total rate limit hits: %d)", c.maxRetries, atomic.LoadInt64(&c.rateLimitHits))
				return rateLimitErr
			}

			// Calculate wait time until rate limit resets
			waitTime := time.Until(rateLimitErr.ResetTime)
			if waitTime > 0 && waitTime < 2*time.Hour { // Reasonable maximum wait time
				log.Printf("Waiting %v for GitHub API rate limit to reset... (hit #%d)", waitTime, atomic.LoadInt64(&c.rateLimitHits))
				time.Sleep(waitTime)
			}

			continue // Retry the operation
		}

		// If it's not a rate limit error, check if we should retry
		// For now, we'll retry on any error, but this could be made more specific
		if attempt >= c.maxRetries {
			log.Printf("Maximum retries (%d) exceeded for GitHub API operation (total ops: %d)", c.maxRetries, atomic.LoadInt64(&c.operationCount))
			break
		}

		log.Printf("GitHub API operation failed (attempt %d/%d): %v", attempt+1, c.maxRetries+1, err)
	}

	return fmt.Errorf("GitHub API operation failed after %d retries: %w", c.maxRetries, lastErr)
}

// GetStats returns operation statistics for monitoring
func (c *Client) GetStats() (operationCount, rateLimitHits int64) {
	return atomic.LoadInt64(&c.operationCount), atomic.LoadInt64(&c.rateLimitHits)
}
