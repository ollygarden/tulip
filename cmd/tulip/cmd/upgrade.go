package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/ollygarden/tulip/cmd/tulip/internal/ui"
	"github.com/ollygarden/tulip/cmd/tulip/internal/upstream"
)

var upgradeDist string

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade distributions to latest upstream versions",
	Long:  "Fetch the latest upstream base and update all component versions.",
	RunE:  runUpgrade,
}

func init() {
	upgradeCmd.Flags().StringVarP(&upgradeDist, "distribution", "d", "", "Upgrade a specific distribution (default: all)")
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println("  Fetching latest tulip base from ollygarden/tulip...")

	upstreamManifest, err := upstream.FetchUpstreamManifest()
	if err != nil {
		return fmt.Errorf("fetching upstream: %w", err)
	}

	upstreamVersion := upstreamManifest.OTelVersion()
	fmt.Printf("  Latest base version: %s\n", upstreamVersion)
	fmt.Println()

	var dists []string
	if upgradeDist != "" {
		dists = []string{upgradeDist}
	} else {
		var err error
		dists, err = manifest.FindDistributions(distributionsDir())
		if err != nil {
			return fmt.Errorf("finding distributions: %w", err)
		}
	}

	for _, name := range dists {
		manifestPath := filepath.Join(distributionsDir(), name, "manifest.yaml")
		m, err := manifest.Parse(manifestPath)
		if err != nil {
			fmt.Println(ui.FormatError(fmt.Sprintf("Skipping %s: %s", name, err)))
			continue
		}

		localVersion := m.OTelVersion()
		if localVersion == upstreamVersion {
			fmt.Println(ui.FormatSuccess(fmt.Sprintf("%s is already up to date (%s)", name, localVersion)))
			continue
		}

		upstream.Upgrade(m, upstreamManifest)

		if err := manifest.Write(manifestPath, m); err != nil {
			return fmt.Errorf("writing %s manifest: %w", name, err)
		}

		fmt.Println(ui.FormatSuccess(fmt.Sprintf("Upgraded %s: %s → %s", name, localVersion, upstreamVersion)))
	}

	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Println("    make generate-sources    Regenerate build files")
	fmt.Println("    go mod tidy              Update dependencies")
	fmt.Println("    make test                Verify tests pass")
	fmt.Println()

	return nil
}
