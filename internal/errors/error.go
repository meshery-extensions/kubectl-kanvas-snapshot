package errors

import (
	"fmt"

	"github.com/layer5io/meshkit/errors"
)

var (
	ErrInvalidManifestURICode       = "kanvas-snapshot-900"
	ErrCreatingMesheryDesignCode    = "kanvas-snapshot-901"
	ErrGeneratingSnapshotCode       = "kanvas-snapshot-902"
	ErrHTTPPostRequestCode          = "kanvas-snapshot-903"
	ErrReadingManifestCode          = "kanvas-snapshot-904"
	ErrDecodingAPICode              = "kanvas-snapshot-905"
	ErrUnexpectedResponseCodeCode   = "kanvas-snapshot-906"
	ErrRequiredFieldNotProvidedCode = "kanvas-snapshot-907"
	ErrInvalidEmailFormatCode       = "kanvas-snapshot-908"
	ErrGitHubAuthCode               = "kanvas-snapshot-909"
	ErrDownloadingAssetCode         = "kanvas-snapshot-910"
)

func ErrInvalidManifestURI(err error) error {
	return errors.New(ErrInvalidManifestURICode, errors.Alert,
		[]string{"Invalid or missing Kubernetes manifest path or URI."},
		[]string{err.Error()},
		[]string{"The provided file path or URI for the Kubernetes manifest is either missing or invalid."},
		[]string{"Ensure the Kubernetes manifest file or URI exists and is correctly provided using -f / --file."},
	)
}

func ErrReadingManifest(err error) error {
	return errors.New(ErrReadingManifestCode, errors.Alert,
		[]string{"Failed to read Kubernetes manifest file(s)."},
		[]string{err.Error()},
		[]string{"One or more manifest files could not be read from the local file system."},
		[]string{"Check file permissions and ensure the specified path(s) exist."},
	)
}

func ErrCreatingMesheryDesign(err error) error {
	return errors.New(ErrCreatingMesheryDesignCode, errors.Alert,
		[]string{"Failed to create Meshery design."},
		[]string{fmt.Sprintf("Error creating Meshery design: %s", err.Error())},
		[]string{"Meshery Design creation failed due to an error during import."},
		[]string{"Check Meshery API connection and ensure the manifest contents are valid Kubernetes resources."},
	)
}

func ErrGeneratingSnapshot(err error) error {
	return errors.New(ErrGeneratingSnapshotCode, errors.Alert,
		[]string{"Failed to generate snapshot."},
		[]string{err.Error()},
		[]string{"Snapshot generation failed due to an error."},
		[]string{"Check Meshery Cloud API connection and GitHub workflow credentials."},
	)
}

func ErrHTTPPostRequest(err error) error {
	return errors.New(ErrHTTPPostRequestCode, errors.Alert,
		[]string{"Failed to perform HTTP POST request."},
		[]string{err.Error()},
		[]string{"HTTP POST request failed during interaction with Meshery API."},
		[]string{"Check Meshery API endpoint and ensure the server is reachable."},
	)
}

func ErrDecodingAPI(err error) error {
	return errors.New(ErrDecodingAPICode, errors.Alert,
		[]string{"Failed to decode API response."},
		[]string{err.Error()},
		[]string{"API response could not be decoded into the expected JSON format."},
		[]string{"Ensure the Meshery API response format is compatible."},
	)
}

func ErrUnexpectedResponseCode(statusCode int, body string) error {
	return errors.New(ErrUnexpectedResponseCodeCode, errors.Alert,
		[]string{"Received unexpected response code from Meshery API."},
		[]string{fmt.Sprintf("Status Code: %d, Body: %s", statusCode, body)},
		[]string{"The API returned an unexpected HTTP status code."},
		[]string{"Check the request details and ensure valid credentials and manifest payloads."},
	)
}

func ErrRequiredFieldNotProvided(err error, field string) error {
	return errors.New(ErrRequiredFieldNotProvidedCode, errors.Alert,
		[]string{"All required flags are not passed."},
		[]string{err.Error()},
		[]string{fmt.Sprintf("Required flag \"%s\" is not passed.", field)},
		[]string{fmt.Sprintf("Ensure a valid value for flag \"%s\" is provided.", field)},
	)
}

func ErrInvalidEmailFormat(email string) error {
	return errors.New(ErrInvalidEmailFormatCode, errors.Alert,
		[]string{"Invalid email format provided."},
		[]string{fmt.Sprintf("The provided email '%s' is not in a valid email format.", email)},
		[]string{"The email address provided for notification is not syntactically valid."},
		[]string{"Ensure the email address follows standard format (e.g., user@example.com)."},
	)
}

func ErrGitHubAuth(body string) error {
	return errors.New(ErrGitHubAuthCode, errors.Alert,
		[]string{"GitHub API authentication failed."},
		[]string{fmt.Sprintf("GitHub API returned authentication error. Body: %s", body)},
		[]string{"GitHub workflow dispatch request failed authentication."},
		[]string{"Check that your GitHub workflow access token is valid and has appropriate permissions."},
	)
}

func ErrDownloadingAsset(err error) error {
	return errors.New(ErrDownloadingAssetCode, errors.Alert,
		[]string{"Failed to download snapshot image asset."},
		[]string{err.Error()},
		[]string{"The generated snapshot image could not be downloaded to the specified local path."},
		[]string{"Verify write permissions to the destination directory or check the snapshot URL."},
	)
}
