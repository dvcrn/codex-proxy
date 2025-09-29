package credentials

import (
	"fmt"
	"os"
)

// EnvCredentialsFetcher retrieves credentials from environment variables
type EnvCredentialsFetcher struct{}

// NewEnvCredentialsFetcher creates a new environment-based credentials fetcher
func NewEnvCredentialsFetcher() *EnvCredentialsFetcher {
	return &EnvCredentialsFetcher{}
}

// GetCredentials retrieves credentials from environment variables
func (e *EnvCredentialsFetcher) GetCredentials() (string, string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	userID := os.Getenv("CLAUDE_USER_ID")
	return apiKey, userID, nil
}

// RefreshCredentials is a no-op for environment credentials
func (e *EnvCredentialsFetcher) RefreshCredentials() error {
	return nil
}

// GetFullCredentials returns an error for environment credentials as they don't support OAuth
func (e *EnvCredentialsFetcher) GetFullCredentials() (*OAuthCredentials, error) {
	return nil, fmt.Errorf("environment credentials do not support OAuth tokens")
}

// UpdateTokens returns an error for environment credentials as they don't support OAuth
func (e *EnvCredentialsFetcher) UpdateTokens(accessToken, refreshToken string, expiresAt int64) error {
	return fmt.Errorf("environment credentials do not support OAuth token updates")
}
