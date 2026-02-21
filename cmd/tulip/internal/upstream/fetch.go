package upstream

import (
	"fmt"
	"io"
	"net/http"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

const (
	// UpstreamManifestURL is the raw GitHub URL for the upstream tulip manifest.
	UpstreamManifestURL = "https://raw.githubusercontent.com/ollygarden/tulip/main/distributions/tulip/manifest.yaml"
)

// FetchUpstreamManifest fetches the latest manifest from the upstream ollygarden/tulip repo.
func FetchUpstreamManifest() (*manifest.Manifest, error) {
	return FetchManifestFromURL(UpstreamManifestURL)
}

// FetchManifestFromURL fetches and parses a manifest from the given URL.
func FetchManifestFromURL(url string) (*manifest.Manifest, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching upstream manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching upstream manifest: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading upstream manifest: %w", err)
	}

	return manifest.ParseBytes(data)
}
