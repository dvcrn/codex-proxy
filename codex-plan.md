# Codex Proxy Migration Plan

This document outlines goals, constraints, mappings, and implementation steps to migrate the existing Claude Code proxy to translate OpenAI-compatible Chat Completions requests into ChatGPT’s backend Codex Responses API.

## Goals

- Accept `POST /v1/chat/completions` requests (OpenAI format) from clients.
- Translate request body and headers to ChatGPT backend endpoint `https://chatgpt.com/backend-api/codex/responses`.
- Do not forward any original headers; instead, send a fixed set of target headers matching the captured request.
- Preserve streaming behavior (SSE) end-to-end.
- Introduce a filesystem credential provider that reads `/Users/david/.codex/auth.json`.
- Remove Anthropic/Claude-specific transforms and code paths for this proxy variant.

## Inputs & Captures

- Source (OpenAI): `normal_gpt_request.folder/[135] Request - api.openai.com_v1_chat_completions.txt`
- Target (ChatGPT backend): `Raw_08-08-2025-15-37-32.folder.folder/[11] Request - chatgpt.com_backend-api_codex_responses.txt`

These serve as the canonical examples for request shape and headers.

## Credentials

- File: `/Users/david/.codex/auth.json`
- Shape (provided):
  - `tokens.id_token`: string (JWT)
  - `tokens.access_token`: string
  - `tokens.refresh_token`: string
  - `tokens.account_id`: string (UUID)
  - `OPENAI_API_KEY`: null (unused here)

Usage in this proxy:
- Authorization header: Prefer `tokens.id_token` if present; otherwise use `tokens.access_token`.
- `chatgpt-account-id` header: `tokens.account_id`.
- `session_id` header: generated UUID v4 per request.

Implement a new credentials fetcher that returns `(bearerToken, accountID)` for server use.

## Routing

- Server listens on: `POST /v1/chat/completions`.
- Upstream URL: `https://chatgpt.com/backend-api/codex/responses`.
- Health endpoint remains: `/health`.
- Remove or ignore Anthropic routes (e.g., `/v1/messages`, `/v1/models`) for this variant.

## Header Mapping (source → target)

- Do not forward any source headers.
- Set exactly these headers on the upstream request:
  - `authorization: Bearer <jwt>` — from FS credentials (see above)
  - `version: 0.19.0`
  - `openai-beta: responses=experimental`
  - `session_id: <uuid-v4>` — new random each request
  - `accept: text/event-stream`
  - `content-type: application/json`
  - `chatgpt-account-id: <tokens.account_id>`
  - `originator: codex_cli_rs`

Notes:
- Let HTTP client manage `host` and `content-length`.
- Optional: set `accept-encoding` only if needed; default is fine.

## Body Mapping (OpenAI → Codex Responses)

Baseline fields observed in target capture:
- `model`: string — forward from source `model` (validate/allowlist optional).
- `instructions`: string — derive from OpenAI `messages` system content.

Planned translation rules:
- `model`: rewrite to `gpt-5` for all requests (current requirement).
- `messages`:
  - Collect all `role == "system"` contents, concatenate with `\n\n`, and set as `instructions`.
  - For non-system messages (user/assistant): in phase 1, omit (the capture shows a large instructions block; additional fields like conversation state may not be required for initial compatibility). If later testing indicates the backend needs conversation context, add a subsequent `input` or `messages` field matching the Codex schema (to be validated against further captures).
- `temperature`, `top_p`, `stream`, `tools`, etc.: phase 1 — ignore unless capture indicates a corresponding Codex field. We will extend mapping incrementally as needed.

Rationale: The provided target capture shows an `instructions`-heavy body. We start minimally with `model` and `instructions` to reach a working baseline, then iterate.

## Streaming & Response Handling

- Set `accept: text/event-stream` upstream.
- Stream the upstream response body to the client as-is.
- Propagate relevant upstream response headers (e.g., content type), but avoid leaking upstream auth identifiers back to clients.

## Security & Privacy

- Do not log bearer tokens or auth.json contents.
- Log only minimal request metadata (model, sizes) without PII.
- Validate `auth.json` presence and shape; return 503 with clear error if missing.

## Implementation Plan

1) Introduce FS credentials fetcher
- File: `internal/credentials/fs.go`
- Reads `/Users/david/.codex/auth.json` once per request (phase 1), returns `(token, accountID)`.
- Token selection: `id_token` else `access_token`.

2) Update server routes
- Replace `/v1/messages` with `/v1/chat/completions` handler.
- Remove Anthropic-specific handlers (`/v1/models`) in this variant or keep gated behind build tags if needed.

3) Implement upstream request builder
- Build minimal target body: `{ model, instructions }` from source.
- Set target headers exactly as specified.
- Generate `session_id` (`uuidv4`).
- POST to `https://chatgpt.com/backend-api/codex/responses`.

4) Stream responses
- Mirror current streaming implementation to flush chunks to client.
- For non-200, read full body and relay with status.

5) Remove Claude-specific transforms
- Stop using `transform.go`/`prompts.go` and their system/message rewriting.
- Delete or fence with build tags to avoid accidental use.

6) Smoke test locally
- Use the captured OpenAI request body, POST to our `/v1/chat/completions`.
- Verify upstream receives expected headers and that SSE flows back.
- Iterate mapping if upstream requires additional body fields.

## Open Questions / Decisions

- Exact Codex body shape beyond `model` and `instructions`: expand mapping once we confirm needed fields from additional captures or error messages.
- Which token to prefer: assuming `id_token` best matches the captured `authorization` header; fallback to `access_token` if missing.
- Session lifecycle: new `session_id` per request vs. per process — we’ll use per-request for safety.

## Milestones

- M1: FS credentials + new route + minimal mapping works for simple prompt.
- M2: Add conversation context mapping if required by upstream.
- M3: Remove/guard Anthropic code and update README.
