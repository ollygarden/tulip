package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ollygarden/tulip/cmd/tulip/internal/catalog"
	"github.com/ollygarden/tulip/cmd/tulip/internal/ui"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "List all available OTel components",
	Long:  "List all available OpenTelemetry Collector components from the contrib repository.",
	Args:  cobra.NoArgs,
	RunE:  runCatalog,
}

func runCatalog(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println("  Loading component catalog...")

	cat, err := catalog.LoadOrFetch()
	if err != nil {
		return fmt.Errorf("loading catalog: %w", err)
	}

	if len(cat.Entries) == 0 {
		fmt.Println("  No components found.")
		fmt.Println()
		return nil
	}

	fmt.Printf("\n  All components (%d)\n\n", len(cat.Entries))

	fmt.Printf("  %-28s %-12s %s\n",
		ui.HeaderStyle.Render("Component"),
		ui.HeaderStyle.Render("Type"),
		ui.HeaderStyle.Render("Stability"))
	fmt.Println("  " + ui.DimStyle.Render("──────────────────────────────────────────────────────"))

	for _, e := range cat.Entries {
		stability := e.Stability
		if stability == "" {
			stability = "-"
		}
		fmt.Printf("  %-28s %-12s %s\n",
			e.Name,
			e.Type,
			ui.FormatStability(stability))
	}
	fmt.Println()

	return nil
}
