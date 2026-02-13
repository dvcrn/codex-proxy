package server

import "testing"

func TestUpstreamTransportForModel_DefaultsToHTTP(t *testing.T) {
	if got := upstreamTransportForModel(modelGPT53Codex); got != "http" {
		t.Fatalf("expected non-spark model to use http transport, got %q", got)
	}
}

func TestUpstreamTransportForModel_SparkUsesWebSocketWhenSupported(t *testing.T) {
	if !supportsWebSocketUpstream() {
		t.Skip("websocket upstream is not available in this build")
	}

	if got := upstreamTransportForModel(modelGPT53CodexSpark); got != "websocket" {
		t.Fatalf("expected spark model to use websocket transport, got %q", got)
	}
}
