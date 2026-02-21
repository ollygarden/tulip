package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseValidManifest(t *testing.T) {
	yaml := `
dist:
  module: github.com/ollygarden/tulip/test
  name: test
  description: Test distribution
  version: 0.145.0
  output_path: ./_build
extensions:
  - gomod: go.opentelemetry.io/collector/extension/zpagesextension v0.145.0
receivers:
  - gomod: go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0
  - gomod: go.opentelemetry.io/collector/receiver/nopreceiver v0.145.0
exporters:
  - gomod: go.opentelemetry.io/collector/exporter/debugexporter v0.145.0
processors:
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/processor/batchprocessor v0.145.0
connectors:
  - gomod: go.opentelemetry.io/collector/connector/forwardconnector v0.145.0
providers:
  - gomod: go.opentelemetry.io/collector/confmap/provider/envprovider v1.51.0
`
	m, err := ParseBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if m.Dist.Name != "test" {
		t.Errorf("Dist.Name = %q, want %q", m.Dist.Name, "test")
	}
	if m.Dist.Version != "0.145.0" {
		t.Errorf("Dist.Version = %q, want %q", m.Dist.Version, "0.145.0")
	}
	if len(m.Extensions) != 1 {
		t.Errorf("len(Extensions) = %d, want 1", len(m.Extensions))
	}
	if len(m.Receivers) != 2 {
		t.Errorf("len(Receivers) = %d, want 2", len(m.Receivers))
	}
	if len(m.Exporters) != 1 {
		t.Errorf("len(Exporters) = %d, want 1", len(m.Exporters))
	}
	if len(m.Processors) != 1 {
		t.Errorf("len(Processors) = %d, want 1", len(m.Processors))
	}
	if len(m.Connectors) != 1 {
		t.Errorf("len(Connectors) = %d, want 1", len(m.Connectors))
	}
	if len(m.Providers) != 1 {
		t.Errorf("len(Providers) = %d, want 1", len(m.Providers))
	}
}

func TestParseInvalidYAML(t *testing.T) {
	_, err := ParseBytes([]byte("not: valid: yaml: ["))
	if err == nil {
		t.Error("ParseBytes() expected error for invalid YAML")
	}
}

func TestWriteRoundTrip(t *testing.T) {
	original := &Manifest{
		Dist: Dist{
			Module:      "github.com/test/test",
			Name:        "test",
			Description: "Test",
			Version:     "0.145.0",
			OutputPath:  "./_build",
		},
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
		Exporters: []Component{
			{GoMod: "go.opentelemetry.io/collector/exporter/debugexporter v0.145.0"},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.yaml")

	if err := Write(path, original); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	parsed, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if parsed.Dist.Name != original.Dist.Name {
		t.Errorf("Dist.Name = %q, want %q", parsed.Dist.Name, original.Dist.Name)
	}
	if len(parsed.Receivers) != len(original.Receivers) {
		t.Errorf("len(Receivers) = %d, want %d", len(parsed.Receivers), len(original.Receivers))
	}
	if len(parsed.Exporters) != len(original.Exporters) {
		t.Errorf("len(Exporters) = %d, want %d", len(parsed.Exporters), len(original.Exporters))
	}
}

func TestComponentName(t *testing.T) {
	tests := []struct {
		gomod string
		want  string
	}{
		{"go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0", "otlpreceiver"},
		{"github.com/open-telemetry/opentelemetry-collector-contrib/processor/batchprocessor v0.145.0", "batchprocessor"},
		{"go.opentelemetry.io/collector/confmap/provider/envprovider v1.51.0", "envprovider"},
		{"", ""},
	}

	for _, tt := range tests {
		c := Component{GoMod: tt.gomod}
		if got := c.Name(); got != tt.want {
			t.Errorf("Component{%q}.Name() = %q, want %q", tt.gomod, got, tt.want)
		}
	}
}

func TestComponentVersion(t *testing.T) {
	tests := []struct {
		gomod string
		want  string
	}{
		{"go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0", "v0.145.0"},
		{"go.opentelemetry.io/collector/confmap/provider/envprovider v1.51.0", "v1.51.0"},
		{"noversion", ""},
	}

	for _, tt := range tests {
		c := Component{GoMod: tt.gomod}
		if got := c.Version(); got != tt.want {
			t.Errorf("Component{%q}.Version() = %q, want %q", tt.gomod, got, tt.want)
		}
	}
}

func TestComponentIsCore(t *testing.T) {
	tests := []struct {
		gomod string
		want  bool
	}{
		{"go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0", true},
		{"github.com/open-telemetry/opentelemetry-collector-contrib/processor/batchprocessor v0.145.0", false},
	}

	for _, tt := range tests {
		c := Component{GoMod: tt.gomod}
		if got := c.IsCore(); got != tt.want {
			t.Errorf("Component{%q}.IsCore() = %v, want %v", tt.gomod, got, tt.want)
		}
	}
}

func TestComponentCount(t *testing.T) {
	m := &Manifest{
		Extensions: []Component{{GoMod: "a v1"}, {GoMod: "b v1"}},
		Receivers:  []Component{{GoMod: "c v1"}},
		Exporters:  []Component{{GoMod: "d v1"}},
	}

	if got := m.ComponentCount(); got != 4 {
		t.Errorf("ComponentCount() = %d, want 4", got)
	}
}

func TestHasComponent(t *testing.T) {
	m := &Manifest{
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}

	if !m.HasComponent("otlpreceiver") {
		t.Error("HasComponent(otlpreceiver) = false, want true")
	}
	if m.HasComponent("kafkareceiver") {
		t.Error("HasComponent(kafkareceiver) = true, want false")
	}
}

func TestAddComponent(t *testing.T) {
	m := &Manifest{}
	c := Component{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0"}

	if err := m.AddComponent(c); err != nil {
		t.Fatalf("AddComponent() error = %v", err)
	}
	if len(m.Receivers) != 1 {
		t.Errorf("len(Receivers) = %d, want 1", len(m.Receivers))
	}

	// Adding duplicate should error
	if err := m.AddComponent(c); err == nil {
		t.Error("AddComponent() expected error for duplicate")
	}
}

func TestRemoveComponent(t *testing.T) {
	m := &Manifest{
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0"},
		},
	}

	if err := m.RemoveComponent("kafkareceiver"); err != nil {
		t.Fatalf("RemoveComponent() error = %v", err)
	}
	if len(m.Receivers) != 1 {
		t.Errorf("len(Receivers) = %d, want 1", len(m.Receivers))
	}

	// Removing non-existent should error
	if err := m.RemoveComponent("nonexistent"); err == nil {
		t.Error("RemoveComponent() expected error for non-existent component")
	}
}

func TestCategorizeComponents(t *testing.T) {
	base := &Manifest{
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}
	dist := &Manifest{
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.145.0"},
		},
	}

	baseComps, customComps := CategorizeComponents(dist, base)
	if len(baseComps) != 1 {
		t.Errorf("len(baseComps) = %d, want 1", len(baseComps))
	}
	if len(customComps) != 1 {
		t.Errorf("len(customComps) = %d, want 1", len(customComps))
	}
}

func TestFindDistributions(t *testing.T) {
	dir := t.TempDir()

	// Create dist1 with manifest
	dist1 := filepath.Join(dir, "dist1")
	os.MkdirAll(dist1, 0755)
	os.WriteFile(filepath.Join(dist1, "manifest.yaml"), []byte("dist:\n  name: dist1"), 0644)

	// Create dist2 with manifest
	dist2 := filepath.Join(dir, "dist2")
	os.MkdirAll(dist2, 0755)
	os.WriteFile(filepath.Join(dist2, "manifest.yaml"), []byte("dist:\n  name: dist2"), 0644)

	// Create dir without manifest (should be skipped)
	os.MkdirAll(filepath.Join(dir, "notadist"), 0755)

	dists, err := FindDistributions(dir)
	if err != nil {
		t.Fatalf("FindDistributions() error = %v", err)
	}
	if len(dists) != 2 {
		t.Errorf("len(dists) = %d, want 2", len(dists))
	}
}

func TestOTelVersion(t *testing.T) {
	m := &Manifest{
		Dist: Dist{Version: "0.144.0"},
		Receivers: []Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
	}

	if got := m.OTelVersion(); got != "v0.145.0" {
		t.Errorf("OTelVersion() = %q, want %q", got, "v0.145.0")
	}

	// Empty manifest should fall back to dist version
	empty := &Manifest{Dist: Dist{Version: "0.144.0"}}
	if got := empty.OTelVersion(); got != "0.144.0" {
		t.Errorf("OTelVersion() = %q, want %q", got, "0.144.0")
	}
}
