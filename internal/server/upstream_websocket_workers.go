//go:build js && wasm

package server

import (
	"fmt"
	"net/http"
)

func supportsWebSocketUpstream() bool {
	return false
}

func (s *Server) makeChatGPTWebSocketRequest(r *http.Request, rawURL string, body []byte, token, accountID string) (*http.Response, int, error) {
	return nil, 0, fmt.Errorf("websocket upstream transport is not supported in js/wasm builds")
}
