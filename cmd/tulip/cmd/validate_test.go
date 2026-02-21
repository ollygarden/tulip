package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

func TestValidateValid(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create", "valid-dist"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"validate", "-d", "valid-dist"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("validate command failed: %v", err)
	}
}

func TestValidateInvalidManifest(t *testing.T) {
	dir := t.TempDir()
	resetFlags()

	tulipDir := filepath.Join(dir, "distributions", "tulip")
	os.MkdirAll(tulipDir, 0755)
	manifest.Write(filepath.Join(tulipDir, "manifest.yaml"), &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.145.0"},
	})

	badDir := filepath.Join(dir, "distributions", "bad-dist")
	os.MkdirAll(badDir, 0755)
	os.WriteFile(filepath.Join(badDir, "manifest.yaml"), []byte("not: valid: yaml: ["), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer func() {
		os.Chdir(origDir)
		resetFlags()
	}()

	rootCmd.SetArgs([]string{"validate", "-d", "bad-dist"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid manifest")
	}
}

func TestValidateMismatchedVersions(t *testing.T) {
	dir := t.TempDir()
	resetFlags()

	tulipDir := filepath.Join(dir, "distributions", "tulip")
	os.MkdirAll(tulipDir, 0755)
	manifest.Write(filepath.Join(tulipDir, "manifest.yaml"), &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.145.0"},
	})

	badDir := filepath.Join(dir, "distributions", "mismatched")
	os.MkdirAll(badDir, 0755)
	m := &manifest.Manifest{
		Dist: manifest.Dist{
			Name:    "mismatched",
			Module:  "github.com/test/test",
			Version: "0.145.0",
		},
		Receivers: []manifest.Component{
			{GoMod: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.145.0"},
			{GoMod: "go.opentelemetry.io/collector/receiver/nopreceiver v0.143.0"},
			{GoMod: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.142.0"},
		},
	}
	manifest.Write(filepath.Join(badDir, "manifest.yaml"), m)

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer func() {
		os.Chdir(origDir)
		resetFlags()
	}()

	rootCmd.SetArgs([]string{"validate", "-d", "mismatched"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for mismatched versions")
	}
}
