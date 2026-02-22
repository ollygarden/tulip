package cmd

import (
	"fmt"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade all local distributions to the latest OllyGarden base versions",
	RunE:  runUpgrade,
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	fmt.Println("Fetching latest tulip base from ollygarden/tulip...")
	upstream, err := manifest.FetchUpstream()
	if err != nil {
		return err
	}
	fmt.Printf("Latest base version: %s\n\n", upstream.Dist.Version)

	locals, err := manifest.FindLocalDistributions()
	if err != nil {
		return err
	}

	if len(locals) == 0 {
		fmt.Println("No distributions found.")
		return nil
	}

	upstreamVersions := upstream.GoModVersionMap()
	upgraded := 0

	for name, local := range locals {
		localVersions := local.GoModVersionMap()

		// Check if this distribution needs upgrading.
		needsUpgrade := false
		if local.Dist.Version != upstream.Dist.Version {
			needsUpgrade = true
		}
		if !needsUpgrade {
			for mod, upVer := range upstreamVersions {
				if localVer, ok := localVersions[mod]; ok && localVer != upVer {
					needsUpgrade = true
					break
				}
			}
		}

		if !needsUpgrade {
			fmt.Printf("  %s: already up to date\n", name)
			continue
		}

		result := manifest.Upgrade(local, upstream)
		if _, err := manifest.WriteToDistributions(name, result); err != nil {
			return fmt.Errorf("writing upgraded manifest for %s: %w", name, err)
		}

		fmt.Printf("  %s: %s -> %s\n", name, local.Dist.Version, upstream.Dist.Version)
		upgraded++
	}

	fmt.Println()
	if upgraded == 0 {
		fmt.Println("All distributions are already up to date.")
	} else {
		fmt.Printf("Upgraded %d distribution(s).\n", upgraded)
	}

	return nil
}
