package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

func TestAddCommand(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create", "add-test-dist"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"add", "-d", "add-test-dist", "--receiver", "kafka"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add command failed: %v", err)
	}

	m, err := manifest.Parse(filepath.Join("distributions", "add-test-dist", "manifest.yaml"))
	if err != nil {
		t.Fatalf("parsing manifest: %v", err)
	}

	if !m.HasComponent("kafkareceiver") {
		t.Error("kafkareceiver not found in manifest after add")
	}
}

func TestAddNonexistentDist(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"add", "-d", "nonexistent", "--receiver", "kafka"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent distribution")
	}
}

func TestAddDuplicateComponent(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create", "dup-test-dist"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"add", "-d", "dup-test-dist", "--receiver", "kafka"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"add", "-d", "dup-test-dist", "--receiver", "kafka"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when adding duplicate component")
	}
}

func TestAddNoComponents(t *testing.T) {
	dir := t.TempDir()
	resetFlags()

	tulipDir := filepath.Join(dir, "distributions", "tulip")
	os.MkdirAll(tulipDir, 0755)
	manifest.Write(filepath.Join(tulipDir, "manifest.yaml"), &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.145.0"},
	})

	distDir := filepath.Join(dir, "distributions", "empty-dist")
	os.MkdirAll(distDir, 0755)
	manifest.Write(filepath.Join(distDir, "manifest.yaml"), &manifest.Manifest{
		Dist: manifest.Dist{Name: "empty-dist", Version: "0.145.0"},
	})

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer func() {
		os.Chdir(origDir)
		resetFlags()
	}()

	rootCmd.SetArgs([]string{"add", "-d", "empty-dist"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when no components specified")
	}
}
