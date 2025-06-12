package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPClient interface {
	Get(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)
	Post(ctx context.Context, url string, formData map[string][]string, headers map[string]string) (*HTTPResponse, error)
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
}

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Get(ctx context.Context, reqURL string, headers map[string]string) (*HTTPResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request w context: %w", err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.doRequest(req)
}

func (c *Client) Post(ctx context.Context, reqURL string, formData map[string][]string, headers map[string]string) (*HTTPResponse, error) {
	form := url.Values{}
	for key, values := range formData {
		for _, value := range values {
			form.Add(key, value)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating post request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return c.doRequest(req)
}

func (c *Client) doRequest(req *http.Request) (*HTTPResponse, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}
