//go:build !js || !wasm

package server

import (
	"net/http"
	"time"
)

// NewHTTPClient creates a new HTTP client for regular environments
func NewHTTPClient() HTTPClient {
	return &http.Client{
		Timeout: 60 * time.Second,
	}
}
