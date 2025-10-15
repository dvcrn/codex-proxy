package credentials

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInitFromOAuth(t *testing.T) {
	tmpDir := t.TempDir()
	authPath := filepath.Join(tmpDir, "nested", "auth.json")

	testCreds := &OAuthCredentials{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    1234567890000,
		UserID:       "test-user-id",
	}

	err := InitFromOAuth(authPath, testCreds)
	if err != nil {
		t.Fatalf("InitFromOAuth failed: %v", err)
	}

	info, err := os.Stat(authPath)
	if err != nil {
		t.Fatalf("Failed to stat created file: %v", err)
	}

	expectedPerm := os.FileMode(0600)
	if info.Mode().Perm() != expectedPerm {
		t.Errorf("Expected file permissions %v, got %v", expectedPerm, info.Mode().Perm())
	}

	data, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	var auth fsAuth
	if err := json.Unmarshal(data, &auth); err != nil {
		t.Fatalf("Failed to parse created JSON: %v", err)
	}

	if auth.Tokens.AccessToken != testCreds.AccessToken {
		t.Errorf("AccessToken mismatch: expected %s, got %s", testCreds.AccessToken, auth.Tokens.AccessToken)
	}

	if auth.Tokens.RefreshToken != testCreds.RefreshToken {
		t.Errorf("RefreshToken mismatch: expected %s, got %s", testCreds.RefreshToken, auth.Tokens.RefreshToken)
	}

	if auth.Tokens.ExpiresAt != testCreds.ExpiresAt {
		t.Errorf("ExpiresAt mismatch: expected %d, got %d", testCreds.ExpiresAt, auth.Tokens.ExpiresAt)
	}

	if auth.Tokens.AccountID != testCreds.UserID {
		t.Errorf("AccountID mismatch: expected %s, got %s", testCreds.UserID, auth.Tokens.AccountID)
	}
}

func TestInitFromOAuthCreatesParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	authPath := filepath.Join(tmpDir, "deeply", "nested", "structure", "auth.json")

	testCreds := &OAuthCredentials{
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
		ExpiresAt:    1234567890000,
		UserID:       "test-user",
	}

	err := InitFromOAuth(authPath, testCreds)
	if err != nil {
		t.Fatalf("InitFromOAuth failed: %v", err)
	}

	parentDir := filepath.Dir(authPath)
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("Parent directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected parent to be a directory")
	}

	expectedPerm := os.FileMode(0700)
	if info.Mode().Perm() != expectedPerm {
		t.Errorf("Expected directory permissions %v, got %v", expectedPerm, info.Mode().Perm())
	}
}
