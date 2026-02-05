package client

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/configuration"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

func UsePrivateKeyAuth() bool {
	appIDStr := configuration.GetGithubAppId()
	installationIDStr := configuration.GetGitHubInstallationId()
	privateKeyPath := configuration.GetGitHubPrivateKeyPath()

	return appIDStr != "" && installationIDStr != "" && privateKeyPath != ""
}

func GetOathClient(jwt string) *github.Client {
	ctx := context.Background()

	// Create a token source with the JWT
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: jwt},
	)

	// Create an HTTP client with the token
	tc := oauth2.NewClient(ctx, ts)

	// Create and return the GitHub client
	return github.NewClient(tc)
}

func GetClientLocal() *github.Client {

	// Load environment variables for security
	appIDStr := os.Getenv("GITHUB_APP_ID")
	installationIDStr := os.Getenv("GITHUB_INSTALLATION_ID")
	privateKeyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH")

	appID := mustParseInt64(appIDStr)
	installationID := mustParseInt64(installationIDStr)

	// Create an http.RoundTripper that signs requests as a GitHub App installation
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, installationID, privateKeyPath)
	if err != nil {
		log.Fatalf("Error creating ghinstallation transport: %v", err)
	}

	// Create the GitHub client with the authenticated transport
	client := github.NewClient(&http.Client{Transport: itr})

	// Create GitHub client
	return client
}

func GetClient(accessToken string) *github.Client {
	if UsePrivateKeyAuth() {
		return GetClientLocal()
	}

	return GetOathClient(accessToken)
}

func mustParseInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatalf("Invalid integer: %s", s)
	}
	return i
}
