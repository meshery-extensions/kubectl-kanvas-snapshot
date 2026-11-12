package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/layer5io/meshkit/logger"
	"github.com/meshery-extensions/kubectl-kanvas-snapshot/internal/errors"
	"github.com/meshery-extensions/kubectl-kanvas-snapshot/internal/log"
	"github.com/spf13/cobra"
)

var (
	ProviderToken          string
	MesheryAPIBaseURL      string
	MesheryCloudAPIBaseURL string
	WorkflowAccessToken    string
	Log                    logger.Handler
)

var (
	manifestFiles []string
	email         string
	designName    string
	outputPath    string
	overrideURL   string
	overrideToken string
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

var generateKanvasSnapshotCmd = &cobra.Command{
	Use:   "kanvas-snapshot",
	Short: "Generate a Kanvas snapshot from Kubernetes manifests",
	Long: `Generate a Kanvas visual snapshot in Meshery using Kubernetes manifests.

This plugin bridges Kubernetes configurations with their visual representation in Meshery UI.
It allows users to generate a visual snapshot of a single manifest, a collection of manifests,
or an entire directory of Kubernetes resources.

Example usage:

  # Generate snapshot from a single manifest file
  kubectl kanvas-snapshot -f ./deployment.yaml

  # Generate snapshot from multiple manifests
  kubectl kanvas-snapshot -f ./deployment.yaml -f ./service.yaml

  # Generate snapshot from a directory of manifests
  kubectl kanvas-snapshot -f ./manifests/

  # Generate snapshot from a remote URL with email notification
  kubectl kanvas-snapshot -f https://raw.githubusercontent.com/kubernetes/website/main/content/en/examples/application/deployment.yaml -e user@example.com --name my-app

  # Generate snapshot from stdin
  cat deployment.yaml | kubectl kanvas-snapshot -f -

Flags:
  -f, --file     strings   Path(s) or URI(s) to Kubernetes manifest file(s), directory, or '-' for stdin (required)
  -e, --email    string    Email address to notify when snapshot is ready (optional)
      --name     string    Custom name for the Meshery design (optional)
  -o, --output   string    Save the snapshot image to a local file path (optional)
      --meshery-url string Override Meshery API base URL (optional)
      --token    string    Override Meshery provider authentication token (optional)
  -h, --help               Help for kanvas-snapshot plugin`,

	RunE: func(_ *cobra.Command, _ []string) error {
		if overrideURL != "" {
			MesheryAPIBaseURL = strings.TrimRight(overrideURL, "/")
		}
		if overrideToken != "" {
			ProviderToken = overrideToken
		}

		if len(manifestFiles) == 0 {
			err := errors.ErrRequiredFieldNotProvided(fmt.Errorf("missing -f/--file flag"), "file")
			handleError(err)
			return err
		}

		if email != "" && !IsValidEmail(email) {
			err := errors.ErrInvalidEmailFormat(email)
			handleError(err)
			return err
		}

		// Check if first argument is a remote URL
		firstFile := manifestFiles[0]
		isRemoteURL := IsURL(firstFile)

		if designName == "" {
			if isRemoteURL {
				designName = ExtractNameFromURI(firstFile)
			} else if firstFile == "-" {
				designName = "stdin-manifest"
			} else {
				designName = ExtractNameFromPath(firstFile)
			}
			Log.Warnf("No design name provided. Using derived name: %s", designName)
		}

		var designID string
		var err error

		if isRemoteURL && len(manifestFiles) == 1 {
			// Remote URL Import
			designID, err = CreateMesheryDesignFromURL(firstFile, designName, email)
			if err != nil {
				handleError(errors.ErrCreatingMesheryDesign(err))
				return err
			}
		} else {
			// Local files, directory, stdin, or multiple files
			combinedBytes, derivedFileName, err := AggregateManifests(manifestFiles)
			if err != nil {
				handleError(errors.ErrReadingManifest(err))
				return err
			}
			designID, err = CreateMesheryDesignFromBytes(combinedBytes, derivedFileName, designName, email)
			if err != nil {
				handleError(errors.ErrCreatingMesheryDesign(err))
				return err
			}
		}

		assetLocation := fmt.Sprintf("https://raw.githubusercontent.com/layer5labs/meshery-extensions-packages/master/action-assets/kubectl-plugin-assets/%s.png", designID)

		if WorkflowAccessToken != "" {
			err = GenerateSnapshot(designID, assetLocation, email, WorkflowAccessToken)
			if err != nil {
				Log.Warnf("Workflow dispatch notice: %v", err)
			}
		}

		Log.Infof("\n✓ Design created successfully! Design ID: %s", designID)
		Log.Infof("Snapshot URL: %s", assetLocation)
		Log.Infof("It may take 1-3 minutes for the Kanvas snapshot to render at the above URL.")

		if email != "" {
			Log.Infof("A notification will be sent to %s once rendering completes.", email)
		}

		if outputPath != "" {
			Log.Infof("Downloading snapshot image to %s...", outputPath)
			err = DownloadAssetWithRetry(assetLocation, outputPath, 6, 10*time.Second)
			if err != nil {
				Log.Warnf("Could not immediately download image (may still be rendering): %v", err)
				Log.Infof("You can download it later from: %s", assetLocation)
			} else {
				Log.Infof("✓ Saved snapshot image to %s", outputPath)
			}
		}

		return nil
	},
}

// MesheryDesignURLPayload represents URL-based design import payload
type MesheryDesignURLPayload struct {
	Save  bool   `json:"save"`
	URL   string `json:"url"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// MesheryDesignFilePayload represents File-based design import payload
type MesheryDesignFilePayload struct {
	Save     bool   `json:"save"`
	File     []byte `json:"file"`
	FileName string `json:"fileName"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
}

// IsURL checks if a string is an HTTP/HTTPS URL
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

// ExtractNameFromURI derives a clean name from a URL
func ExtractNameFromURI(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return "manifest"
	}
	base := filepath.Base(u.Path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if name == "" || name == "." || name == "/" {
		return "manifest"
	}
	return name
}

// ExtractNameFromPath derives a clean name from a local path
func ExtractNameFromPath(path string) string {
	clean := filepath.Clean(path)
	base := filepath.Base(clean)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if name == "" || name == "." || name == "/" {
		return "k8s-manifest"
	}
	return name
}

// IsValidEmail verifies email syntax
func IsValidEmail(e string) bool {
	return emailRegex.MatchString(e)
}

// AggregateManifests reads single/multiple files, directories, or stdin into a combined YAML document
func AggregateManifests(paths []string) ([]byte, string, error) {
	var combined bytes.Buffer
	derivedName := "manifest.yaml"

	for _, p := range paths {
		if p == "-" {
			// Read from stdin
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read from stdin: %w", err)
			}
			appendYAMLDocument(&combined, data)
			continue
		}

		info, err := os.Stat(p)
		if err != nil {
			return nil, "", fmt.Errorf("failed to stat %s: %w", p, err)
		}

		if info.IsDir() {
			// Walk directory for YAML/JSON files
			err = filepath.Walk(p, func(path string, f os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if f.IsDir() {
					return nil
				}
				ext := strings.ToLower(filepath.Ext(path))
				if ext == ".yaml" || ext == ".yml" || ext == ".json" {
					content, rErr := os.ReadFile(path)
					if rErr != nil {
						return rErr
					}
					appendYAMLDocument(&combined, content)
				}
				return nil
			})
			if err != nil {
				return nil, "", fmt.Errorf("failed reading directory %s: %w", p, err)
			}
		} else {
			// Single file
			content, rErr := os.ReadFile(p)
			if rErr != nil {
				return nil, "", fmt.Errorf("failed to read file %s: %w", p, rErr)
			}
			derivedName = filepath.Base(p)
			appendYAMLDocument(&combined, content)
		}
	}

	if combined.Len() == 0 {
		return nil, "", fmt.Errorf("no manifest content found in the provided path(s)")
	}

	return combined.Bytes(), derivedName, nil
}

func appendYAMLDocument(buf *bytes.Buffer, doc []byte) {
	trimmed := bytes.TrimSpace(doc)
	if len(trimmed) == 0 {
		return
	}
	if buf.Len() > 0 {
		buf.WriteString("\n---\n")
	}
	buf.Write(trimmed)
}

// CreateMesheryDesignFromURL imports a remote URL into Meshery
func CreateMesheryDesignFromURL(manifestURL, name, email string) (string, error) {
	payload := MesheryDesignURLPayload{
		Save:  true,
		URL:   manifestURL,
		Name:  name,
		Email: email,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", errors.ErrDecodingAPI(err)
	}
	return postDesignImport(payloadBytes)
}

// CreateMesheryDesignFromBytes imports raw manifest bytes into Meshery
func CreateMesheryDesignFromBytes(manifestBytes []byte, fileName, name, email string) (string, error) {
	payload := MesheryDesignFilePayload{
		Save:     true,
		File:     manifestBytes,
		FileName: fileName,
		Name:     name,
		Email:    email,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", errors.ErrDecodingAPI(err)
	}
	return postDesignImport(payloadBytes)
}

func postDesignImport(payloadBytes []byte) (string, error) {
	apiEndpoint := fmt.Sprintf("%s/api/pattern/import", MesheryAPIBaseURL)
	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", errors.ErrHTTPPostRequest(err)
	}

	req.Header.Set("Content-Type", "application/json")
	if ProviderToken != "" {
		req.Header.Set("Cookie", fmt.Sprintf("token=%s;meshery-provider=Layer5", ProviderToken))
	}
	req.Header.Set("Origin", MesheryAPIBaseURL)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.ErrHTTPPostRequest(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.ErrDecodingAPI(err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", errors.ErrUnexpectedResponseCode(resp.StatusCode, string(body))
	}

	// Response can be a slice of objects or a single object
	var sliceResult []map[string]interface{}
	if jsonErr := json.Unmarshal(body, &sliceResult); jsonErr == nil && len(sliceResult) > 0 {
		if id, ok := sliceResult[0]["id"].(string); ok && id != "" {
			return id, nil
		}
	}

	var mapResult map[string]interface{}
	if jsonErr := json.Unmarshal(body, &mapResult); jsonErr == nil {
		if id, ok := mapResult["id"].(string); ok && id != "" {
			return id, nil
		}
	}

	return "", errors.ErrCreatingMesheryDesign(fmt.Errorf("failed to extract design ID from response: %s", string(body)))
}

// GenerateSnapshot dispatches the GitHub workflow for snapshot rendering
func GenerateSnapshot(contentID, assetLocation, email, workflowAccessToken string) error {
	workflowURL := "https://api.github.com/repos/meshery-extensions/kubectl-kanvas-snapshot/actions/workflows/meshmap.yml/dispatches"
	payload := fmt.Sprintf(`{"ref":"master","inputs":{"contentID":"%s","assetLocation":"%s","email":"%s"}}`, contentID, assetLocation, email)

	req, err := http.NewRequest("POST", workflowURL, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+workflowAccessToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		return errors.ErrGitHubAuth(string(body))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d from workflow dispatch: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DownloadAssetWithRetry attempts to download the snapshot image with retry intervals
func DownloadAssetWithRetry(url, destPath string, attempts int, interval time.Duration) error {
	client := &http.Client{Timeout: 15 * time.Second}

	for i := 1; i <= attempts; i++ {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			out, cErr := os.Create(destPath)
			if cErr != nil {
				return cErr
			}
			defer out.Close()
			_, err = io.Copy(out, resp.Body)
			return err
		}
		if resp != nil {
			resp.Body.Close()
		}
		if i < attempts {
			time.Sleep(interval)
		}
	}

	return fmt.Errorf("snapshot image is still processing or not yet available at %s", url)
}

func handleError(err error) {
	if err == nil {
		return
	}
	if Log != nil {
		Log.Error(err)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

// Main is the entry point for the CLI command
func Main(providerToken, mesheryCloudAPIBaseURL, mesheryAPIBaseURL, workflowAccessToken string) {
	ProviderToken = providerToken
	MesheryCloudAPIBaseURL = mesheryCloudAPIBaseURL
	MesheryAPIBaseURL = mesheryAPIBaseURL
	WorkflowAccessToken = workflowAccessToken

	generateKanvasSnapshotCmd.Flags().StringSliceVarP(&manifestFiles, "file", "f", []string{}, "Path(s) or URI(s) to Kubernetes manifest file(s), directory, or '-' for stdin (required)")
	generateKanvasSnapshotCmd.Flags().StringVar(&designName, "name", "", "Custom name for the Meshery design")
	generateKanvasSnapshotCmd.Flags().StringVarP(&email, "email", "e", "", "Email address to notify when snapshot is ready")
	generateKanvasSnapshotCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Local path to save the snapshot image (optional)")
	generateKanvasSnapshotCmd.Flags().StringVar(&overrideURL, "meshery-url", "", "Override Meshery API endpoint")
	generateKanvasSnapshotCmd.Flags().StringVar(&overrideToken, "token", "", "Override Meshery provider token")

	_ = generateKanvasSnapshotCmd.MarkFlagRequired("file")

	if err := generateKanvasSnapshotCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		Log = log.SetupMeshkitLogger("kubectl-kanvas-snapshot", false, os.Stdout)
	})
}
