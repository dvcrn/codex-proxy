#!/usr/bin/env fish

# Script to update chainenv and .envrc with Anthropic credentials

echo "Fetching Anthropic API key from Keychain..."

# Fetch the raw JSON string from Keychain
set raw_json_key (security find-generic-password -s "Claude Code-credentials" -w)

if test $status -ne 0
  echo "Error: Could not retrieve password from Keychain."
  echo "Please ensure you have a generic password saved with the service name 'Claude Code-credentials'."
  exit 1
end

if test -z "$raw_json_key"
  echo "Error: Retrieved API key JSON is empty."
  exit 1
end

echo "Parsing API key from JSON..."
# Parse the JSON and extract the accessToken using jq
# The -r flag outputs the raw string without quotes
set api_key (echo $raw_json_key | jq -r '.claudeAiOauth.accessToken')

if test $status -ne 0
  echo "Error: Failed to parse JSON or extract accessToken using jq."
  echo "Please ensure jq is installed and the JSON structure is as expected: {\"claudeAiOauth\":{\"accessToken\":\"...\"}}"
  echo "Raw JSON retrieved: $raw_json_key"
  exit 1
end

if test -z "$api_key"
  echo "Error: Extracted accessToken is empty."
  exit 1
end

if string match -q "null" $api_key
  echo "Error: Extracted accessToken is 'null'. Check JSON path and content."
  echo "Raw JSON retrieved: $raw_json_key"
  exit 1
end

echo "Reading user ID from ~/.claude.json..."
set claude_config_file "$HOME/.claude.json"

if not test -f "$claude_config_file"
  echo "Error: ~/.claude.json not found."
  exit 1
end

# Extract userID from ~/.claude.json
set user_id (cat "$claude_config_file" | jq -r '.userID')

if test $status -ne 0
  echo "Error: Failed to parse ~/.claude.json or extract userID using jq."
  exit 1
end

if test -z "$user_id"
  echo "Error: Extracted userID is empty."
  exit 1
end

if string match -q "null" $user_id
  echo "Error: Extracted userID is 'null'. Check JSON path and content in ~/.claude.json."
  exit 1
end

echo "Found user ID: $user_id"

echo "Storing credentials in chainenv..."

# Store API key in chainenv
chainenv update CLAUDE_CODE_PROXY_API_KEY "$api_key"
if test $status -ne 0
  echo "Error: Failed to store API key in chainenv."
  exit 1
end

# Store user ID in chainenv
chainenv update CLAUDE_CODE_PROXY_USER_ID "$user_id"
if test $status -ne 0
  echo "Error: Failed to store user ID in chainenv."
  exit 1
end

echo "Successfully stored credentials in chainenv."

set envrc_file "./.envrc"

# Write the export commands to .envrc, pulling from chainenv
echo "export ANTHROPIC_API_KEY=\$(chainenv get CLAUDE_CODE_PROXY_API_KEY)" > "$envrc_file"
echo "export CLAUDE_USER_ID=\$(chainenv get CLAUDE_CODE_PROXY_USER_ID)" >> "$envrc_file"

if test $status -eq 0
  echo "Successfully updated $envrc_file to pull credentials from chainenv."
  echo "Run 'direnv allow' in this directory to apply the changes."
else
  echo "Error: Failed to write to $envrc_file."
  exit 1
end
