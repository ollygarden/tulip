package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
)

// resetFlags resets all package-level flag variables to their zero values
// so tests don't bleed state into each other.
func resetFlags() {
	// create flags
	createAddReceivers = nil
	createAddProcessors = nil
	createAddExporters = nil
	createAddExtensions = nil

	// add flags
	addDist = ""
	addReceivers = nil
	addProcessors = nil
	addExporters = nil
	addExtensions = nil

	// remove flags
	removeDist = ""
	removeReceivers = nil
	removeProcessors = nil
	removeExporters = nil
	removeExtensions = nil

	// list flags
	listDist = ""

	// build flags
	buildDist = ""
	buildDocker = false

	// validate flags
	validateDist = ""

	// doctor flags
	doctorJSON = false
	doctorCheck = false

	// upgrade flags
	upgradeDist = ""
}

func setupTestRepoClean(t *testing.T) string {
	t.Helper()
	resetFlags()

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
	t.Cleanup(func() {
		os.Chdir(origDir)
		resetFlags()
	})

	return dir
}
