//go:build !js || !wasm

package env

import "os"

// Get retrieves an environment variable
func Get(key string) (string, bool) {
	value := os.Getenv(key)
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
