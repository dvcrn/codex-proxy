package credentials

import (
	"testing"
)

func TestEnvCredentialsFetcher(t *testing.T) {
	fetcher := NewEnvCredentialsFetcher()

	// Test GetCredentials (will return empty values if env vars not set)
	apiKey, userID, err := fetcher.GetCredentials()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should always return something (even if empty)
	_ = apiKey
	_ = userID

	// Test RefreshCredentials (should be no-op)
	err = fetcher.RefreshCredentials()
	if err != nil {
		t.Errorf("Expected no error from RefreshCredentials, got %v", err)
	}
}
