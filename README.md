# Codex Proxy

Go proxy server that forwards OpenAI-compatible requests to the ChatGPT Codex Responses backend.

## Setup

### Credentials Storage & Migration

The proxy now uses **independent credential storage** to avoid token collisions with the system Codex CLI.

**Default behavior (`--creds-store=auto`)**:
- Stores credentials in `~/.config/codex-proxy/auth.json` (XDG config directory)
- On first launch, automatically migrates from:
  1. Legacy file (`~/.codex/auth.json`) if it exists
  2. System Keychain if no legacy file found
- After migration, immediately refreshes tokens to establish an independent token chain
- All subsequent token refreshes are stored in the new location

**Credential store modes**:
```bash
# Auto migration (default) - uses XDG config directory
./codex-proxy --creds-store=auto

# Explicit XDG path
./codex-proxy --creds-store=xdg

# Custom path
./codex-proxy --creds-store=xdg --creds-path=/custom/path/auth.json

# Legacy mode (shares with system CLI)
./codex-proxy --creds-store=legacy --creds-path=~/.codex/auth.json

# Keychain mode (macOS only)
./codex-proxy --creds-store=keychain

# Environment variables mode
./codex-proxy --creds-store=env
```

**Migration flags**:
```bash
# Skip immediate token refresh after migration (not recommended)
./codex-proxy --disable-migrate-refresh
```

**Environment variables** (for `--creds-store=env` mode):
```bash
export ACCESS_TOKEN="your-access-token"
export ACCOUNT_ID="your-account-id"
```

**Server config**:
```bash
export PORT="3000"  # default: 9879
export ENV="production"  # default: development (console logs)
```

**Migration logs**:
The server provides detailed logging during migration:
- `🔍` - Checking for existing credentials
- `📄` - Reading from legacy file or keychain
- `💾` - Writing credentials to new location
- `🔄` - Performing token refresh
- `✅` - Success indicators
- `⚠️` - Warnings (e.g., refresh failures)
- `❌` - Errors

**Troubleshooting**:
- If migration fails, the server will continue with existing credentials if available
- Check logs for detailed error messages
- Use `--creds-store=legacy` to temporarily revert to old behavior
- Manually inspect `~/.config/codex-proxy/auth.json` for credential status

## Usage

```bash
just build  # Build binary
just run    # Run server
just test   # Run tests
```

## Endpoints

- `POST /v1/chat/completions` - OpenAI chat completions-compatible endpoint
- `POST /v1/responses` - OpenAI Responses-compatible endpoint (Codex)
- `GET /health` - Health check

## Example

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"messages": [{"role": "user", "content": "Hello!"}]}'
```

## Cloudflare Workers Deployment

### Prerequisites

1. Create a KV namespace in Cloudflare:
   ```bash
   wrangler kv:namespace create "GEMINI_CLI_KV"
   ```

2. Update `wrangler.toml` with your KV namespace ID and account ID

### Deployment

```bash
# Build and deploy
wrangler deploy

# Set required secrets
wrangler secret put ADMIN_API_KEY  # Enter your admin API key for credential management
```

### Managing Credentials

After deployment, populate credentials in KV storage using the admin API.

#### Setting Credentials

Use the POST `/admin/credentials` endpoint to update tokens:

```bash
curl -X POST https://your-worker.workers.dev/admin/credentials \
  -H "Authorization: Bearer YOUR_ADMIN_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "accessToken": "your-access-token",
    "refreshToken": "your-refresh-token",
    "expiresAt": 1234567890000,
    "userID": "your-user-id"
  }'
```

**Required fields:**
- `accessToken`: Access token for ChatGPT/Codex backend
- `refreshToken`: Refresh token for automatic renewal
- `expiresAt`: Token expiration timestamp in milliseconds (Unix timestamp * 1000)
- `userID` (optional): User identifier for tracking

**Getting tokens:**
- Retrieve your ChatGPT/Codex session tokens from your OpenAI account session (e.g., via DevTools Network panel).
- Ensure `expiresAt` reflects the token expiry in milliseconds.

#### Checking Credential Status

To verify the credentials are properly stored and check their expiration:

```bash
curl https://your-worker.workers.dev/admin/credentials/status \
  -H "Authorization: Bearer YOUR_ADMIN_API_KEY"
```

This returns:
```json
{
  "type": "oauth",
  "hasCredentials": true,
  "userID": "your-user-id",
  "expiresAt": 1234567890000,
  "minutesUntilExpiry": 120,
  "isExpired": false,
  "needsRefreshSoon": false
}
```

**Note:** You can use either `Authorization: Bearer <key>` or `X-API-Key: <key>` headers for authentication.

### Environment Variables for Workers

- `ADMIN_API_KEY` (secret) - Required for accessing admin endpoints
- KV namespace binding - Configured in `wrangler.toml` as `GEMINI_CLI_KV`

### Token Refresh

The worker automatically refreshes tokens when they expire (within 60 minutes of expiry). Refreshed tokens are automatically saved back to KV storage.
