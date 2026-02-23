package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	name := "acme-collector"
	version := "0.145.0"
	outputPath := "./_build"

	if err := Write(dir, name, version, outputPath); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	t.Run("config.yaml", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
		if err != nil {
			t.Fatalf("reading config.yaml: %v", err)
		}
		content := string(data)

		for _, want := range []string{"receivers:", "processors:", "exporters:", "service:", "0.0.0.0:4317"} {
			if !strings.Contains(content, want) {
				t.Errorf("config.yaml missing %q", want)
			}
		}
	})

	t.Run("Dockerfile", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
		if err != nil {
			t.Fatalf("reading Dockerfile: %v", err)
		}
		content := string(data)

		for _, want := range []string{
			"FROM golang:1.24-alpine AS builder",
			"ARG OCB_VERSION=" + version,
			"ocb --config=manifest.yaml",
			"COPY --from=builder /build/_build/" + name,
			"ENTRYPOINT [\"/acme-collector\"]",
			"EXPOSE 4317 4318",
		} {
			if !strings.Contains(content, want) {
				t.Errorf("Dockerfile missing %q", want)
			}
		}
	})
}
