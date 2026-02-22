package catalog

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

const upstreamURL = "https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/main/cmd/otelcontribcol/builder-config.yaml"

// Component represents a single OTel component entry.
type Component struct {
	GoMod   string // full module path, e.g. "github.com/open-telemetry/.../kafkareceiver"
	Name    string // short name extracted from path, e.g. "kafkareceiver"
	Version string // e.g. "v0.146.0"
}

// Catalog holds all components grouped by type.
type Catalog struct {
	Extensions []Component
	Receivers  []Component
	Processors []Component
	Exporters  []Component
	Connectors []Component
	Providers  []Component
}

// raw mirrors the YAML structure of builder-config.yaml.
type raw struct {
	Extensions []rawEntry `yaml:"extensions"`
	Receivers  []rawEntry `yaml:"receivers"`
	Processors []rawEntry `yaml:"processors"`
	Exporters  []rawEntry `yaml:"exporters"`
	Connectors []rawEntry `yaml:"connectors"`
	Providers  []rawEntry `yaml:"providers"`
}

type rawEntry struct {
	GoMod string `yaml:"gomod"`
}

// Fetch downloads the upstream builder-config.yaml and parses it into a Catalog.
func Fetch() (*Catalog, error) {
	resp, err := http.Get(upstreamURL)
	if err != nil {
		return nil, fmt.Errorf("fetching upstream catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from upstream", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return Parse(body)
}

// Parse parses raw YAML bytes into a Catalog.
func Parse(data []byte) (*Catalog, error) {
	var r raw
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parsing catalog YAML: %w", err)
	}

	return &Catalog{
		Extensions: parseEntries(r.Extensions),
		Receivers:  parseEntries(r.Receivers),
		Processors: parseEntries(r.Processors),
		Exporters:  parseEntries(r.Exporters),
		Connectors: parseEntries(r.Connectors),
		Providers:  parseEntries(r.Providers),
	}, nil
}

func parseEntries(entries []rawEntry) []Component {
	components := make([]Component, 0, len(entries))
	for _, e := range entries {
		c, ok := parseGoMod(e.GoMod)
		if !ok {
			continue
		}
		components = append(components, c)
	}
	return components
}

func parseGoMod(gomod string) (Component, bool) {
	// Format: "github.com/open-telemetry/.../kafkareceiver v0.146.0"
	parts := strings.Fields(gomod)
	if len(parts) < 2 {
		return Component{}, false
	}

	modulePath := parts[0]
	version := parts[1]
	name := path.Base(modulePath)

	return Component{
		GoMod:   modulePath,
		Name:    name,
		Version: version,
	}, true
}
