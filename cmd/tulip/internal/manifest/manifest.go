package manifest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Component represents a single OTel component entry in a manifest.
type Component struct {
	GoMod string `yaml:"gomod"`
}

// Name extracts the component name from the gomod path.
// e.g. "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0" → "otlpreceiver"
func (c Component) Name() string {
	parts := strings.Fields(c.GoMod)
	if len(parts) == 0 {
		return ""
	}
	path := parts[0]
	return filepath.Base(path)
}

// Version extracts the version from the gomod string.
// e.g. "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0" → "v0.145.0"
func (c Component) Version() string {
	parts := strings.Fields(c.GoMod)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// ModPath extracts the module path (without version) from the gomod string.
func (c Component) ModPath() string {
	parts := strings.Fields(c.GoMod)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

// ComponentType returns the type of the component (receiver, processor, exporter, etc.)
// based on the gomod path.
func (c Component) ComponentType() string {
	modPath := c.ModPath()
	for _, t := range []string{"extension", "receiver", "exporter", "processor", "connector", "provider"} {
		if strings.Contains(modPath, "/"+t+"/") || strings.Contains(modPath, "/"+t+"s/") {
			return t
		}
	}
	// Check path segment more broadly
	name := c.Name()
	for _, t := range []string{"extension", "receiver", "exporter", "processor", "connector", "provider"} {
		if strings.HasSuffix(name, t) {
			return t
		}
	}
	return "unknown"
}

// IsCore returns true if the component is from the core OTel collector (not contrib).
func (c Component) IsCore() bool {
	return strings.HasPrefix(c.ModPath(), "go.opentelemetry.io/collector/")
}

// Dist holds the distribution metadata from the manifest.
type Dist struct {
	Module      string `yaml:"module"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	OutputPath  string `yaml:"output_path"`
	BuildTags   string `yaml:"build_tags,omitempty"`
}

// Manifest represents an OCB manifest.yaml file.
type Manifest struct {
	Dist       Dist        `yaml:"dist"`
	Extensions []Component `yaml:"extensions,omitempty"`
	Receivers  []Component `yaml:"receivers,omitempty"`
	Exporters  []Component `yaml:"exporters,omitempty"`
	Processors []Component `yaml:"processors,omitempty"`
	Connectors []Component `yaml:"connectors,omitempty"`
	Providers  []Component `yaml:"providers,omitempty"`
}

// Parse reads and parses a manifest.yaml file.
func Parse(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	return ParseBytes(data)
}

// ParseBytes parses manifest YAML from bytes.
func ParseBytes(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return &m, nil
}

// Write serializes the manifest to a YAML file.
func Write(path string, m *Manifest) error {
	data, err := Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Marshal serializes the manifest to YAML bytes.
func Marshal(m *Manifest) ([]byte, error) {
	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshaling manifest: %w", err)
	}
	return data, nil
}

// AllComponents returns all components across all sections.
func (m *Manifest) AllComponents() []Component {
	var all []Component
	all = append(all, m.Extensions...)
	all = append(all, m.Receivers...)
	all = append(all, m.Exporters...)
	all = append(all, m.Processors...)
	all = append(all, m.Connectors...)
	all = append(all, m.Providers...)
	return all
}

// ComponentCount returns the total number of components.
func (m *Manifest) ComponentCount() int {
	return len(m.Extensions) + len(m.Receivers) + len(m.Exporters) +
		len(m.Processors) + len(m.Connectors) + len(m.Providers)
}

// HasComponent checks if a component with the given name exists in the manifest.
func (m *Manifest) HasComponent(name string) bool {
	for _, c := range m.AllComponents() {
		if c.Name() == name {
			return true
		}
	}
	return false
}

// AddComponent adds a component to the appropriate section based on its type.
// Returns an error if the component already exists.
func (m *Manifest) AddComponent(c Component) error {
	if m.HasComponent(c.Name()) {
		return fmt.Errorf("component %s already exists", c.Name())
	}
	ct := c.ComponentType()
	switch ct {
	case "extension":
		m.Extensions = append(m.Extensions, c)
	case "receiver":
		m.Receivers = append(m.Receivers, c)
	case "exporter":
		m.Exporters = append(m.Exporters, c)
	case "processor":
		m.Processors = append(m.Processors, c)
	case "connector":
		m.Connectors = append(m.Connectors, c)
	case "provider":
		m.Providers = append(m.Providers, c)
	default:
		return fmt.Errorf("unknown component type for %s", c.Name())
	}
	return nil
}

// RemoveComponent removes a component by name. Returns an error if not found.
func (m *Manifest) RemoveComponent(name string) error {
	if removed := removeFrom(&m.Extensions, name); removed {
		return nil
	}
	if removed := removeFrom(&m.Receivers, name); removed {
		return nil
	}
	if removed := removeFrom(&m.Exporters, name); removed {
		return nil
	}
	if removed := removeFrom(&m.Processors, name); removed {
		return nil
	}
	if removed := removeFrom(&m.Connectors, name); removed {
		return nil
	}
	if removed := removeFrom(&m.Providers, name); removed {
		return nil
	}
	return fmt.Errorf("component %s not found", name)
}

func removeFrom(components *[]Component, name string) bool {
	for i, c := range *components {
		if c.Name() == name {
			*components = append((*components)[:i], (*components)[i+1:]...)
			return true
		}
	}
	return false
}

// CategorizeComponents returns which components in this manifest are "base" (present in the
// base manifest) and which are "custom" (user-added).
func CategorizeComponents(dist *Manifest, base *Manifest) (baseComps, customComps []Component) {
	baseNames := make(map[string]bool)
	for _, c := range base.AllComponents() {
		baseNames[c.Name()] = true
	}
	for _, c := range dist.AllComponents() {
		if baseNames[c.Name()] {
			baseComps = append(baseComps, c)
		} else {
			customComps = append(customComps, c)
		}
	}
	return
}

// FindDistributions returns the names of all distributions found under the given
// distributions directory (each subdirectory containing a manifest.yaml).
func FindDistributions(distributionsDir string) ([]string, error) {
	entries, err := os.ReadDir(distributionsDir)
	if err != nil {
		return nil, fmt.Errorf("reading distributions directory: %w", err)
	}
	var dists []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		manifestPath := filepath.Join(distributionsDir, e.Name(), "manifest.yaml")
		if _, err := os.Stat(manifestPath); err == nil {
			dists = append(dists, e.Name())
		}
	}
	return dists, nil
}

// OTelVersion extracts the OTel version used by the manifest by looking at
// the first contrib component's version. Falls back to dist.version.
func (m *Manifest) OTelVersion() string {
	for _, c := range m.AllComponents() {
		v := c.Version()
		if v != "" {
			return v
		}
	}
	return m.Dist.Version
}

// baseManifestYAML is the embedded copy of distributions/tulip/manifest.yaml.
// It is used as a fallback when the file is not found on disk, so the CLI
// works from any directory after `go install`.
var baseManifestYAML = `dist:
  module: github.com/ollygarden/tulip/tulip
  name: tulip
  description: OllyGarden Tulip - Enterprise-supported OpenTelemetry Collector distribution
  version: 0.144.0
  output_path: ./_build
  build_tags: "grpcnotrace"

extensions:
  - gomod: go.opentelemetry.io/collector/extension/zpagesextension v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/extension/bearertokenauthextension v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/extension/oauth2clientauthextension v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/extension/oidcauthextension v0.145.0

receivers:
  - gomod: go.opentelemetry.io/collector/receiver/nopreceiver v0.145.0
  - gomod: go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0

exporters:
  - gomod: go.opentelemetry.io/collector/exporter/debugexporter v0.145.0
  - gomod: go.opentelemetry.io/collector/exporter/nopexporter v0.145.0
  - gomod: go.opentelemetry.io/collector/exporter/otlpexporter v0.145.0
  - gomod: go.opentelemetry.io/collector/exporter/otlphttpexporter v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.145.0

processors:
  - gomod: go.opentelemetry.io/collector/processor/batchprocessor v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.145.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.145.0

connectors:
  - gomod: go.opentelemetry.io/collector/connector/forwardconnector v0.145.0

providers:
  - gomod: go.opentelemetry.io/collector/confmap/provider/envprovider v1.51.0
  - gomod: go.opentelemetry.io/collector/confmap/provider/fileprovider v1.51.0
  - gomod: go.opentelemetry.io/collector/confmap/provider/yamlprovider v1.51.0
`

// LoadBaseManifest tries to read the base manifest from diskPath. If the file
// is not found, it falls back to the embedded copy compiled into the binary.
func LoadBaseManifest(diskPath string) (*Manifest, error) {
	m, err := Parse(diskPath)
	if err == nil {
		return m, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		// File exists but is malformed — surface the real error.
		return nil, err
	}
	return ParseBytes([]byte(baseManifestYAML))
}
