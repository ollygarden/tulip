package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestLoadBase loads the real base manifest from distributions/tulip/manifest.yaml
// and asserts that every component section is populated.
func TestLoadBase(t *testing.T) {
	t.Parallel()

	m, err := LoadBase()
	if err != nil {
		t.Fatalf("LoadBase() error = %v", err)
	}

	if m.Dist.Name == "" {
		t.Error("Dist.Name is empty")
	}
	if m.Dist.Module == "" {
		t.Error("Dist.Module is empty")
	}
	if m.Dist.Version == "" {
		t.Error("Dist.Version is empty")
	}

	sections := []struct {
		name    string
		entries []Entry
	}{
		{"extensions", m.Extensions},
		{"receivers", m.Receivers},
		{"exporters", m.Exporters},
		{"processors", m.Processors},
		{"connectors", m.Connectors},
		{"providers", m.Providers},
	}

	for _, s := range sections {
		if len(s.entries) == 0 {
			t.Errorf("section %q has no entries", s.name)
		}
		for i, e := range s.entries {
			if e.GoMod == "" {
				t.Errorf("section %q entry %d has empty GoMod", s.name, i)
			}
		}
	}
}

// TestGenerate verifies that Generate correctly merges base components with
// custom components and deduplicates entries whose module path already exists
// in the base.
func TestGenerate(t *testing.T) {
	t.Parallel()

	base := &Manifest{
		Dist: Dist{
			Module:     "github.com/ollygarden/tulip/tulip",
			Name:       "tulip",
			Version:    "0.144.0",
			OutputPath: "./_build",
			BuildTags:  "grpcnotrace",
		},
		Extensions: []Entry{
			{GoMod: "go.opentelemetry.io/collector/extension/zpagesextension v0.145.0"},
		},
		Receivers: []Entry{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
		Exporters: []Entry{
			{GoMod: "go.opentelemetry.io/collector/exporter/otlpexporter v0.145.0"},
		},
		Processors: []Entry{
			{GoMod: "go.opentelemetry.io/collector/processor/batchprocessor v0.145.0"},
		},
		Connectors: []Entry{
			{GoMod: "go.opentelemetry.io/collector/connector/forwardconnector v0.145.0"},
		},
		Providers: []Entry{
			{GoMod: "go.opentelemetry.io/collector/confmap/provider/envprovider v1.51.0"},
		},
	}

	const distName = "acme-custom"

	custom := map[string][]string{
		// New extension not in base.
		"extensions": {"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.145.0"},
		// Duplicate of base receiver — must be deduplicated.
		"receivers": {"go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		// New exporter not in base.
		"exporters": {"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.145.0"},
		// No custom processors — base should pass through unchanged.
		// No custom connectors.
		// No custom providers.
	}

	got := Generate(distName, base, custom)

	t.Run("dist_metadata", func(t *testing.T) {
		if got.Dist.Name != distName {
			t.Errorf("Dist.Name = %q, want %q", got.Dist.Name, distName)
		}
		wantModule := "github.com/ollygarden/tulip/" + distName
		if got.Dist.Module != wantModule {
			t.Errorf("Dist.Module = %q, want %q", got.Dist.Module, wantModule)
		}
		if got.Dist.Version != base.Dist.Version {
			t.Errorf("Dist.Version = %q, want %q (inherited from base)", got.Dist.Version, base.Dist.Version)
		}
		if got.Dist.BuildTags != base.Dist.BuildTags {
			t.Errorf("Dist.BuildTags = %q, want %q (inherited from base)", got.Dist.BuildTags, base.Dist.BuildTags)
		}
	})

	t.Run("extensions_merged", func(t *testing.T) {
		// Base has 1 extension, custom adds 1 new → expect 2.
		if len(got.Extensions) != 2 {
			t.Errorf("len(Extensions) = %d, want 2", len(got.Extensions))
		}
		assertContainsGoMod(t, got.Extensions,
			"go.opentelemetry.io/collector/extension/zpagesextension v0.145.0",
			"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.145.0",
		)
	})

	t.Run("receivers_deduplicated", func(t *testing.T) {
		// Custom has only the same receiver as base → no addition, length stays 1.
		if len(got.Receivers) != 1 {
			t.Errorf("len(Receivers) = %d, want 1 (duplicate should be dropped)", len(got.Receivers))
		}
		assertContainsGoMod(t, got.Receivers,
			"go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0",
		)
	})

	t.Run("exporters_merged", func(t *testing.T) {
		// Base has 1, custom adds 1 new → expect 2.
		if len(got.Exporters) != 2 {
			t.Errorf("len(Exporters) = %d, want 2", len(got.Exporters))
		}
		assertContainsGoMod(t, got.Exporters,
			"go.opentelemetry.io/collector/exporter/otlpexporter v0.145.0",
			"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.145.0",
		)
	})

	t.Run("processors_passthrough", func(t *testing.T) {
		// No custom processors — base passes through unchanged.
		if len(got.Processors) != len(base.Processors) {
			t.Errorf("len(Processors) = %d, want %d", len(got.Processors), len(base.Processors))
		}
	})

	t.Run("connectors_passthrough", func(t *testing.T) {
		if len(got.Connectors) != len(base.Connectors) {
			t.Errorf("len(Connectors) = %d, want %d", len(got.Connectors), len(base.Connectors))
		}
	})

	t.Run("providers_passthrough", func(t *testing.T) {
		if len(got.Providers) != len(base.Providers) {
			t.Errorf("len(Providers) = %d, want %d", len(got.Providers), len(base.Providers))
		}
	})
}

// TestWriteToDistributions generates a manifest, writes it to a temp distribution
// directory, reads it back and verifies the round-trip is lossless.
//
// WriteToDistributions writes into ./distributions/ relative to the cwd, so
// the test uses a temp directory as the working directory.
func TestWriteToDistributions(t *testing.T) {
	// Not parallel: changes the process working directory.

	const distName = "test-manifest-roundtrip"

	// Use a temp directory as cwd so we don't pollute the real repo.
	tmp := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting cwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir to temp: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	base := &Manifest{
		Dist: Dist{
			Module:     "github.com/ollygarden/tulip/tulip",
			Name:       "tulip",
			Version:    "0.99.0",
			OutputPath: "./_build",
		},
		Extensions: []Entry{
			{GoMod: "go.opentelemetry.io/collector/extension/zpagesextension v0.99.0"},
		},
		Receivers: []Entry{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.99.0"},
		},
	}

	m := Generate(distName, base, nil)

	path, err := WriteToDistributions(distName, m)
	if err != nil {
		t.Fatalf("WriteToDistributions() error = %v", err)
	}

	if path == "" {
		t.Fatal("WriteToDistributions() returned empty path")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written manifest: %v", err)
	}

	var got Manifest
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshaling written manifest: %v", err)
	}

	if got.Dist.Name != m.Dist.Name {
		t.Errorf("round-trip Dist.Name = %q, want %q", got.Dist.Name, m.Dist.Name)
	}
	if got.Dist.Module != m.Dist.Module {
		t.Errorf("round-trip Dist.Module = %q, want %q", got.Dist.Module, m.Dist.Module)
	}
	if got.Dist.Version != m.Dist.Version {
		t.Errorf("round-trip Dist.Version = %q, want %q", got.Dist.Version, m.Dist.Version)
	}
	if len(got.Extensions) != len(m.Extensions) {
		t.Errorf("round-trip len(Extensions) = %d, want %d", len(got.Extensions), len(m.Extensions))
	}
	if len(got.Receivers) != len(m.Receivers) {
		t.Errorf("round-trip len(Receivers) = %d, want %d", len(got.Receivers), len(m.Receivers))
	}

	// Verify the written file is inside the distributions/ directory.
	if filepath.Base(path) != "manifest.yaml" {
		t.Errorf("written file name = %q, want manifest.yaml", filepath.Base(path))
	}
	if filepath.Base(filepath.Dir(path)) != distName {
		t.Errorf("parent dir name = %q, want %q", filepath.Base(filepath.Dir(path)), distName)
	}
}

// TestUpgrade verifies that Upgrade updates known component versions to
// upstream while preserving custom (non-upstream) components as-is.
func TestUpgrade(t *testing.T) {
	t.Parallel()

	local := &Manifest{
		Dist: Dist{
			Module:     "github.com/ollygarden/tulip/my-dist",
			Name:       "my-dist",
			Version:    "0.140.0",
			OutputPath: "./_build",
		},
		Receivers: []Entry{
			// Known upstream component at old version.
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.140.0"},
		},
		Exporters: []Entry{
			// Known upstream component at old version.
			{GoMod: "go.opentelemetry.io/collector/exporter/otlpexporter v0.140.0"},
			// Custom component not in upstream — should stay as-is.
			{GoMod: "github.com/acme/custom-exporter v1.2.3"},
		},
		Processors: []Entry{
			{GoMod: "go.opentelemetry.io/collector/processor/batchprocessor v0.140.0"},
		},
	}

	upstream := &Manifest{
		Dist: Dist{
			Module:  "github.com/ollygarden/tulip/tulip",
			Name:    "tulip",
			Version: "0.145.0",
		},
		Receivers: []Entry{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
		Exporters: []Entry{
			{GoMod: "go.opentelemetry.io/collector/exporter/otlpexporter v0.145.0"},
		},
		Processors: []Entry{
			{GoMod: "go.opentelemetry.io/collector/processor/batchprocessor v0.145.0"},
		},
	}

	got := Upgrade(local, upstream)

	t.Run("dist_version_updated", func(t *testing.T) {
		if got.Dist.Version != "0.145.0" {
			t.Errorf("Dist.Version = %q, want %q", got.Dist.Version, "0.145.0")
		}
	})

	t.Run("dist_name_preserved", func(t *testing.T) {
		if got.Dist.Name != "my-dist" {
			t.Errorf("Dist.Name = %q, want %q", got.Dist.Name, "my-dist")
		}
	})

	t.Run("known_receiver_upgraded", func(t *testing.T) {
		assertContainsGoMod(t, got.Receivers,
			"go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0",
		)
	})

	t.Run("known_exporter_upgraded", func(t *testing.T) {
		assertContainsGoMod(t, got.Exporters,
			"go.opentelemetry.io/collector/exporter/otlpexporter v0.145.0",
		)
	})

	t.Run("custom_exporter_preserved", func(t *testing.T) {
		assertContainsGoMod(t, got.Exporters,
			"github.com/acme/custom-exporter v1.2.3",
		)
	})

	t.Run("exporters_count", func(t *testing.T) {
		if len(got.Exporters) != 2 {
			t.Errorf("len(Exporters) = %d, want 2", len(got.Exporters))
		}
	})

	t.Run("known_processor_upgraded", func(t *testing.T) {
		assertContainsGoMod(t, got.Processors,
			"go.opentelemetry.io/collector/processor/batchprocessor v0.145.0",
		)
	})
}

// assertContainsGoMod is a test helper that fails the test if any of the
// expected GoMod strings are absent from entries.
func assertContainsGoMod(t *testing.T, entries []Entry, want ...string) {
	t.Helper()

	present := make(map[string]bool, len(entries))
	for _, e := range entries {
		present[e.GoMod] = true
	}

	for _, w := range want {
		if !present[w] {
			t.Errorf("entries missing GoMod %q", w)
		}
	}
}
