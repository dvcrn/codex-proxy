package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/dvcrn/claude-code-proxy/internal/app"
	"github.com/dvcrn/claude-code-proxy/internal/auth"
	"github.com/dvcrn/claude-code-proxy/internal/credentials"
	"github.com/dvcrn/claude-code-proxy/internal/logger"
	"github.com/rs/zerolog"
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
		fsFetcher := credentials.NewFSCredentialsFetcher(*fsPath)
		credsFetcher = auth.NewOAuthFetcher(fsFetcher, &log)
		log.Info().Str("path", *fsPath).Msg("📄 Using filesystem credentials fetcher with OAuth token refresh")
	} else {
		credsFetcher = credentials.NewEnvCredentialsFetcher()
		log.Info().Msg("📝 Using environment credentials fetcher")
	}

	// Validate credentials at startup
	validateCredentialsAtStartup(credsFetcher, log)

	// Create server using shared setup
	srv := app.NewServer(credsFetcher, log)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9879"
	}

	log.Info().Str("port", port).Msg("Starting server")
	log.Fatal().Err(http.ListenAndServe(":"+port, srv)).Msg("Server failed to start")
}

func validateCredentialsAtStartup(credsFetcher credentials.CredentialsFetcher, log zerolog.Logger) {
	// Try to get basic credentials
	token, userID, err := credsFetcher.GetCredentials()
	if err != nil {
		log.Error().Err(err).Msg("⚠️  Failed to validate credentials at startup")
		return
	}

	log.Info().
		Str("user_id", userID).
		Int("token_length", len(token)).
		Msg("✅ Credentials loaded successfully")

	// Check if this is an OAuth fetcher with expiry information
	if oauthFetcher, ok := credsFetcher.(credentials.OAuthCredentialsFetcher); ok {
		creds, err := oauthFetcher.GetFullCredentials()
		if err != nil {
			log.Warn().Err(err).Msg("⚠️  Could not get full OAuth credentials for validation")
			return
		}

		// Calculate time until expiry
		now := auth.UnixMillis()
		minutesUntilExpiry := (creds.ExpiresAt - now) / 1000 / 60

		if minutesUntilExpiry <= 0 {
			log.Warn().
				Int64("minutes_expired", -minutesUntilExpiry).
				Msg("⚠️  Token is already expired, will attempt refresh on first request")
		} else if minutesUntilExpiry <= 60 {
			log.Warn().
				Int64("minutes_until_expiry", minutesUntilExpiry).
				Msg("⚠️  Token expires soon, will refresh shortly")
		} else {
			log.Info().
				Int64("minutes_until_expiry", minutesUntilExpiry).
				Msg("✅ Token is valid and not expiring soon")
		}
	}
}
