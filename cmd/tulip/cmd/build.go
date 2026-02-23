package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	buildDocker bool
	buildTag    string
)

func init() {
	buildCmd.Flags().BoolVar(&buildDocker, "docker", false, "Build a Docker image (no local OCB required)")
	buildCmd.Flags().StringVar(&buildTag, "tag", "", "Docker image tag (default: <name>:local)")
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build <name>",
	Short: "Build a Tulip distribution binary or Docker image",
	Args:  cobra.ExactArgs(1),
	RunE:  runBuild,
}

func runBuild(cmd *cobra.Command, args []string) error {
	name := args[0]

	distDir := filepath.Join("distributions", name)
	manifestPath := filepath.Join(distDir, "manifest.yaml")

	if _, err := os.Stat(manifestPath); err != nil {
		return fmt.Errorf("distribution %q not found (expected %s)", name, manifestPath)
	}

	if buildDocker {
		return runDockerBuild(name, distDir)
	}

	return runLocalBuild(name, distDir)
}

// runLocalBuild compiles a native binary using a locally installed OCB.
func runLocalBuild(name, distDir string) error {
	ocbPath, err := findOCB()
	if err != nil {
		return err
	}

	fmt.Printf("Building %s...\n", name)
	if err := runOCB(ocbPath, distDir); err != nil {
		return fmt.Errorf("ocb build failed: %w", err)
	}
	fmt.Println("Build succeeded.")
	fmt.Printf("\nBinary: %s\n", filepath.Join(distDir, "_build", name))
	return nil
}

// runDockerBuild builds a Docker image using the distribution's multi-stage
// Dockerfile. OCB is downloaded and the binary is compiled inside the Docker
// build — no local tooling beyond Docker is required.
func runDockerBuild(name, distDir string) error {
	dockerfile := filepath.Join(distDir, "Dockerfile")
	if _, err := os.Stat(dockerfile); err != nil {
		return fmt.Errorf("no Dockerfile in distribution %q (expected %s)", name, dockerfile)
	}

	tag := buildTag
	if tag == "" {
		tag = name + ":local"
	}

	fmt.Printf("Building Docker image %s...\n", tag)
	docker := exec.Command("docker", "build", "-t", tag, ".")
	docker.Dir = distDir
	docker.Stdout = os.Stdout
	docker.Stderr = os.Stderr
	if err := docker.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	fmt.Printf("\nDocker image built: %s\n", tag)
	return nil
}

// findOCB locates the ocb binary in PATH or the well-known fallback location.
func findOCB() (string, error) {
	if p, err := exec.LookPath("ocb"); err == nil {
		return p, nil
	}

	home, err := os.UserHomeDir()
	if err == nil {
		fallback := filepath.Join(home, "bin", "ocb")
		if _, err := os.Stat(fallback); err == nil {
			return fallback, nil
		}
	}

	return "", fmt.Errorf("ocb not found in PATH or ~/bin/ocb\nInstall it with: make install-ocb  (or see https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder)")
}

// runOCB executes ocb with the given working directory.
func runOCB(ocbPath, distDir string) error {
	cmd := exec.Command(ocbPath, "--config", "manifest.yaml")
	cmd.Dir = distDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
