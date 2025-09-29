package credentials

import (
	"testing"
	"time"
)

func TestKeychainCredentialsFetcher(t *testing.T) {
	fetcher := NewKeychainCredentialsFetcher()
	defer fetcher.Close()

	// Test that the fetcher was created
	if fetcher == nil {
		t.Fatal("Expected fetcher to be created")
	}

	// Test that cache TTL was set
	if fetcher.cacheTTL != 5*time.Minute {
		t.Errorf("Expected cacheTTL to be 5 minutes, got %v", fetcher.cacheTTL)
	}

	// Test that stopCh was created
	if fetcher.stopCh == nil {
		t.Error("Expected stopCh to be created")
	}
}
