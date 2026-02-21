//go:build integration

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/cmd"
	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/ollygarden/tulip/cmd/tulip/internal/upstream"
)

func setupIntegrationRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	tulipDir := filepath.Join(dir, "distributions", "tulip")
	os.MkdirAll(tulipDir, 0755)

	baseManifest := &manifest.Manifest{
		Dist: manifest.Dist{
			Module:      "github.com/ollygarden/tulip/tulip",
			Name:        "tulip",
			Description: "Base distribution",
			Version:     "0.145.0",
			OutputPath:  "./_build",
			BuildTags:   "grpcnotrace",
		},
		Extensions: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/extension/zpagesextension v0.145.0"},
		},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			{GoMod: "go.opentelemetry.io/collector/receiver/nopreceiver v0.145.0"},
		},
		Exporters: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/exporter/debugexporter v0.145.0"},
			{GoMod: "go.opentelemetry.io/collector/exporter/otlpexporter v0.145.0"},
		},
		Processors: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/processor/batchprocessor v0.145.0"},
		},
		Connectors: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/connector/forwardconnector v0.145.0"},
		},
		Providers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/confmap/provider/envprovider v1.51.0"},
		},
	}

	manifest.Write(filepath.Join(tulipDir, "manifest.yaml"), baseManifest)

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(origDir) })

	return dir
}

func TestFullLifecycle(t *testing.T) {
	setupIntegrationRepo(t)

	// 1. Create
	os.Args = []string{"tulip", "create", "integration-test-dist"}
	if err := cmd.Execute(); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 2. Verify files exist
	expectedFiles := []string{
		"manifest.yaml", "Dockerfile", "entrypoint.sh", "config.yaml",
		"integration-test-dist.service", "integration-test-dist.conf",
		"preinstall.sh", "postinstall.sh", "preremove.sh",
		"integration-test-dist-test.yaml", ".gitignore",
	}
	for _, f := range expectedFiles {
		path := filepath.Join("distributions", "integration-test-dist", f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", f)
		}
	}

	// 3. Add component
	os.Args = []string{"tulip", "add", "-d", "integration-test-dist", "--receiver", "kafka"}
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// 4. Verify component present
	m, _ := manifest.Parse(filepath.Join("distributions", "integration-test-dist", "manifest.yaml"))
	if !m.HasComponent("kafkareceiver") {
		t.Error("kafkareceiver not found after add")
	}

	// 5. Validate
	os.Args = []string{"tulip", "validate", "-d", "integration-test-dist"}
	if err := cmd.Execute(); err != nil {
		t.Fatalf("validate failed: %v", err)
	}

	// 6. Remove component
	os.Args = []string{"tulip", "remove", "-d", "integration-test-dist", "--receiver", "kafka"}
	if err := cmd.Execute(); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	// 7. Verify removed
	m, _ = manifest.Parse(filepath.Join("distributions", "integration-test-dist", "manifest.yaml"))
	if m.HasComponent("kafkareceiver") {
		t.Error("kafkareceiver still present after remove")
	}
}

func TestDoctorReport(t *testing.T) {
	setupIntegrationRepo(t)

	// Create a distribution with old versions
	os.Args = []string{"tulip", "create", "old-dist"}
	cmd.Execute()

	// Use the compare logic directly (avoid network call)
	oldManifest, _ := manifest.Parse(filepath.Join("distributions", "old-dist", "manifest.yaml"))

	upstreamM := &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.148.0"},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.148.0"},
			{GoMod: "go.opentelemetry.io/collector/receiver/nopreceiver v0.148.0"},
		},
	}

	report := upstream.Compare(oldManifest, upstreamM)
	if !report.Outdated {
		t.Error("Old distribution should be reported as outdated")
	}
}

func TestMultiDistribution(t *testing.T) {
	setupIntegrationRepo(t)

	names := []string{"multi-dist-1", "multi-dist-2", "multi-dist-3"}
	for _, name := range names {
		os.Args = []string{"tulip", "create", name}
		if err := cmd.Execute(); err != nil {
			t.Fatalf("create %s failed: %v", name, err)
		}
	}

	// Verify all exist
	dists, _ := manifest.FindDistributions("distributions")
	// Should have base tulip + 3 created = 4
	if len(dists) != 4 {
		t.Errorf("Found %d distributions, want 4", len(dists))
	}
}
