package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/ollygarden/tulip/cmd/tulip/internal/ui"
)

var listDist string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List distributions or components",
	Long:  "List all distributions, or list components in a specific distribution.",
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVarP(&listDist, "distribution", "d", "", "Show components for a specific distribution")
}

func runList(cmd *cobra.Command, args []string) error {
	if listDist != "" {
		return listComponents(listDist)
	}
	return listDistributions()
}

func listDistributions() error {
	dists, err := manifest.FindDistributions(distributionsDir())
	if err != nil {
		return fmt.Errorf("finding distributions: %w", err)
	}

	if len(dists) == 0 {
		fmt.Println("  No distributions found.")
		return nil
	}

	// Load base manifest for comparison (falls back to embedded copy)
	base, _ := manifest.LoadBaseManifest(baseManifestPath())
	baseCount := 0
	if base != nil {
		baseCount = base.ComponentCount()
	}

	fmt.Println()
	fmt.Println("  " + ui.TitleStyle.Render("Tulip Distributions"))
	fmt.Println()
	fmt.Printf("  %-22s %-12s %s\n",
		ui.HeaderStyle.Render("Distribution"),
		ui.HeaderStyle.Render("Components"),
		ui.HeaderStyle.Render("Source"))
	fmt.Println("  " + ui.DimStyle.Render("──────────────────────────────────────────────────"))

	for _, name := range dists {
		m, err := manifest.Parse(filepath.Join(distributionsDir(), name, "manifest.yaml"))
		if err != nil {
			continue
		}

		count := m.ComponentCount()
		source := "custom"
		countStr := fmt.Sprintf("%d", count)

		if name == "tulip" {
			source = "base"
		} else if base != nil {
			diff := count - baseCount
			if diff > 0 {
				countStr = fmt.Sprintf("%d (+%d)", count, diff)
			} else if diff < 0 {
				countStr = fmt.Sprintf("%d (%d)", count, diff)
			}
		}

		fmt.Printf("  %-22s %-12s %s\n",
			name,
			countStr,
			ui.FormatSource(source))
	}
	fmt.Println()

	return nil
}

func listComponents(distName string) error {
	m, err := manifest.Parse(filepath.Join(distributionsDir(), distName, "manifest.yaml"))
	if err != nil {
		return fmt.Errorf("distribution %q not found: %w", distName, err)
	}

	// Load base for categorization (falls back to embedded copy)
	base, _ := manifest.LoadBaseManifest(baseManifestPath())

	fmt.Println()
	fmt.Printf("  %s  %s\n",
		ui.TitleStyle.Render(distName),
		ui.DimStyle.Render(m.OTelVersion()))
	fmt.Println()
	fmt.Printf("  %-32s %-12s %s\n",
		ui.HeaderStyle.Render("Component"),
		ui.HeaderStyle.Render("Type"),
		ui.HeaderStyle.Render("Source"))
	fmt.Println("  " + ui.DimStyle.Render("──────────────────────────────────────────────────────"))

	printSection := func(components []manifest.Component, compType string) {
		for _, c := range components {
			source := "custom"
			if base != nil && base.HasComponent(c.Name()) {
				source = "base"
			}
			fmt.Printf("  %-32s %-12s %s\n",
				c.Name(),
				compType,
				ui.FormatSource(source))
		}
	}

	printSection(m.Extensions, "extension")
	printSection(m.Receivers, "receiver")
	printSection(m.Exporters, "exporter")
	printSection(m.Processors, "processor")
	printSection(m.Connectors, "connector")
	printSection(m.Providers, "provider")

	fmt.Println()
	fmt.Println("  " + ui.DimStyle.Render("base = supported by OllyGarden | custom = user-added"))
	fmt.Println()

	return nil
}
