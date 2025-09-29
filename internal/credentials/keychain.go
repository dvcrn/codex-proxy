package credentials

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// keychain-specific data structures
type keychainCredentials struct {
	ClaudeAiOauth struct {
		AccessToken  string   `json:"accessToken"`
		RefreshToken string   `json:"refreshToken"`
		ExpiresAt    int64    `json:"expiresAt"`
		Scopes       []string `json:"scopes"`
		IsMax        bool     `json:"isMax"`
	} `json:"claudeAiOauth"`
}

type claudeConfig struct {
	UserID string `json:"userID"`
}

type fullCredentials struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
	UserID       string
}

// KeychainCredentialsFetcher retrieves credentials from macOS keychain with caching
type KeychainCredentialsFetcher struct {
	mu          sync.RWMutex
	cachedKey   string
	cachedUser  string
	lastRefresh time.Time
	cacheTTL    time.Duration
	stopCh      chan struct{}
	logger      *zerolog.Logger
}

// NewKeychainCredentialsFetcher creates a new keychain-based credentials fetcher
func NewKeychainCredentialsFetcher() *KeychainCredentialsFetcher {
	f := &KeychainCredentialsFetcher{
		cacheTTL: 5 * time.Minute, // Cache credentials for 5 minutes
		stopCh:   make(chan struct{}),
	}
	go f.backgroundRefresh()
	return f
}

// NewKeychainCredentialsFetcherWithLogger creates a new keychain-based credentials fetcher with logger
func NewKeychainCredentialsFetcherWithLogger(logger zerolog.Logger) *KeychainCredentialsFetcher {
	f := &KeychainCredentialsFetcher{
		cacheTTL: 5 * time.Minute, // Cache credentials for 5 minutes
		stopCh:   make(chan struct{}),
		logger:   &logger,
	}
	go f.backgroundRefresh()
	return f
}

// GetCredentials retrieves credentials from cache or keychain
func (k *KeychainCredentialsFetcher) GetCredentials() (string, string, error) {
	k.mu.RLock()
	if k.cachedKey != "" && k.cachedUser != "" && time.Since(k.lastRefresh) < k.cacheTTL {
		apiKey, userID := k.cachedKey, k.cachedUser
		k.mu.RUnlock()
		return apiKey, userID, nil
	}
	k.mu.RUnlock()
	return k.refreshAndGet()
}

// GetFullCredentials retrieves full OAuth credentials from keychain
func (k *KeychainCredentialsFetcher) GetFullCredentials() (*OAuthCredentials, error) {
	creds, err := getFullCredentials()
	if err != nil {
		return nil, err
	}
	return &OAuthCredentials{
		AccessToken:  creds.AccessToken,
		RefreshToken: creds.RefreshToken,
		ExpiresAt:    creds.ExpiresAt,
		UserID:       creds.UserID,
	}, nil
}

// UpdateTokens updates the OAuth tokens in keychain
func (k *KeychainCredentialsFetcher) UpdateTokens(accessToken, refreshToken string, expiresAt int64) error {
	err := updateTokens(accessToken, refreshToken, expiresAt)
	if err != nil {
		return err
	}
	k.mu.Lock()
	k.cachedKey = accessToken
	k.lastRefresh = time.Now()
	k.mu.Unlock()
	return nil
}

// RefreshCredentials forces a fresh fetch from keychain
func (k *KeychainCredentialsFetcher) RefreshCredentials() error {
	_, _, err := k.refreshAndGet()
	return err
}

func (k *KeychainCredentialsFetcher) refreshAndGet() (string, string, error) {
	apiKey, userID, err := getCredentials()
	if err != nil {
		return "", "", err
	}
	k.mu.Lock()
	k.cachedKey = apiKey
	k.cachedUser = userID
	k.lastRefresh = time.Now()
	k.mu.Unlock()
	return apiKey, userID, nil
}

func (k *KeychainCredentialsFetcher) backgroundRefresh() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := k.RefreshCredentials()
			if k.logger != nil {
				if err != nil {
					k.logger.Error().Err(err).Msg("Failed to refresh credentials from keychain")
				} else {
					k.logger.Info().Msg("🔄 Refreshed credentials from keychain")
				}
			}
		case <-k.stopCh:
			return
		}
	}
}

// Close stops the background refresh goroutine
func (k *KeychainCredentialsFetcher) Close() {
	close(k.stopCh)
}

// unexported functions, previously in internal/keychain
func getCredentials() (apiKey, userID string, err error) {
	cmd := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve password from Keychain: %w", err)
	}

	var creds keychainCredentials
	if err := json.Unmarshal(output, &creds); err != nil {
		return "", "", fmt.Errorf("failed to parse JSON from keychain: %w", err)
	}

	if creds.ClaudeAiOauth.AccessToken == "" {
		return "", "", fmt.Errorf("accessToken is empty in keychain credentials")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeConfigPath := filepath.Join(homeDir, ".claude.json")
	configData, err := os.ReadFile(claudeConfigPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read ~/.claude.json: %w", err)
	}

	var config claudeConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return "", "", fmt.Errorf("failed to parse ~/.claude.json: %w", err)
	}

	if config.UserID == "" {
		return "", "", fmt.Errorf("userID is empty in ~/.claude.json")
	}

	return creds.ClaudeAiOauth.AccessToken, config.UserID, nil
}

func getFullCredentials() (*fullCredentials, error) {
	cmd := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve password from Keychain: %w", err)
	}

	var creds keychainCredentials
	if err := json.Unmarshal(output, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from keychain: %w", err)
	}

	if creds.ClaudeAiOauth.AccessToken == "" {
		return nil, fmt.Errorf("accessToken is empty in keychain credentials")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeConfigPath := filepath.Join(homeDir, ".claude.json")
	configData, err := os.ReadFile(claudeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ~/.claude.json: %w", err)
	}

	var config claudeConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse ~/.claude.json: %w", err)
	}

	if config.UserID == "" {
		return nil, fmt.Errorf("userID is empty in ~/.claude.json")
	}

	return &fullCredentials{
		AccessToken:  creds.ClaudeAiOauth.AccessToken,
		RefreshToken: creds.ClaudeAiOauth.RefreshToken,
		ExpiresAt:    creds.ClaudeAiOauth.ExpiresAt,
		UserID:       config.UserID,
	}, nil
}

func updateTokens(accessToken, refreshToken string, expiresAt int64) error {
	cmd := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to retrieve current credentials from Keychain: %w", err)
	}

	var creds keychainCredentials
	if err := json.Unmarshal(output, &creds); err != nil {
		return fmt.Errorf("failed to parse current credentials: %w", err)
	}

	creds.ClaudeAiOauth.AccessToken = accessToken
	creds.ClaudeAiOauth.RefreshToken = refreshToken
	creds.ClaudeAiOauth.ExpiresAt = expiresAt

	updatedJSON, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal updated credentials: %w", err)
	}

	deleteCmd := exec.Command("security", "delete-generic-password", "-s", "Claude Code-credentials")
	deleteCmd.Run()

	addCmd := exec.Command("security", "add-generic-password", "-s", "Claude Code-credentials", "-a", "claude-code", "-w", string(updatedJSON), "-U")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to update keychain: %w", err)
	}

	return nil
}
