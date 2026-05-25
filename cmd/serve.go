package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wchomik/picolo/config"
	"github.com/wchomik/picolo/docker"
)

var serveCmd = &cobra.Command{
	Use:   "serve [directory]",
	Short: "Start pi agent via ttyd (browser access)",
	Long: `Start the pi agent container in ttyd mode and print the access address.

The agent will be accessible through a web browser at the ttyd port.
The specified directory is mounted at /app inside the container.
If no directory is specified, the current working directory is used.

The pi config directory defaults to ~/.picolo/pi`,
	Example: `  picolo serve
  picolo serve ~/my-project
  picolo serve /path/to/code`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDocker()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServe(args)
	},
	Args: cobra.MaximumNArgs(1),
}

func runServe(args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	client := docker.New(cfg)

	// Resolve work directory
	workDir := "."
	if len(args) > 0 {
		workDir = args[0]
	}

	// Expand ~ to home directory
	workDir, err = expandTilde(workDir)
	if err != nil {
		return fmt.Errorf("failed to resolve directory: %w", err)
	}

	// Ensure directory exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", workDir)
	}

	// Ensure docker network exists
	if err := client.EnsureNetwork(); err != nil {
		return fmt.Errorf("failed to ensure docker network: %w", err)
	}

	// Start llama server if enabled
	if cfg.IsLlamaEnabled() {
		running, _ := client.IsContainerRunning(docker.LlamaContainer)
		if !running {
			fmt.Println("  Starting llama-cpp server...")
			if err := client.StartLlama(); err != nil {
				return err
			}
		}
	}

	fmt.Printf("  Working directory: %s\n", workDir)
	if len(cfg.Extensions) > 0 {
		fmt.Printf("  Extensions: %s\n", cfg.ExtensionsString())
	}

	config.SaveLastMode("serve")
	return client.RunServe(workDir)
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
