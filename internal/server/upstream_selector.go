package server

import "strings"

// shouldUseWebSocketUpstream determines whether a normalized backend model
// should be routed to the WebSocket upstream transport.
func shouldUseWebSocketUpstream(normalizedModel string) bool {
	if !supportsWebSocketUpstream() {
		return false
	}
	return strings.TrimSpace(normalizedModel) == modelGPT53CodexSpark
}

func upstreamTransportForModel(normalizedModel string) string {
	if shouldUseWebSocketUpstream(normalizedModel) {
		return "websocket"
	}
	return "http"
}
