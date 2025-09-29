# Credentials Management Refactor

## Overview

This refactor introduces a unified credential management system that automatically handles token refresh when receiving 401 errors from the Anthropic API, particularly when using the `-use-keychain` flag.

## Key Changes

### 1. New CredentialsFetcher Interface

Created `internal/credentials/fetcher.go` with a clean interface:

```go
type CredentialsFetcher interface {
    GetCredentials() (apiKey, userID string, err error)
    RefreshCredentials() error
}
```

### 2. Two Implementations

**EnvCredentialsFetcher**

- Reads credentials from environment variables
- `RefreshCredentials()` is a no-op (can't refresh env vars)
- Used when `-use-keychain` flag is not set

**KeychainCredentialsFetcher**

- Reads from macOS keychain with intelligent caching
- Caches credentials for 5 minutes to avoid excessive keychain calls
- Background goroutine refreshes credentials every 10 minutes
- `RefreshCredentials()` forces immediate fresh fetch
- Used when `-use-keychain` flag is set

### 3. Server Refactor

**Before:**

- Server stored credentials in struct fields
- Used mutex for thread-safe credential updates
- No automatic retry on 401 errors

**After:**

- Server uses injected CredentialsFetcher
- Calls `fetcher.GetCredentials()` for each request
- Automatic 401 retry logic:
  1. Make request with current credentials
  2. If 401 received, call `fetcher.RefreshCredentials()`
  3. Get fresh credentials and retry once
  4. Log success/failure appropriately

### 4. Simplified Main Function

**Before:**

- Complex credential initialization
- Manual background goroutine for keychain refresh
- Direct credential updates to server

**After:**

- Simple fetcher selection based on flag
- All refresh logic encapsulated in fetcher
- Clean dependency injection

## Benefits

1. **Automatic Recovery**: 401 errors trigger immediate credential refresh and retry
2. **Clean Architecture**: Single responsibility principle - fetchers handle credentials, server handles requests
3. **Testable**: Easy to mock CredentialsFetcher for unit tests
4. **Extensible**: Can easily add new credential sources (files, vaults, etc.)
5. **Performant**: Intelligent caching avoids excessive keychain calls
6. **Thread-Safe**: All credential access is properly synchronized

## Usage

### Environment Variables Mode (Default)

```bash
./claude-code-proxy
# Uses ANTHROPIC_API_KEY and CLAUDE_USER_ID environment variables
```

### Keychain Mode

```bash
./claude-code-proxy -use-keychain
# Automatically fetches and refreshes credentials from macOS keychain
# Handles 401 errors with automatic token refresh
```

## Error Handling

- **401 Errors**: Automatic refresh + retry (once per request)
- **Keychain Errors**: Logged and handled gracefully
- **Network Errors**: Passed through to client as before

## Logging

- Credential refresh attempts are logged
- 401 retry logic is logged for debugging
- Background refresh success/failure is logged
- Fetcher type selection is logged on startup

## Testing

Basic tests included in `internal/credentials/fetcher_test.go` to verify:

- Environment fetcher functionality
- Keychain fetcher initialization
- No-op refresh for environment fetcher
