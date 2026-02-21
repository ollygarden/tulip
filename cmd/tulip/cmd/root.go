package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "tulip",
	Short:   "Tulip - Multi-distribution OpenTelemetry Collector platform",
	Long:    "Tulip helps you create, manage, and maintain custom OpenTelemetry Collector distributions.",
	Version: version,
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(catalogCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// distributionsDir returns the path to the distributions directory.
// It looks for it relative to the current working directory.
func distributionsDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "distributions"
	}
	return filepath.Join(cwd, "distributions")
}

// distDir returns the path to a specific distribution's directory.
func distDir(name string) string {
	return filepath.Join(distributionsDir(), name)
}

// baseManifestPath returns the path to the tulip base manifest.
func baseManifestPath() string {
	return filepath.Join(distributionsDir(), "tulip", "manifest.yaml")
}
