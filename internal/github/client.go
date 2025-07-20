// Package github provides a GitHub GraphQL API client for creating and managing
// GitHub Discussions. It includes rate limiting, retry mechanisms, error handling,
// and comprehensive statistics tracking for migration operations.
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

// Client provides a GitHub GraphQL API client with built-in rate limiting,
// retry mechanisms, and statistics tracking. It manages GitHub Discussions
// operations with automatic error recovery and monitoring.
type Client struct {
	client               *githubv4.Client // GitHub GraphQL client
	repositoryID         string           // Target repository ID
	repositoryName       string           // Repository name for logging
	rateLimitDelay       time.Duration    // Delay between API calls
	maxRetries           int              // Maximum retry attempts
	retryBackoffMultiple int              // Exponential backoff multiplier
	operationCount       int64            // Total operations attempted (atomic)
	rateLimitHits        int64            // Rate limit encounters (atomic)
}

// RateLimitError represents a GitHub API rate limit violation.
// Contains timing information for retry scheduling and quota details.
type RateLimitError struct {
	ResetTime time.Time // When the rate limit resets
	Remaining int       // Remaining quota (usually 0)
	Message   string    // Original error message
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("GitHub API rate limit exceeded: %s (remaining: %d, resets at %s)",
		e.Message, e.Remaining, e.ResetTime.Format(time.RFC3339))
}

// NewClient creates a new GitHub GraphQL API client with comprehensive validation.
// Validates token format, rate limiting parameters, and retry configuration.
// Returns an initialized client ready for GitHub Discussions operations.
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

// SetRepositoryID configures the target repository ID for GitHub operations.
// This ID is used for GraphQL queries and mutations.
func (c *Client) SetRepositoryID(id string) {
	c.repositoryID = id
}

// GetRepositoryID returns the currently configured repository ID.
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
	var lastErr error
	atomic.AddInt64(&c.operationCount, 1)

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := c.checkContextCancellation(ctx); err != nil {
			return err
		}

		if err := c.handleDelays(ctx, attempt); err != nil {
			return err
		}

		err := operation()
		if err == nil {
			c.logSuccessAfterRetries(attempt)
			return nil
		}

		lastErr = err

		if shouldContinue, retryErr := c.handleRetryableError(ctx, err, attempt); retryErr != nil {
			return retryErr
		} else if !shouldContinue {
			return err
		}

		c.logRetryAttempt(attempt, err)
	}

	return fmt.Errorf("GitHub API operation failed after %d retries: %w", c.maxRetries, lastErr)
}

// checkContextCancellation checks if the context has been cancelled
func (c *Client) checkContextCancellation(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %w", ctx.Err())
	default:
		return nil
	}
}

// handleDelays manages exponential backoff and rate limiting delays
func (c *Client) handleDelays(ctx context.Context, attempt int) error {
	const maxBackoffDuration = 5 * time.Minute

	if attempt > 0 {
		backoffDuration := c.calculateBackoffDuration(attempt, maxBackoffDuration)
		log.Printf("GitHub API retry attempt %d/%d, waiting %v... (total ops: %d, rate limit hits: %d)",
			attempt, c.maxRetries, backoffDuration, atomic.LoadInt64(&c.operationCount), atomic.LoadInt64(&c.rateLimitHits))

		return c.waitWithContext(ctx, backoffDuration, "operation cancelled during backoff")
	} else if c.rateLimitDelay > 0 {
		return c.waitWithContext(ctx, c.rateLimitDelay, "operation cancelled during rate limit delay")
	}

	return nil
}

// calculateBackoffDuration calculates the exponential backoff duration with maximum cap
func (c *Client) calculateBackoffDuration(attempt int, maxDuration time.Duration) time.Duration {
	backoffDuration := time.Duration(attempt*c.retryBackoffMultiple) * time.Second
	if backoffDuration > maxDuration {
		backoffDuration = maxDuration
	}
	return backoffDuration
}

// waitWithContext waits for the specified duration while respecting context cancellation
func (c *Client) waitWithContext(ctx context.Context, duration time.Duration, cancelMessage string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", cancelMessage, ctx.Err())
	case <-time.After(duration):
		return nil
	}
}

// handleRetryableError processes rate limit and retryable errors
// Returns (shouldContinue, error) where shouldContinue indicates if the retry loop should continue
func (c *Client) handleRetryableError(ctx context.Context, err error, attempt int) (bool, error) {
	if rateLimitErr, isRateLimit := c.parseRateLimitFromError(err); isRateLimit {
		return c.handleRateLimitError(ctx, rateLimitErr, attempt)
	}

	if !c.isRetryableError(err) {
		log.Printf("GitHub API operation failed with non-retryable error: %v", err)
		return false, nil
	}

	if attempt >= c.maxRetries {
		log.Printf("Maximum retries (%d) exceeded for GitHub API operation (total ops: %d)", c.maxRetries, atomic.LoadInt64(&c.operationCount))
		return false, nil
	}

	return true, nil
}

// handleRateLimitError processes rate limit errors with appropriate waiting
func (c *Client) handleRateLimitError(ctx context.Context, rateLimitErr *RateLimitError, attempt int) (bool, error) {
	atomic.AddInt64(&c.rateLimitHits, 1)
	log.Printf("GitHub API rate limit detected (#%d): %s", atomic.LoadInt64(&c.rateLimitHits), rateLimitErr.Error())

	if attempt >= c.maxRetries {
		log.Printf("Maximum retries (%d) exceeded for GitHub API rate limit (total rate limit hits: %d)", c.maxRetries, atomic.LoadInt64(&c.rateLimitHits))
		return false, rateLimitErr
	}

	waitTime := time.Until(rateLimitErr.ResetTime)
	if waitTime > 0 && waitTime < 2*time.Hour {
		log.Printf("Waiting %v for GitHub API rate limit to reset... (hit #%d)", waitTime, atomic.LoadInt64(&c.rateLimitHits))

		if err := c.waitWithContext(ctx, waitTime, "operation cancelled during rate limit wait"); err != nil {
			return false, err
		}
	}

	return true, nil
}

// logSuccessAfterRetries logs successful operations after retries
func (c *Client) logSuccessAfterRetries(attempt int) {
	if attempt > 0 {
		log.Printf("GitHub API operation succeeded after %d retries (total ops: %d)", attempt, atomic.LoadInt64(&c.operationCount))
	}
}

// logRetryAttempt logs retry attempts
func (c *Client) logRetryAttempt(attempt int, err error) {
	log.Printf("GitHub API operation failed (attempt %d/%d): %v", attempt+1, c.maxRetries+1, err)
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
