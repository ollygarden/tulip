package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

func testManifest() *manifest.Manifest {
	return &manifest.Manifest{
		Dist: manifest.Dist{
			Module:      "github.com/ollygarden/tulip/test-dist",
			Name:        "test-dist",
			Description: "Test distribution",
			Version:     "0.145.0",
			OutputPath:  "./_build",
			BuildTags:   "grpcnotrace",
		},
		Extensions: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/extension/zpagesextension v0.145.0"},
		},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
		},
		Exporters: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/exporter/debugexporter v0.145.0"},
		},
	}
}

func TestGenerateCreatesAllFiles(t *testing.T) {
	dir := t.TempDir()
	distDir := filepath.Join(dir, "test-dist")
	cfg := DefaultConfig("test-dist")

	if err := Generate(distDir, cfg, testManifest()); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	expectedFiles := []string{
		"manifest.yaml",
		"Dockerfile",
		"entrypoint.sh",
		"config.yaml",
		".gitignore",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(distDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", f)
		}
	}
}

func TestGenerateTemplateVariables(t *testing.T) {
	dir := t.TempDir()
	distDir := filepath.Join(dir, "my-collector")
	cfg := DefaultConfig("my-collector")
	cfg.Description = "My Custom Collector"
	cfg.Org = "TestOrg"

	if err := Generate(distDir, cfg, testManifest()); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check Dockerfile references the distribution name
	dockerfile, err := os.ReadFile(filepath.Join(distDir, "Dockerfile"))
	if err != nil {
		t.Fatalf("reading Dockerfile: %v", err)
	}
	if !strings.Contains(string(dockerfile), "my-collector") {
		t.Error("Dockerfile does not contain distribution name")
	}
	if !strings.Contains(string(dockerfile), "TestOrg") {
		t.Error("Dockerfile does not contain org name")
	}

	// Check entrypoint.sh references correct paths
	entrypoint, err := os.ReadFile(filepath.Join(distDir, "entrypoint.sh"))
	if err != nil {
		t.Fatalf("reading entrypoint.sh: %v", err)
	}
	if !strings.Contains(string(entrypoint), "/etc/my-collector/") {
		t.Error("entrypoint.sh does not contain correct config path")
	}
	if !strings.Contains(string(entrypoint), "/my-collector") {
		t.Error("entrypoint.sh does not contain correct binary path")
	}
}

func TestGenerateManifestContainsComponents(t *testing.T) {
	dir := t.TempDir()
	distDir := filepath.Join(dir, "test-dist")
	cfg := DefaultConfig("test-dist")

	m := testManifest()
	if err := Generate(distDir, cfg, m); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Parse the generated manifest
	parsed, err := manifest.Parse(filepath.Join(distDir, "manifest.yaml"))
	if err != nil {
		t.Fatalf("parsing generated manifest: %v", err)
	}

	if len(parsed.Extensions) != 1 {
		t.Errorf("len(Extensions) = %d, want 1", len(parsed.Extensions))
	}
	if len(parsed.Receivers) != 1 {
		t.Errorf("len(Receivers) = %d, want 1", len(parsed.Receivers))
	}
	if len(parsed.Exporters) != 1 {
		t.Errorf("len(Exporters) = %d, want 1", len(parsed.Exporters))
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	distDir := filepath.Join(dir, "test-dist")
	cfg := DefaultConfig("test-dist")

	if Exists(distDir) {
		t.Error("Exists() = true before generation")
	}

	Generate(distDir, cfg, testManifest())

	if !Exists(distDir) {
		t.Error("Exists() = false after generation")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("my-k8s-collector")

	if cfg.Name != "my-k8s-collector" {
		t.Errorf("Name = %q, want %q", cfg.Name, "my-k8s-collector")
	}
	if !strings.Contains(cfg.Module, "my-k8s-collector") {
		t.Errorf("Module = %q, should contain distribution name", cfg.Module)
	}
}
