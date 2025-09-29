package credentials

// CredentialsFetcher defines the interface for retrieving credentials
type CredentialsFetcher interface {
	GetCredentials() (apiKey, userID string, err error)
	RefreshCredentials() error
}

// OAuthCredentials represents full OAuth credential information
type OAuthCredentials struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
	UserID       string
}

// OAuthCredentialsFetcher extends CredentialsFetcher with OAuth-specific operations
type OAuthCredentialsFetcher interface {
	CredentialsFetcher
	GetFullCredentials() (*OAuthCredentials, error)
	UpdateTokens(accessToken, refreshToken string, expiresAt int64) error
}
