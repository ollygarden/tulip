package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

func TestListDistributions(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create", "list-dist-1"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"create", "list-dist-2"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

func TestListComponents(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create", "list-comp-dist"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"list", "-d", "list-comp-dist"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("list -d command failed: %v", err)
	}
}

func TestListNonexistentDist(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"list", "-d", "nonexistent"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent distribution")
	}
}

func TestListEmptyDistributions(t *testing.T) {
	dir := t.TempDir()
	resetFlags()

	os.MkdirAll(filepath.Join(dir, "distributions"), 0755)

	tulipDir := filepath.Join(dir, "distributions", "tulip")
	os.MkdirAll(tulipDir, 0755)
	manifest.Write(filepath.Join(tulipDir, "manifest.yaml"), &manifest.Manifest{
		Dist: manifest.Dist{Name: "tulip", Version: "0.145.0"},
	})

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer func() {
		os.Chdir(origDir)
		resetFlags()
	}()

	rootCmd.SetArgs([]string{"list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}
