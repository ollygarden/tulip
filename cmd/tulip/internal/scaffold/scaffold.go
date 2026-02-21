package scaffold

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

//go:embed templates/*
var templateFS embed.FS

// Config holds the parameters used to scaffold a new distribution.
type Config struct {
	Name        string
	Description string
	Module      string
	Org         string
	Version     string
}

// DefaultConfig returns a Config with sensible defaults for the given name.
func DefaultConfig(name string) Config {
	return Config{
		Name:        name,
		Description: fmt.Sprintf("Custom OpenTelemetry Collector distribution: %s", name),
		Module:      fmt.Sprintf("github.com/ollygarden/tulip/%s", name),
		Org:         "OllyGarden",
		Version:     "0.144.0",
	}
}

// templateFile maps a template name to its output filename (relative to the dist dir).
type templateFile struct {
	Template string
	Output   string
	Perm     os.FileMode
}

func templateFiles() []templateFile {
	return []templateFile{
		{"templates/Dockerfile.tmpl", "Dockerfile", 0644},
		{"templates/entrypoint.sh.tmpl", "entrypoint.sh", 0755},
		{"templates/config.yaml.tmpl", "config.yaml", 0644},
		{"templates/gitignore.tmpl", ".gitignore", 0644},
	}
}

// Generate creates a new distribution directory with all scaffold files.
// It also writes the manifest.yaml based on the provided manifest.
func Generate(distDir string, cfg Config, m *manifest.Manifest) error {
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("creating distribution directory: %w", err)
	}

	// Write manifest.yaml
	manifestPath := filepath.Join(distDir, "manifest.yaml")
	if err := manifest.Write(manifestPath, m); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	// Generate each template file
	for _, tf := range templateFiles() {
		if err := renderTemplate(distDir, tf, cfg); err != nil {
			return fmt.Errorf("rendering %s: %w", tf.Output, err)
		}
	}

	return nil
}

func renderTemplate(distDir string, tf templateFile, cfg Config) error {
	tmplData, err := templateFS.ReadFile(tf.Template)
	if err != nil {
		return fmt.Errorf("reading template %s: %w", tf.Template, err)
	}

	tmpl, err := template.New(tf.Output).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("parsing template %s: %w", tf.Template, err)
	}

	outPath := filepath.Join(distDir, tf.Output)
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, tf.Perm)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", outPath, err)
	}
	defer f.Close()

	return tmpl.Execute(f, cfg)
}

// Exists checks if a distribution directory already exists.
func Exists(distDir string) bool {
	_, err := os.Stat(filepath.Join(distDir, "manifest.yaml"))
	return err == nil
}
