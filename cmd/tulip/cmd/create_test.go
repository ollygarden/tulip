package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

func TestCreateCommand(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create", "test-dist"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("create command failed: %v", err)
	}

	expectedFiles := []string{
		"manifest.yaml",
		"Dockerfile",
		"entrypoint.sh",
		"config.yaml",
	}

	for _, f := range expectedFiles {
		path := filepath.Join("distributions", "test-dist", f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", f)
		}
	}

	m, err := manifest.Parse(filepath.Join("distributions", "test-dist", "manifest.yaml"))
	if err != nil {
		t.Fatalf("parsing created manifest: %v", err)
	}
	if !m.HasComponent("otlpreceiver") {
		t.Error("Created manifest missing base component otlpreceiver")
	}
}

func TestCreateWithNoName(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when no name provided")
	}
}

func TestCreateExistingDist(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create", "existing-dist"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"create", "existing-dist"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when distribution already exists")
	}
}

func TestCreateWithAddReceiver(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create", "custom-dist", "--add-receiver", "kafka"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("create command failed: %v", err)
	}

	m, err := manifest.Parse(filepath.Join("distributions", "custom-dist", "manifest.yaml"))
	if err != nil {
		t.Fatalf("parsing created manifest: %v", err)
	}

	if !m.HasComponent("kafkareceiver") {
		t.Error("Created manifest does not include kafkareceiver")
	}
	if !m.HasComponent("otlpreceiver") {
		t.Error("Created manifest missing base component otlpreceiver")
	}
}
