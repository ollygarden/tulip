package cmd

import (
	"fmt"
	"strings"

	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/ollygarden/tulip/cmd/tulip/internal/catalog"
	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/ollygarden/tulip/cmd/tulip/internal/scaffold"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new OpenTelemetry Collector distribution",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	fmt.Printf("Creating new Tulip distribution: %s\n\n", name)

	// Load the base tulip manifest.
	base, err := manifest.LoadBase()
	if err != nil {
		return err
	}
	baseSet := base.GoModSet()

	// Fetch all available upstream components.
	fmt.Println("Fetching available components...")
	cat, err := catalog.Fetch()
	if err != nil {
		return fmt.Errorf("fetching catalog: %w", err)
	}
	fmt.Printf("Found %d components\n\n", countComponents(cat))

	// Let the user pick extra components (on top of base).
	customGoMods := make(map[string][]string)

	sections := []struct {
		key        string // manifest section key
		title      string
		components []catalog.Component
	}{
		{"extensions", "Extensions", cat.Extensions},
		{"receivers", "Receivers", cat.Receivers},
		{"processors", "Processors", cat.Processors},
		{"exporters", "Exporters", cat.Exporters},
		{"connectors", "Connectors", cat.Connectors},
		{"providers", "Providers", cat.Providers},
	}

	for _, sec := range sections {
		if len(sec.components) == 0 {
			continue
		}

		// Only show components that aren't already in the base.
		var extra []catalog.Component
		for _, c := range sec.components {
			if !baseSet[c.GoMod] {
				extra = append(extra, c)
			}
		}

		if len(extra) == 0 {
			continue
		}

		selected, err := selectComponents(sec.title, extra)
		if err != nil {
			return err
		}
		if len(selected) > 0 {
			customGoMods[sec.key] = selected
		}
	}

	// Generate and write the manifest.
	m := manifest.Generate(name, base, customGoMods)
	path, err := manifest.WriteToDistributions(name, m)
	if err != nil {
		return err
	}

	// Generate config.yaml and Dockerfile.
	distDir := filepath.Dir(path)
	if err := scaffold.Write(distDir, name, m.Dist.Version, m.Dist.OutputPath); err != nil {
		return fmt.Errorf("generating scaffold files: %w", err)
	}

	printSummary(name, base, customGoMods, distDir)
	return nil
}

func selectComponents(title string, components []catalog.Component) ([]string, error) {
	options := make([]huh.Option[string], 0, len(components))
	for _, c := range components {
		label := fmt.Sprintf("%s (%s)", c.Name, c.Version)
		gomod := fmt.Sprintf("%s %s", c.GoMod, c.Version)
		options = append(options, huh.NewOption(label, gomod))
	}

	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(title).
				Options(options...).
				Value(&selected).
				Height(15),
		),
	)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("selecting %s: %w", title, err)
	}

	return selected, nil
}

func countComponents(cat *catalog.Catalog) int {
	return len(cat.Extensions) + len(cat.Receivers) + len(cat.Processors) +
		len(cat.Exporters) + len(cat.Connectors) + len(cat.Providers)
}

func countCustom(customGoMods map[string][]string) int {
	total := 0
	for _, v := range customGoMods {
		total += len(v)
	}
	return total
}

func printSummary(name string, base *manifest.Manifest, customGoMods map[string][]string, distDir string) {
	baseCount := len(base.Extensions) + len(base.Receivers) + len(base.Exporters) +
		len(base.Processors) + len(base.Connectors) + len(base.Providers)
	customCount := countCustom(customGoMods)

	fmt.Printf("\nCreated distribution: %s\n", name)
	fmt.Printf("  Base components:   %d (OllyGarden supported)\n", baseCount)
	fmt.Printf("  Custom components: %d\n", customCount)
	fmt.Printf("  Total:             %d\n", baseCount+customCount)

	if customCount > 0 {
		fmt.Println()
		for _, section := range []string{"extensions", "receivers", "processors", "exporters", "connectors", "providers"} {
			selected := customGoMods[section]
			if len(selected) == 0 {
				continue
			}
			fmt.Printf("  Custom %s:\n", section)
			for _, gomod := range selected {
				parts := strings.Fields(gomod)
				if len(parts) >= 1 {
					fmt.Printf("    + %s\n", parts[0])
				}
			}
		}
	}

	fmt.Printf("\n  Files:\n")
	fmt.Printf("    %s\n", filepath.Join(distDir, "manifest.yaml"))
	fmt.Printf("    %s\n", filepath.Join(distDir, "config.yaml"))
	fmt.Printf("    %s\n", filepath.Join(distDir, "Dockerfile"))
}
