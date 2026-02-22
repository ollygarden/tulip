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

	if err := Write(dir, name); err != nil {
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
			"FROM alpine:latest",
			"COPY --chmod=755 " + name,
			"/etc/" + name + "/config.yaml",
			"EXPOSE 4317",
		} {
			if !strings.Contains(content, want) {
				t.Errorf("Dockerfile missing %q", want)
			}
		}
	})
}
