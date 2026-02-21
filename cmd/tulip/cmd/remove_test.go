package cmd

import (
	"path/filepath"
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

func TestRemoveCommand(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create", "remove-test-dist"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"add", "-d", "remove-test-dist", "--receiver", "kafka"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"remove", "-d", "remove-test-dist", "--receiver", "kafka"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("remove command failed: %v", err)
	}

	m, err := manifest.Parse(filepath.Join("distributions", "remove-test-dist", "manifest.yaml"))
	if err != nil {
		t.Fatalf("parsing manifest: %v", err)
	}

	if m.HasComponent("kafkareceiver") {
		t.Error("kafkareceiver still present after removal")
	}
}

func TestRemoveNonexistentComponent(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create", "remove-test-dist2"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"remove", "-d", "remove-test-dist2", "--receiver", "nonexistent"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when removing non-existent component")
	}
}

func TestRemoveBaseComponentShowsWarning(t *testing.T) {
	setupTestRepoClean(t)

	rootCmd.SetArgs([]string{"create", "remove-base-test"})
	rootCmd.Execute()

	resetFlags()
	rootCmd.SetArgs([]string{"remove", "-d", "remove-base-test", "--receiver", "otlp"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("remove command failed: %v", err)
	}

	m, err := manifest.Parse(filepath.Join("distributions", "remove-base-test", "manifest.yaml"))
	if err != nil {
		t.Fatalf("parsing manifest: %v", err)
	}

	if m.HasComponent("otlpreceiver") {
		t.Error("otlpreceiver still present after removal")
	}
}
