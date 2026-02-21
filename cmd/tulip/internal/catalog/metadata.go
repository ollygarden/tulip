package catalog

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// ContribMetadataBaseURL is the base URL for per-component metadata.yaml files.
	ContribMetadataBaseURL = "https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/main"
)

// ComponentMetadata holds parsed metadata.yaml info for a component.
type ComponentMetadata struct {
	Type      string          `yaml:"type"`
	Status    ComponentStatus `yaml:"status"`
	ShortName string          `yaml:"-"`
}

// ComponentStatus from metadata.yaml.
type ComponentStatus struct {
	Class     string              `yaml:"class"`
	Stability map[string][]string `yaml:"stability"`
}

// StabilityLevel returns the most prominent stability level for this component.
func (m *ComponentMetadata) StabilityLevel() string {
	// Order of preference: stable > beta > alpha > development
	for _, level := range []string{"stable", "beta", "alpha", "development", "deprecated"} {
		if signals, ok := m.Status.Stability[level]; ok && len(signals) > 0 {
			return level
		}
	}
	return "unknown"
}

// FetchMetadata fetches the metadata.yaml for a specific component from the contrib repo.
func FetchMetadata(componentType, name string) (*ComponentMetadata, error) {
	// The contrib repo layout: {type}/{name}/metadata.yaml
	url := fmt.Sprintf("%s/%s/%s/metadata.yaml", ContribMetadataBaseURL, componentType, name)
	return FetchMetadataFromURL(url)
}

// FetchMetadataFromURL fetches and parses a metadata.yaml from the given URL.
func FetchMetadataFromURL(url string) (*ComponentMetadata, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching metadata: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading metadata: %w", err)
	}

	return ParseMetadata(data)
}

// ParseMetadata parses a metadata.yaml file.
func ParseMetadata(data []byte) (*ComponentMetadata, error) {
	var m ComponentMetadata
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing metadata: %w", err)
	}
	return &m, nil
}

// EnrichCatalog adds stability information to catalog entries.
// It fetches metadata for each contrib component. Core components default to "stable".
func EnrichCatalog(c *Catalog) {
	for i, e := range c.Entries {
		if strings.HasPrefix(e.GoMod, "go.opentelemetry.io/") {
			c.Entries[i].Stability = "stable"
			continue
		}
		meta, err := FetchMetadata(e.Type, e.Name)
		if err != nil {
			c.Entries[i].Stability = "unknown"
			continue
		}
		c.Entries[i].Stability = meta.StabilityLevel()
	}
}
