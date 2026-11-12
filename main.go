package main

import (
	"os"

	cmd "github.com/meshery-extensions/kubectl-kanvas-snapshot/cmd/kanvas-snapshot"
)

var (
	providerToken          string
	mesheryCloudAPIBaseURL string
	mesheryAPIBaseURL      string
	workflowAccessToken    string
)

func main() {
	// Fall back to environment variables if not set during compilation via ldflags
	if providerToken == "" {
		providerToken = os.Getenv("PROVIDER_TOKEN")
	}
	if mesheryCloudAPIBaseURL == "" {
		mesheryCloudAPIBaseURL = os.Getenv("MESHERY_CLOUD_API_BASE_URL")
		if mesheryCloudAPIBaseURL == "" {
			mesheryCloudAPIBaseURL = "https://cloud.layer5.io"
		}
	}
	if mesheryAPIBaseURL == "" {
		mesheryAPIBaseURL = os.Getenv("MESHERY_API_BASE_URL")
		if mesheryAPIBaseURL == "" {
			mesheryAPIBaseURL = "https://playground.meshery.io"
		}
	}
	if workflowAccessToken == "" {
		workflowAccessToken = os.Getenv("GH_ACCESS_TOKEN")
		if workflowAccessToken == "" {
			workflowAccessToken = os.Getenv("GITHUB_TOKEN")
		}
	}

	cmd.Main(providerToken, mesheryCloudAPIBaseURL, mesheryAPIBaseURL, workflowAccessToken)
}
