# Codex Proxy

Go proxy server that forwards OpenAI-compatible requests to the ChatGPT Codex Responses backend.

## Setup

Credentials can be provided via the admin API (recommended) or environment variables.

Environment variables (optional):
```bash
export ACCESS_TOKEN="your-access-token"
export ACCOUNT_ID="your-account-id"
```

Optional server config:
```bash
export PORT="3000"  # default: 8080
export ENV="production"  # default: development (console logs)
```

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
