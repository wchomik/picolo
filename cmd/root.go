package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wchomik/picolo/config"
)

var rootCmd = &cobra.Command{
	Use:   "picolo",
	Short: "Quick, safe sandboxed environment for the pi coding agent",
	Long: `Picolo (Pi + Container + Local) provides a quick, easy, and safe
sandboxed environment to play with the pi coding agent.

Everything runs in Docker containers. Local AI is optional but
pretty cool — picolo makes it easy to set up.

Configuration is stored in ~/.picolo/

Commands:
  init      Initialize the environment (pull images, start llama server)
  start     Start containers and resume last session
  status    Show status of all picolo containers
  stop      Stop all running picolo containers
  update    Update all containers to newest versions
  chat      Start pi agent interactively in a terminal
  serve     Start pi agent via ttyd (browser access)

Examples:
  picolo init
  picolo init --env vulkan --skip-llama
  picolo chat ~/my-project
  picolo serve ~/my-project
  picolo start
  picolo stop`,
	Example: `  picolo init
  picolo init --env rocm
  picolo chat ~/my-project
  picolo serve ~/my-project
  picolo start
  picolo stop`,
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
