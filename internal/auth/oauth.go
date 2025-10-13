package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	// OAuthTokenURL is the endpoint for refreshing OAuth tokens
	OAuthTokenURL = "https://auth.openai.com/oauth/token"
	// ClientID is the OAuth client ID for ChatGPT/Codex
	ClientID = "app_EMoamEEZ73f0CkXaXp7hrann"
	// TokenExpiryBuffer is the buffer time before token expiry to trigger refresh (60 minutes)
	TokenExpiryBuffer = 60 * time.Minute
)

// TokenExpired checks if the token is expired or will expire soon
func TokenExpired(expiresAtMs int64) bool {
	bufferMs := TokenExpiryBuffer.Milliseconds()
	currentTimeMs := time.Now().UnixMilli()
	return currentTimeMs >= (expiresAtMs - bufferMs)
}

// RefreshToken performs an OAuth token refresh and returns new credentials
func RefreshToken(refreshToken string) (*TokenRefreshResponse, error) {
	request := TokenRefreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
		ClientID:     ClientID,
		Scope:        "openid profile email",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refresh request: %w", err)
	}

	resp, err := http.Post(OAuthTokenURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, errorBody.String())
	}

	var tokenResp TokenRefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &tokenResp, nil
}

// CalculateExpiresAt calculates the expiry timestamp from expires_in seconds
func CalculateExpiresAt(expiresIn int) int64 {
	return (time.Now().Unix() + int64(expiresIn)) * 1000 // Convert to milliseconds
}
