//go:build js && wasm

package env

import "github.com/syumai/workers/cloudflare"

// Get retrieves an environment variable from Cloudflare Workers environment
func Get(key string) (string, bool) {
	value := cloudflare.Getenv(key)
	if value == "" {
		return "", false
	}
	return value, true
}

// GetOrDefault retrieves an environment variable with a default value
func GetOrDefault(key, defaultValue string) string {
	if value, ok := Get(key); ok {
		return value
	}
	return defaultValue
}
