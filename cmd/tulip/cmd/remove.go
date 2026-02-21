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
	removeDist       string
	removeReceivers  []string
	removeProcessors []string
	removeExporters  []string
	removeExtensions []string
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove components from a distribution",
	Long:  "Remove OpenTelemetry components from a distribution's manifest.",
	RunE:  runRemove,
}

func init() {
	removeCmd.Flags().StringVarP(&removeDist, "distribution", "d", "", "Distribution name (required)")
	_ = removeCmd.MarkFlagRequired("distribution")
	removeCmd.Flags().StringSliceVar(&removeReceivers, "receiver", nil, "Receivers to remove (comma-separated)")
	removeCmd.Flags().StringSliceVar(&removeProcessors, "processor", nil, "Processors to remove (comma-separated)")
	removeCmd.Flags().StringSliceVar(&removeExporters, "exporter", nil, "Exporters to remove (comma-separated)")
	removeCmd.Flags().StringSliceVar(&removeExtensions, "extension", nil, "Extensions to remove (comma-separated)")
}

func runRemove(cmd *cobra.Command, args []string) error {
	manifestPath := filepath.Join(distDir(removeDist), "manifest.yaml")

	m, err := manifest.Parse(manifestPath)
	if err != nil {
		return fmt.Errorf("distribution %q not found: %w", removeDist, err)
	}

	// Optionally load base to warn about removing base components
	base, baseErr := manifest.LoadBaseManifest(baseManifestPath())

	removed := 0
	allNames := collectRemovalNames()

	for _, name := range allNames {
		name = strings.TrimSpace(name)
		// Warn if this is a base component
		if baseErr == nil && base.HasComponent(name) {
			fmt.Println(ui.FormatWarning(fmt.Sprintf("%s is a base (OllyGarden-supported) component", name)))
		}
		if err := m.RemoveComponent(name); err != nil {
			return fmt.Errorf("removing %s: %w", name, err)
		}
		removed++
	}

	if removed == 0 {
		return fmt.Errorf("no components specified; use --receiver, --processor, --exporter, or --extension")
	}

	if err := manifest.Write(manifestPath, m); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.FormatSuccess(fmt.Sprintf("Removed %d component(s) from %s", removed, removeDist)))
	fmt.Printf("    Updated: distributions/%s/manifest.yaml\n", removeDist)
	fmt.Println()

	return nil
}

func collectRemovalNames() []string {
	var names []string
	for _, r := range removeReceivers {
		r = strings.TrimSpace(r)
		if !strings.HasSuffix(r, "receiver") {
			r += "receiver"
		}
		names = append(names, r)
	}
	for _, p := range removeProcessors {
		p = strings.TrimSpace(p)
		if !strings.HasSuffix(p, "processor") {
			p += "processor"
		}
		names = append(names, p)
	}
	for _, e := range removeExporters {
		e = strings.TrimSpace(e)
		if !strings.HasSuffix(e, "exporter") {
			e += "exporter"
		}
		names = append(names, e)
	}
	for _, ext := range removeExtensions {
		ext = strings.TrimSpace(ext)
		if !strings.HasSuffix(ext, "extension") {
			ext += "extension"
		}
		names = append(names, ext)
	}
	return names
}
