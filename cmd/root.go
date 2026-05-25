package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wchomik/picolo/config"
)

var rootCmd = &cobra.Command{
	Use:   "picolo",
	Short: "Orchestrate local, sandboxed Docker-based pi agent environment",
	Long: `Picolo is a CLI tool to orchestrate a local, sandboxed Docker-based
pi coding agent environment. It manages llama.cpp servers and pi agent
containers with support for various GPU backends.

Configuration is stored in ~/.picolo/

Commands:
  init      Initialize the environment (pull images, start llama server)
  update    Update all containers to newest versions
  chat      Start pi agent interactively in a terminal
  serve     Start pi agent via ttyd (browser access)

Examples:
  picolo init
  picolo init --env vulkan --skip-llama
  picolo chat ~/my-project
  picolo serve ~/my-project
  picolo update`,
	Example: `  picolo init
  picolo init --env rocm
  picolo chat ~/my-project
  picolo serve ~/my-project`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// loadConfig loads the configuration
func loadConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func init() {
	cobra.OnInitialize()
}
