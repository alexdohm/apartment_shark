package mock

import (
	"apartmenthunter/internal/http"
	"context"
)

type HTTPClient struct {
	GetFunc      func(ctx context.Context, url string, headers map[string]string) (*http.HTTPResponse, error)
	PostFunc     func(ctx context.Context, url string, formData map[string][]string, headers map[string]string) (*http.HTTPResponse, error)
	PostJSONFunc func(ctx context.Context, url string, jsonBody []byte, headers map[string]string) (*http.HTTPResponse, error)
}

func (m *HTTPClient) Get(ctx context.Context, url string, headers map[string]string) (*http.HTTPResponse, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, url, headers)
	}
	return &http.HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
}

func (m *HTTPClient) Post(ctx context.Context, url string, formData map[string][]string, headers map[string]string) (*http.HTTPResponse, error) {
	if m.PostFunc != nil {
		return m.PostFunc(ctx, url, formData, headers)
	}
	return &http.HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
}

func (m *HTTPClient) PostJSON(ctx context.Context, url string, jsonBody []byte, headers map[string]string) (*http.HTTPResponse, error) {
	if m.PostJSONFunc != nil {
		return m.PostJSONFunc(ctx, url, jsonBody, headers)
	}
	return &http.HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{}
}
