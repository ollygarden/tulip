package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tulip",
	Short: "Tulip - OpenTelemetry Collector distribution manager",
	Long:  "Tulip helps you create and manage custom OpenTelemetry Collector distributions.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
