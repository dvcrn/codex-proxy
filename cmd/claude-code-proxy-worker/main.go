//go:build js && wasm

package main

import (
	"github.com/dvcrn/claude-code-proxy/internal/app"
	"github.com/dvcrn/claude-code-proxy/internal/auth"
	"github.com/dvcrn/claude-code-proxy/internal/credentials"
	"github.com/dvcrn/claude-code-proxy/internal/logger"
	"github.com/syumai/workers"
)

func main() {
	// Create logger
	log := logger.New()

	log.Info().Msg("📦 Using Cloudflare KV credentials fetcher with OAuth refresh")
	kvFetcher, err := credentials.NewCloudflareKVFetcher()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Cloudflare KV fetcher")
	}

	// Wrap with OAuth fetcher for automatic token refresh
	oauthFetcher := auth.NewOAuthFetcher(kvFetcher, &log)

	// Create server using OAuth-wrapped fetcher
	srv := app.NewServer(oauthFetcher, log)

	// Serve using workers - it handles all the HTTP server setup
	workers.Serve(srv)
}
