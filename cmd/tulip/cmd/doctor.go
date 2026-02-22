package cmd

import (
	"fmt"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check if local distributions are up to date with the OllyGarden base",
	RunE:  runDoctor,
}

// ComponentDrift describes a version mismatch for a single component.
type ComponentDrift struct {
	Module          string
	LocalVersion    string
	UpstreamVersion string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("Fetching latest tulip base from ollygarden/tulip...")
	upstream, err := manifest.FetchUpstream()
	if err != nil {
		return err
	}
	fmt.Printf("Latest base version: %s\n\n", upstream.Dist.Version)

	upstreamVersions := upstream.GoModVersionMap()

	locals, err := manifest.FindLocalDistributions()
	if err != nil {
		return err
	}

	if len(locals) == 0 {
		fmt.Println("No distributions found.")
		return nil
	}

	outdatedCount := 0

	for name, local := range locals {
		localVersions := local.GoModVersionMap()

		var drifted []ComponentDrift
		var missing []string

		// Check each upstream base component against the local distribution.
		for mod, upstreamVer := range upstreamVersions {
			localVer, exists := localVersions[mod]
			if !exists {
				missing = append(missing, mod)
				continue
			}
			if localVer != upstreamVer {
				drifted = append(drifted, ComponentDrift{
					Module:          mod,
					LocalVersion:    localVer,
					UpstreamVersion: upstreamVer,
				})
			}
		}

		isOutdated := len(drifted) > 0 || len(missing) > 0

		printDistributionReport(name, local, upstream, drifted, missing)

		if isOutdated {
			outdatedCount++
		}
	}

	fmt.Println()
	if outdatedCount == 0 {
		fmt.Println("All distributions are up to date.")
	} else {
		fmt.Printf("%d distribution(s) have version drift.\n", outdatedCount)
		fmt.Println("\nRun 'tulip upgrade' to update all distributions.")
	}

	return nil
}

func printDistributionReport(name string, local, upstream *manifest.Manifest, drifted []ComponentDrift, missing []string) {
	label := name
	if name == "tulip" {
		label = "tulip (base)"
	}

	fmt.Printf("── %s ──\n", label)

	if len(drifted) == 0 && len(missing) == 0 {
		fmt.Printf("  Up to date (%s)\n\n", upstream.Dist.Version)
		return
	}

	if local.Dist.Version != upstream.Dist.Version {
		fmt.Printf("  Outdated: local %s -> upstream %s\n", local.Dist.Version, upstream.Dist.Version)
	}

	if len(drifted) > 0 {
		fmt.Printf("  %d component(s) with version drift:\n", len(drifted))
		for _, d := range drifted {
			fmt.Printf("    %s: %s -> %s\n", d.Module, d.LocalVersion, d.UpstreamVersion)
		}
	}

	if len(missing) > 0 {
		fmt.Printf("  %d new base component(s) available:\n", len(missing))
		for _, mod := range missing {
			fmt.Printf("    + %s\n", mod)
		}
	}

	fmt.Println()
}
