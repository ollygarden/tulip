package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/ollygarden/tulip/cmd/tulip/internal/ui"
)

var validateDist string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a distribution",
	Long:  "Validate manifest syntax, version compatibility, and component integrity.",
	RunE:  runValidate,
}

func init() {
	validateCmd.Flags().StringVarP(&validateDist, "distribution", "d", "", "Distribution to validate (default: all)")
}

func runValidate(cmd *cobra.Command, args []string) error {
	if validateDist != "" {
		return validateOne(validateDist)
	}
	return validateAll()
}

func validateOne(name string) error {
	manifestPath := filepath.Join(distDir(name), "manifest.yaml")

	fmt.Println()
	fmt.Printf("  Validating %s...\n", name)
	fmt.Println()

	// Check 1: Parse manifest
	m, err := manifest.Parse(manifestPath)
	if err != nil {
		fmt.Println(ui.FormatError("manifest.yaml syntax invalid: " + err.Error()))
		return fmt.Errorf("validation failed")
	}
	fmt.Println(ui.FormatSuccess("manifest.yaml syntax valid"))

	// Check 2: Version compatibility — all components should use compatible versions
	versions := make(map[string]int)
	for _, c := range m.AllComponents() {
		v := c.Version()
		if v != "" {
			versions[v]++
		}
	}

	// We expect at most 2 versions: one for contrib, one for core
	if len(versions) > 2 {
		fmt.Println(ui.FormatError(fmt.Sprintf("Mismatched component versions: found %d different versions", len(versions))))
		for v, count := range versions {
			fmt.Printf("    %s: %d components\n", v, count)
		}
		return fmt.Errorf("validation failed")
	}
	fmt.Println(ui.FormatSuccess(fmt.Sprintf("Component versions compatible (%s)", m.OTelVersion())))

	// Check 3: No duplicate components
	seen := make(map[string]bool)
	for _, c := range m.AllComponents() {
		name := c.Name()
		if seen[name] {
			fmt.Println(ui.FormatError(fmt.Sprintf("Duplicate component: %s", name)))
			return fmt.Errorf("validation failed")
		}
		seen[name] = true
	}
	fmt.Println(ui.FormatSuccess("No duplicate components"))

	// Check 4: Dist metadata
	if m.Dist.Name == "" {
		fmt.Println(ui.FormatError("dist.name is empty"))
		return fmt.Errorf("validation failed")
	}
	if m.Dist.Module == "" {
		fmt.Println(ui.FormatError("dist.module is empty"))
		return fmt.Errorf("validation failed")
	}
	fmt.Println(ui.FormatSuccess("Distribution metadata valid"))

	fmt.Println()
	fmt.Printf("  %s is valid\n", name)
	fmt.Println()

	return nil
}

func validateAll() error {
	dists, err := manifest.FindDistributions(distributionsDir())
	if err != nil {
		return fmt.Errorf("finding distributions: %w", err)
	}

	var failed []string
	for _, name := range dists {
		if err := validateOne(name); err != nil {
			failed = append(failed, name)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("%d distribution(s) failed validation", len(failed))
	}
	return nil
}
