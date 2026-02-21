package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/ollygarden/tulip/cmd/tulip/internal/scaffold"
	"github.com/ollygarden/tulip/cmd/tulip/internal/ui"
)

var (
	createAddReceivers  []string
	createAddProcessors []string
	createAddExporters  []string
	createAddExtensions []string
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new distribution",
	Long:  "Create a new OpenTelemetry Collector distribution with the tulip base components.",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func init() {
	createCmd.Flags().StringSliceVar(&createAddReceivers, "add-receiver", nil, "Additional receivers to include (comma-separated)")
	createCmd.Flags().StringSliceVar(&createAddProcessors, "add-processor", nil, "Additional processors to include (comma-separated)")
	createCmd.Flags().StringSliceVar(&createAddExporters, "add-exporter", nil, "Additional exporters to include (comma-separated)")
	createCmd.Flags().StringSliceVar(&createAddExtensions, "add-extension", nil, "Additional extensions to include (comma-separated)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	dir := distDir(name)

	if scaffold.Exists(dir) {
		return fmt.Errorf("distribution %q already exists at %s", name, dir)
	}

	// Load the base tulip manifest (falls back to embedded copy if not on disk)
	base, err := manifest.LoadBaseManifest(baseManifestPath())
	if err != nil {
		return fmt.Errorf("loading base manifest: %w", err)
	}

	// Prepare scaffold config
	cfg := scaffold.DefaultConfig(name)
	cfg.Version = base.Dist.Version

	// Build additions from flags
	var additions []manifest.Component
	otelVersion := base.OTelVersion()

	for _, r := range createAddReceivers {
		r = strings.TrimSpace(r)
		gomod := fmt.Sprintf("github.com/open-telemetry/opentelemetry-collector-contrib/receiver/%sreceiver %s", r, otelVersion)
		additions = append(additions, manifest.Component{GoMod: gomod})
	}
	for _, p := range createAddProcessors {
		p = strings.TrimSpace(p)
		gomod := fmt.Sprintf("github.com/open-telemetry/opentelemetry-collector-contrib/processor/%sprocessor %s", p, otelVersion)
		additions = append(additions, manifest.Component{GoMod: gomod})
	}
	for _, e := range createAddExporters {
		e = strings.TrimSpace(e)
		gomod := fmt.Sprintf("github.com/open-telemetry/opentelemetry-collector-contrib/exporter/%sexporter %s", e, otelVersion)
		additions = append(additions, manifest.Component{GoMod: gomod})
	}
	for _, ext := range createAddExtensions {
		ext = strings.TrimSpace(ext)
		gomod := fmt.Sprintf("github.com/open-telemetry/opentelemetry-collector-contrib/extension/%sextension %s", ext, otelVersion)
		additions = append(additions, manifest.Component{GoMod: gomod})
	}

	// Merge base + additions
	distMeta := manifest.Dist{
		Module:      cfg.Module,
		Name:        cfg.Name,
		Description: cfg.Description,
		Version:     cfg.Version,
		OutputPath:  "./_build",
		BuildTags:   "grpcnotrace",
	}

	result, err := manifest.Merge(base, distMeta, additions, nil)
	if err != nil {
		return fmt.Errorf("merging manifest: %w", err)
	}

	// Generate scaffold
	if err := scaffold.Generate(dir, cfg, result.Manifest); err != nil {
		return fmt.Errorf("generating distribution: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.FormatSuccess(fmt.Sprintf("Created distributions/%s/", name)))

	customCount := len(result.Added)
	if customCount > 0 {
		fmt.Printf("    manifest.yaml  (base + %d custom components)\n", customCount)
	} else {
		fmt.Println("    manifest.yaml  (base components)")
	}

	fmt.Println("    config.yaml")
	fmt.Println("    Dockerfile")
	fmt.Println("    ...")
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Printf("    tulip build -d %s    Build locally\n", name)
	fmt.Println("    git add . && git push              CI builds & pushes image")
	fmt.Println()

	return nil
}
