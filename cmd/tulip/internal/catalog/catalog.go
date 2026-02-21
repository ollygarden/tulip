package catalog

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// ContribBuilderConfigURL is the URL to the builder-config.yaml in the contrib repo.
	ContribBuilderConfigURL = "https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/main/cmd/otelcontribcol/builder-config.yaml"
)

// Entry represents a single component in the catalog.
type Entry struct {
	Name        string `json:"name" yaml:"name"`
	Type        string `json:"type" yaml:"type"`
	GoMod       string `json:"gomod" yaml:"gomod"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Stability   string `json:"stability,omitempty" yaml:"stability,omitempty"`
}

// Catalog holds the full set of available OTel components.
type Catalog struct {
	Entries []Entry `json:"entries" yaml:"entries"`
	Version string  `json:"version" yaml:"version"`
}

// builderConfig mirrors the structure of the contrib builder-config.yaml.
type builderConfig struct {
	Dist       struct{ Version string } `yaml:"dist"`
	Extensions []struct{ GoMod string } `yaml:"extensions"`
	Receivers  []struct{ GoMod string } `yaml:"receivers"`
	Exporters  []struct{ GoMod string } `yaml:"exporters"`
	Processors []struct{ GoMod string } `yaml:"processors"`
	Connectors []struct{ GoMod string } `yaml:"connectors"`
}

// FetchCatalog downloads and parses the builder-config.yaml from the contrib repo.
func FetchCatalog() (*Catalog, error) {
	return FetchCatalogFromURL(ContribBuilderConfigURL)
}

// FetchCatalogFromURL downloads and parses a builder-config.yaml from the given URL.
func FetchCatalogFromURL(url string) (*Catalog, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching catalog: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading catalog response: %w", err)
	}

	return ParseBuilderConfig(data)
}

// ParseBuilderConfig parses the builder-config.yaml format into a Catalog.
func ParseBuilderConfig(data []byte) (*Catalog, error) {
	var cfg builderConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing builder config: %w", err)
	}

	catalog := &Catalog{
		Version: cfg.Dist.Version,
	}

	for _, ext := range cfg.Extensions {
		catalog.Entries = append(catalog.Entries, entryFromGoMod(ext.GoMod, "extension"))
	}
	for _, rcv := range cfg.Receivers {
		catalog.Entries = append(catalog.Entries, entryFromGoMod(rcv.GoMod, "receiver"))
	}
	for _, exp := range cfg.Exporters {
		catalog.Entries = append(catalog.Entries, entryFromGoMod(exp.GoMod, "exporter"))
	}
	for _, proc := range cfg.Processors {
		catalog.Entries = append(catalog.Entries, entryFromGoMod(proc.GoMod, "processor"))
	}
	for _, conn := range cfg.Connectors {
		catalog.Entries = append(catalog.Entries, entryFromGoMod(conn.GoMod, "connector"))
	}

	return catalog, nil
}

func entryFromGoMod(gomod, componentType string) Entry {
	parts := strings.Fields(gomod)
	modPath := ""
	if len(parts) > 0 {
		modPath = parts[0]
	}
	name := filepath.Base(modPath)
	return Entry{
		Name:  name,
		Type:  componentType,
		GoMod: gomod,
	}
}

// Search returns entries whose name contains the query (case-insensitive).
func (c *Catalog) Search(query string) []Entry {
	query = strings.ToLower(query)
	var results []Entry
	for _, e := range c.Entries {
		if strings.Contains(strings.ToLower(e.Name), query) {
			results = append(results, e)
		}
	}
	return results
}

// FilterByType returns entries of the given component type.
func (c *Catalog) FilterByType(componentType string) []Entry {
	var results []Entry
	for _, e := range c.Entries {
		if e.Type == componentType {
			results = append(results, e)
		}
	}
	return results
}

// Find returns the entry with the given name, or nil if not found.
func (c *Catalog) Find(name string) *Entry {
	for _, e := range c.Entries {
		if e.Name == name {
			return &e
		}
	}
	return nil
}
