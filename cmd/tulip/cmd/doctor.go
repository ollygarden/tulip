package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/ollygarden/tulip/cmd/tulip/internal/ui"
	"github.com/ollygarden/tulip/cmd/tulip/internal/upstream"
)

var (
	doctorJSON  bool
	doctorCheck bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check distributions against upstream",
	Long:  "Fetch the latest tulip base manifest and validate all local distributions against it.",
	RunE:  runDoctor,
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Output JSON report")
	doctorCmd.Flags().BoolVar(&doctorCheck, "check", false, "Exit code 0 if up to date, 1 if outdated (for CI)")
}

// DoctorOutput is the JSON output schema for tulip doctor --json.
type DoctorOutput struct {
	UpstreamVersion string                 `json:"upstream_version"`
	Outdated        bool                   `json:"outdated"`
	Distributions   []upstream.DriftReport `json:"distributions"`
}

func runDoctor(cmd *cobra.Command, args []string) error {
	if !doctorJSON {
		fmt.Println()
		fmt.Println("  Fetching latest tulip base from ollygarden/tulip...")
	}

	upstreamManifest, err := upstream.FetchUpstreamManifest()
	if err != nil {
		return fmt.Errorf("fetching upstream: %w", err)
	}

	upstreamVersion := upstreamManifest.OTelVersion()
	if !doctorJSON {
		fmt.Printf("  Latest base version: %s\n", upstreamVersion)
	}

	// Find all local distributions
	dists, err := manifest.FindDistributions(distributionsDir())
	if err != nil {
		return fmt.Errorf("finding distributions: %w", err)
	}

	var locals []*manifest.Manifest
	for _, name := range dists {
		m, err := manifest.Parse(filepath.Join(distributionsDir(), name, "manifest.yaml"))
		if err != nil {
			continue
		}
		locals = append(locals, m)
	}

	reports := upstream.CompareAll(locals, upstreamManifest)

	output := DoctorOutput{
		UpstreamVersion: upstreamVersion,
		Distributions:   reports,
	}

	anyOutdated := false
	for _, r := range reports {
		if r.Outdated {
			anyOutdated = true
			break
		}
	}
	output.Outdated = anyOutdated

	if doctorJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	// Human-readable output
	for _, r := range reports {
		suffix := ""
		if r.Name == "tulip" {
			suffix = " (base)"
		}
		fmt.Println(ui.FormatHeader(r.Name + suffix))

		if r.Outdated {
			fmt.Println(ui.FormatWarning(fmt.Sprintf("Outdated: local %s → upstream %s", r.LocalVersion, r.UpstreamVersion)))
			fmt.Printf("    %d components need version bump\n", r.OutdatedComponents)

			if len(r.NewBaseComponents) > 0 {
				fmt.Println(ui.FormatInfo("New base component(s) available:"))
				for _, c := range r.NewBaseComponents {
					fmt.Printf("    + %s\n", c)
				}
			}
		} else {
			fmt.Println(ui.FormatSuccess(fmt.Sprintf("Up to date (%s)", r.LocalVersion)))
		}
	}

	fmt.Println()
	if anyOutdated {
		outdatedCount := 0
		for _, r := range reports {
			if r.Outdated {
				outdatedCount++
			}
		}
		fmt.Printf("  Summary:\n    %d distribution(s) outdated\n", outdatedCount)
		fmt.Println()
		fmt.Println("  Run 'tulip upgrade' to update all distributions automatically.")
		fmt.Println("  Run 'tulip upgrade -d <name>' to update one.")
	} else {
		fmt.Println("  All distributions are up to date.")
	}
	fmt.Println()

	if doctorCheck && anyOutdated {
		os.Exit(1)
	}

	return nil
}
