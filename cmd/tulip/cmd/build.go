package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	buildDocker bool
	buildTag    string
)

func init() {
	buildCmd.Flags().BoolVar(&buildDocker, "docker", false, "Build a Docker image after compiling")
	buildCmd.Flags().StringVar(&buildTag, "tag", "", "Docker image tag (default: <name>:local)")
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build <name>",
	Short: "Build a Tulip distribution binary (and optionally a Docker image)",
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

	ocbPath, err := findOCB()
	if err != nil {
		return err
	}

	// Build the native binary.
	fmt.Printf("Building %s...\n", name)
	if err := runOCB(ocbPath, distDir, nil); err != nil {
		return fmt.Errorf("ocb build failed: %w", err)
	}
	fmt.Println("Build succeeded.")

	if !buildDocker {
		fmt.Printf("\nBinary: %s\n", filepath.Join(distDir, "_build", name))
		return nil
	}

	// Docker build: cross-compile a linux binary, copy it next to the
	// Dockerfile, build the image, then clean up.
	tag := buildTag
	if tag == "" {
		tag = name + ":local"
	}

	fmt.Printf("\nCross-compiling for linux/%s...\n", runtime.GOARCH)
	env := []string{
		"CGO_ENABLED=0",
		"GOOS=linux",
		"GOARCH=" + runtime.GOARCH,
	}
	if err := runOCB(ocbPath, distDir, env); err != nil {
		return fmt.Errorf("linux cross-compile failed: %w", err)
	}

	// Copy binary next to Dockerfile so the COPY instruction works.
	src := filepath.Join(distDir, "_build", name)
	dst := filepath.Join(distDir, name)
	if err := copyFile(src, dst); err != nil {
		return fmt.Errorf("copying binary for docker build: %w", err)
	}
	defer os.Remove(dst)

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

// runOCB executes ocb with the given working directory and optional extra env vars.
func runOCB(ocbPath, distDir string, extraEnv []string) error {
	cmd := exec.Command(ocbPath, "--config", "manifest.yaml")
	cmd.Dir = distDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	return cmd.Run()
}

// copyFile copies src to dst using read/write (no hardlink issues across filesystems).
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o755)
}
