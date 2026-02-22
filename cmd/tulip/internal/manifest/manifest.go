package manifest

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	tulip "github.com/ollygarden/tulip"
	"gopkg.in/yaml.v3"
)

const upstreamManifestURL = "https://raw.githubusercontent.com/ollygarden/tulip/main/distributions/tulip/manifest.yaml"

// Manifest represents an OCB manifest.yaml.
type Manifest struct {
	Dist       Dist    `yaml:"dist"`
	Extensions []Entry `yaml:"extensions,omitempty"`
	Receivers  []Entry `yaml:"receivers,omitempty"`
	Exporters  []Entry `yaml:"exporters,omitempty"`
	Processors []Entry `yaml:"processors,omitempty"`
	Connectors []Entry `yaml:"connectors,omitempty"`
	Providers  []Entry `yaml:"providers,omitempty"`
}

// Dist holds the distribution metadata.
type Dist struct {
	Module      string `yaml:"module"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	OutputPath  string `yaml:"output_path"`
	BuildTags   string `yaml:"build_tags,omitempty"`
}

// Entry represents a single component entry.
type Entry struct {
	GoMod string `yaml:"gomod"`
}

// LoadBase returns the base tulip manifest embedded in the binary.
// This works regardless of where the CLI is run — the manifest is
// baked in at build time from distributions/tulip/manifest.yaml.
func LoadBase() (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(tulip.BaseManifestYAML, &m); err != nil {
		return nil, fmt.Errorf("parsing embedded base manifest: %w", err)
	}
	return &m, nil
}

// GoModSet returns a set of all gomod module paths (without version) in the manifest.
func (m *Manifest) GoModSet() map[string]bool {
	set := make(map[string]bool)
	for _, sections := range [][]Entry{m.Extensions, m.Receivers, m.Exporters, m.Processors, m.Connectors, m.Providers} {
		for _, e := range sections {
			mod, _ := splitGoMod(e.GoMod)
			set[mod] = true
		}
	}
	return set
}

// FetchUpstream downloads the latest base manifest from the ollygarden/tulip main branch.
func FetchUpstream() (*Manifest, error) {
	resp, err := http.Get(upstreamManifestURL)
	if err != nil {
		return nil, fmt.Errorf("fetching upstream manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from upstream", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading upstream manifest: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("parsing upstream manifest: %w", err)
	}
	return &m, nil
}

// FindLocalDistributions scans ./distributions/ relative to the current
// working directory for all manifest.yaml files, returning them keyed by
// distribution name.
func FindLocalDistributions() (map[string]*Manifest, error) {
	entries, err := os.ReadDir("distributions")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no distributions/ directory found in the current directory\n\nRun this command from your project root, or create a distribution first:\n  tulip create <name>")
		}
		return nil, fmt.Errorf("reading distributions directory: %w", err)
	}

	result := make(map[string]*Manifest)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join("distributions", e.Name(), "manifest.yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			continue // skip directories without manifest.yaml
		}
		var m Manifest
		if err := yaml.Unmarshal(data, &m); err != nil {
			continue
		}
		result[e.Name()] = &m
	}

	return result, nil
}

// GoModVersionMap returns a map of module path → version for all components.
func (m *Manifest) GoModVersionMap() map[string]string {
	versions := make(map[string]string)
	for _, sections := range [][]Entry{m.Extensions, m.Receivers, m.Exporters, m.Processors, m.Connectors, m.Providers} {
		for _, e := range sections {
			mod, ver := splitGoMod(e.GoMod)
			versions[mod] = ver
		}
	}
	return versions
}

// Generate creates a new manifest for a distribution with base components + custom selections.
// customGoMods maps section name ("extensions","receivers",...) to a list of "module version" strings.
func Generate(name string, base *Manifest, customGoMods map[string][]string) *Manifest {
	return &Manifest{
		Dist: Dist{
			Module:      fmt.Sprintf("github.com/ollygarden/tulip/%s", name),
			Name:        name,
			Description: fmt.Sprintf("Custom OpenTelemetry Collector distribution: %s", name),
			Version:     base.Dist.Version,
			OutputPath:  "./_build",
			BuildTags:   base.Dist.BuildTags,
		},
		Extensions: mergeEntries(base.Extensions, customGoMods["extensions"]),
		Receivers:  mergeEntries(base.Receivers, customGoMods["receivers"]),
		Exporters:  mergeEntries(base.Exporters, customGoMods["exporters"]),
		Processors: mergeEntries(base.Processors, customGoMods["processors"]),
		Connectors: mergeEntries(base.Connectors, customGoMods["connectors"]),
		Providers:  mergeEntries(base.Providers, customGoMods["providers"]),
	}
}

// WriteToDistributions writes the manifest to ./distributions/<name>/manifest.yaml
// relative to the current working directory.
func WriteToDistributions(name string, m *Manifest) (string, error) {
	dir := filepath.Join("distributions", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating distribution directory: %w", err)
	}

	out, err := yaml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("marshaling manifest: %w", err)
	}

	path := filepath.Join(dir, "manifest.yaml")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return "", fmt.Errorf("writing manifest: %w", err)
	}

	return path, nil
}

// Upgrade returns a new manifest with component versions updated to match
// upstream for known modules. Custom components (not in upstream) are kept as-is.
func Upgrade(local, upstream *Manifest) *Manifest {
	upVersions := upstream.GoModVersionMap()

	upgraded := &Manifest{
		Dist: Dist{
			Module:      local.Dist.Module,
			Name:        local.Dist.Name,
			Description: local.Dist.Description,
			Version:     upstream.Dist.Version,
			OutputPath:  local.Dist.OutputPath,
			BuildTags:   local.Dist.BuildTags,
		},
		Extensions: upgradeEntries(local.Extensions, upVersions),
		Receivers:  upgradeEntries(local.Receivers, upVersions),
		Exporters:  upgradeEntries(local.Exporters, upVersions),
		Processors: upgradeEntries(local.Processors, upVersions),
		Connectors: upgradeEntries(local.Connectors, upVersions),
		Providers:  upgradeEntries(local.Providers, upVersions),
	}

	return upgraded
}

func upgradeEntries(entries []Entry, upVersions map[string]string) []Entry {
	out := make([]Entry, len(entries))
	for i, e := range entries {
		mod, ver := splitGoMod(e.GoMod)
		if upVer, ok := upVersions[mod]; ok {
			ver = upVer
		}
		out[i] = Entry{GoMod: mod + " " + ver}
	}
	return out
}

func mergeEntries(base []Entry, custom []string) []Entry {
	if len(custom) == 0 {
		return base
	}

	// Build set of base module paths for dedup.
	seen := make(map[string]bool)
	for _, e := range base {
		mod, _ := splitGoMod(e.GoMod)
		seen[mod] = true
	}

	merged := make([]Entry, len(base))
	copy(merged, base)

	for _, gomod := range custom {
		mod, _ := splitGoMod(gomod)
		if seen[mod] {
			continue
		}
		merged = append(merged, Entry{GoMod: gomod})
		seen[mod] = true
	}

	return merged
}

func splitGoMod(gomod string) (module, version string) {
	parts := strings.Fields(gomod)
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return gomod, ""
}

