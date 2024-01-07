package clink

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// Client is a wrapper around http.Client with additional functionality.
type Client struct {
	HttpClient      *http.Client
	Headers         map[string]string
	RateLimiter     *rate.Limiter
	MaxRetries      int
	ShouldRetryFunc func(*http.Request, *http.Response, error) bool
}

// NewClient creates a new client with the given options.
func NewClient(opts ...Option) *Client {
	c := defaultClient()

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func defaultClient() *Client {
	return &Client{
		HttpClient: http.DefaultClient,
		Headers:    make(map[string]string),
	}
}

// Do sends the given request and returns the response.
// If the request is rate limited, the client will wait for the rate limiter to allow the request.
// If the request fails, the client will retry the request the number of times specified by MaxRetries.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	if c.RateLimiter != nil {
		if err := c.RateLimiter.Wait(req.Context()); err != nil {
			return nil, fmt.Errorf("failed to wait for rate limiter: %w", err)
		}
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		resp, err = c.HttpClient.Do(req)

		if c.ShouldRetryFunc != nil && !c.ShouldRetryFunc(req, resp, err) {
			break
		}

		if attempt < c.MaxRetries {
			// Exponential backoff only if we're going to retry.
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	return resp, nil
}

// Head sends a HEAD request to the given URL.
func (c *Client) Head(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Get sends a GET request to the given URL.
func (c *Client) Options(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodOptions, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Get sends a GET request to the given URL.
func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post sends a POST request to the given URL with the given body.
func (c *Client) Post(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Put sends a PUT request to the given URL.
func (c *Client) Put(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Patch sends a PATCH request to the given URL.
func (c *Client) Patch(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPatch, url, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Delete sends a DELETE request to the given URL.
func (c *Client) Delete(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

type Option func(*Client)

// WithClient sets the http client for the client.
func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.HttpClient = client
	}
}

// WithHeader sets a header for the client.
func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.Headers[key] = value
	}
}

// WithHeaders sets the headers for the client.
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		for key, value := range headers {
			c.Headers[key] = value
		}
	}
}

// WithRateLimit sets the rate limit for the client in requests per minute.
func WithRateLimit(rpm int) Option {
	return func(c *Client) {
		interval := time.Minute / time.Duration(rpm)
		c.RateLimiter = rate.NewLimiter(rate.Every(interval), 1)
	}
}

// WithBasicAuth sets the basic auth header for the client.
func WithBasicAuth(username, password string) Option {
	return func(c *Client) {
		auth := username + ":" + password
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
		c.Headers["Authorization"] = "Basic " + encodedAuth
	}
}

// WithBearerAuth sets the bearer auth header for the client.
func WithBearerAuth(token string) Option {
	return func(c *Client) {
		c.Headers["Authorization"] = "Bearer " + token
	}
}

// WithUserAgent sets the user agent header for the client.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.Headers["User-Agent"] = ua
	}
}

// WithRetries sets the retry count and retry function for the client.
func WithRetries(count int, retryFunc func(*http.Request, *http.Response, error) bool) Option {
	return func(c *Client) {
		c.MaxRetries = count
		c.ShouldRetryFunc = retryFunc
	}
}

// ResponseToJson decodes the response body into the target.
func ResponseToJson[T any](response *http.Response, target *T) error {
	if response == nil {
		return fmt.Errorf("response is nil")
	}

	if response.Body == nil {
		return fmt.Errorf("response body is nil")
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
