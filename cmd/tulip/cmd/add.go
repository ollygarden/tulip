package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/ollygarden/tulip/cmd/tulip/internal/ui"
)

var (
	addDist       string
	addReceivers  []string
	addProcessors []string
	addExporters  []string
	addExtensions []string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add components to a distribution",
	Long:  "Add OpenTelemetry components to an existing distribution's manifest.",
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&addDist, "distribution", "d", "", "Distribution name (required)")
	_ = addCmd.MarkFlagRequired("distribution")
	addCmd.Flags().StringSliceVar(&addReceivers, "receiver", nil, "Receivers to add (comma-separated)")
	addCmd.Flags().StringSliceVar(&addProcessors, "processor", nil, "Processors to add (comma-separated)")
	addCmd.Flags().StringSliceVar(&addExporters, "exporter", nil, "Exporters to add (comma-separated)")
	addCmd.Flags().StringSliceVar(&addExtensions, "extension", nil, "Extensions to add (comma-separated)")
}

func runAdd(cmd *cobra.Command, args []string) error {
	manifestPath := filepath.Join(distDir(addDist), "manifest.yaml")

	m, err := manifest.Parse(manifestPath)
	if err != nil {
		return fmt.Errorf("distribution %q not found: %w", addDist, err)
	}

	otelVersion := m.OTelVersion()
	added := 0

	for _, r := range addReceivers {
		r = strings.TrimSpace(r)
		gomod := fmt.Sprintf("github.com/open-telemetry/opentelemetry-collector-contrib/receiver/%sreceiver %s", r, otelVersion)
		if err := m.AddComponent(manifest.Component{GoMod: gomod}); err != nil {
			return err
		}
		added++
	}
	for _, p := range addProcessors {
		p = strings.TrimSpace(p)
		gomod := fmt.Sprintf("github.com/open-telemetry/opentelemetry-collector-contrib/processor/%sprocessor %s", p, otelVersion)
		if err := m.AddComponent(manifest.Component{GoMod: gomod}); err != nil {
			return err
		}
		added++
	}
	for _, e := range addExporters {
		e = strings.TrimSpace(e)
		gomod := fmt.Sprintf("github.com/open-telemetry/opentelemetry-collector-contrib/exporter/%sexporter %s", e, otelVersion)
		if err := m.AddComponent(manifest.Component{GoMod: gomod}); err != nil {
			return err
		}
		added++
	}
	for _, ext := range addExtensions {
		ext = strings.TrimSpace(ext)
		gomod := fmt.Sprintf("github.com/open-telemetry/opentelemetry-collector-contrib/extension/%sextension %s", ext, otelVersion)
		if err := m.AddComponent(manifest.Component{GoMod: gomod}); err != nil {
			return err
		}
		added++
	}

	if added == 0 {
		return fmt.Errorf("no components specified; use --receiver, --processor, --exporter, or --extension")
	}

	if err := manifest.Write(manifestPath, m); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.FormatSuccess(fmt.Sprintf("Added %d component(s) to %s", added, addDist)))
	fmt.Printf("    Updated: distributions/%s/manifest.yaml\n", addDist)
	fmt.Println()

	return nil
}
