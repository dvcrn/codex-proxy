package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/dvcrn/claude-code-proxy/internal/app"
	"github.com/dvcrn/claude-code-proxy/internal/auth"
	"github.com/dvcrn/claude-code-proxy/internal/credentials"
	"github.com/dvcrn/claude-code-proxy/internal/logger"
)

func main() {
	useKeychain := flag.Bool("use-keychain", false, "Extract credentials from macOS keychain")
	useFS := flag.Bool("use-fs-creds", true, "Use filesystem credentials at /Users/david/.codex/auth.json")
	fsPath := flag.String("fs-creds-path", "/Users/david/.codex/auth.json", "Path to filesystem credentials auth.json")
	flag.Parse()

	log := logger.New()

	var credsFetcher credentials.CredentialsFetcher
	if *useKeychain {
		keychainFetcher := credentials.NewKeychainCredentialsFetcherWithLogger(log)
		credsFetcher = auth.NewOAuthFetcher(keychainFetcher, &log)
		log.Info().Msg("🔑 Using keychain credentials fetcher with OAuth token refresh")
	} else if *useFS {
		credsFetcher = credentials.NewFSCredentialsFetcher(*fsPath)
		log.Info().Str("path", *fsPath).Msg("📄 Using filesystem credentials fetcher")
	} else {
		credsFetcher = credentials.NewEnvCredentialsFetcher()
		log.Info().Msg("📝 Using environment credentials fetcher")
	}

	// Create server using shared setup
	srv := app.NewServer(credsFetcher, log)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9879"
	}

	log.Info().Str("port", port).Msg("Starting server")
	log.Fatal().Err(http.ListenAndServe(":"+port, srv)).Msg("Server failed to start")
}
