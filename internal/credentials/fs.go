package credentials

import (
	"encoding/json"
	"fmt"
	"os"
)

type fsAuth struct {
	Tokens struct {
		IDToken      string `json:"id_token"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		AccountID    string `json:"account_id"`
	} `json:"tokens"`
}

type FSCredentialsFetcher struct {
	Path string
}

func NewFSCredentialsFetcher(path string) *FSCredentialsFetcher {
	return &FSCredentialsFetcher{Path: path}
}

func (f *FSCredentialsFetcher) GetCredentials() (string, string, error) {
	b, err := os.ReadFile(f.Path)
	if err != nil {
		return "", "", fmt.Errorf("failed to read credentials file: %w", err)
	}
	var a fsAuth
	if err := json.Unmarshal(b, &a); err != nil {
		return "", "", fmt.Errorf("failed to parse credentials file: %w", err)
	}
	// Prefer OAuth access token for upstream API; fall back to ID token
	token := a.Tokens.AccessToken
	if token == "" {
		token = a.Tokens.IDToken
	}
	if token == "" || a.Tokens.AccountID == "" {
		return "", "", fmt.Errorf("missing token or account_id in credentials file")
	}
	return token, a.Tokens.AccountID, nil
}

func (f *FSCredentialsFetcher) RefreshCredentials() error {
	return nil
}
