package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/ollygarden/tulip/cmd/tulip/internal/manifest"
	"github.com/ollygarden/tulip/cmd/tulip/internal/ui"
)

var (
	buildDist   string
	buildDocker bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build distributions",
	Long:  "Build one or all distributions. Wraps the Makefile build with progress feedback.",
	RunE:  runBuild,
}

func init() {
	buildCmd.Flags().StringVarP(&buildDist, "distribution", "d", "", "Build a specific distribution (default: all)")
	buildCmd.Flags().BoolVar(&buildDocker, "docker", false, "Build Docker image instead of binary")
}

func runBuild(cmd *cobra.Command, args []string) error {
	if buildDist != "" {
		return buildOne(buildDist)
	}
	return buildAll()
}

func buildOne(name string) error {
	fmt.Println()
	if buildDocker {
		fmt.Printf("  %s Building %s (docker)...\n", ui.Bullet, name)
	} else {
		fmt.Printf("  %s Building %s...\n", ui.Bullet, name)
	}

	start := time.Now()

	target := "build"
	if buildDocker {
		target = "docker"
	}

	buildCmd := exec.Command("make", target, fmt.Sprintf("DISTRIBUTIONS=%s", name))
	buildCmd.Dir = findRepoRoot()
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		fmt.Printf("    └── %s\n", ui.ErrorStyle.Render("Build failed"))
		return fmt.Errorf("build failed: %w", err)
	}

	elapsed := time.Since(start).Round(time.Second)
	fmt.Printf("    └── Done in %s\n", elapsed)
	fmt.Println()

	if !buildDocker {
		fmt.Printf("  Binary: distributions/%s/_build/%s\n", name, name)
	} else {
		fmt.Printf("  Image: %s:local\n", name)
	}
	fmt.Println()

	return nil
}

func buildAll() error {
	dists, err := manifest.FindDistributions(distributionsDir())
	if err != nil {
		return fmt.Errorf("finding distributions: %w", err)
	}

	fmt.Println()
	fmt.Printf("  %s Building all distributions...\n", ui.Bullet)

	for _, name := range dists {
		start := time.Now()

		target := "build"
		if buildDocker {
			target = "docker"
		}

		cmd := exec.Command("make", target, fmt.Sprintf("DISTRIBUTIONS=%s", name))
		cmd.Dir = findRepoRoot()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("    ├── %-30s %s\n", name, ui.ErrorStyle.Render("FAILED"))
			return fmt.Errorf("building %s: %w", name, err)
		}

		elapsed := time.Since(start).Round(time.Second)
		fmt.Printf("    ├── %-30s %s  (%s)\n", name, ui.CheckMark, elapsed)
	}

	fmt.Println()
	return nil
}

func findRepoRoot() string {
	// Walk up to find directory with Makefile
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(fmt.Sprintf("%s/Makefile", dir)); err == nil {
			return dir
		}
		parent := fmt.Sprintf("%s/..", dir)
		abs, err := filepath.Abs(parent)
		if err != nil || abs == dir {
			break
		}
		dir = abs
	}
	// Fallback
	cwd, _ := os.Getwd()
	return cwd
}
