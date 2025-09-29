//go:build js && wasm

package server

import (
	"net/http"

	"github.com/syumai/workers/cloudflare/fetch"
)

// WorkersHTTPClient implements HTTPClient for Cloudflare Workers
type WorkersHTTPClient struct {
	client *fetch.Client
}

// NewHTTPClient creates a new HTTP client for Workers environment
func NewHTTPClient() HTTPClient {
	return &WorkersHTTPClient{
		client: fetch.NewClient(),
	}
}

// Do performs an HTTP request using Cloudflare Workers fetch
func (c *WorkersHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Create a new fetch request
	fetchReq, err := fetch.NewRequest(req.Context(), req.Method, req.URL.String(), req.Body)
	if err != nil {
		return nil, err
	}

	// Copy headers from the original request
	for key, values := range req.Header {
		for _, value := range values {
			fetchReq.Header.Set(key, value)
		}
	}

	// Perform the request
	return c.client.Do(fetchReq, nil)
}
