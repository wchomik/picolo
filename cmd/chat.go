package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wchomik/picolo/docker"
)

var chatCmd = &cobra.Command{
	Use:   "chat [directory]",
	Short: "Start pi agent interactively in a terminal",
	Long: `Start the pi agent container interactively and attach to it.

The agent will mount the specified directory at /app inside the container.
If no directory is specified, the current working directory is used.

The pi config directory defaults to ~/.picolo/pi`,
	Example: `  picolo chat
  picolo chat ~/my-project
  picolo chat /path/to/code`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDocker()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChat(args)
	},
	Args: cobra.MaximumNArgs(1),
}

func runChat(args []string) error {
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

	// Ensure compose file exists
	composePath := client.ComposeFile()
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		fmt.Println("  Compose file not found. Run 'picolo init' first.")
		if err := client.GenerateComposeFile(); err != nil {
			return fmt.Errorf("failed to generate compose file: %w", err)
		}
	}

	// Start llama server if enabled
	if cfg.IsLlamaEnabled() {
		running, _ := client.IsContainerRunning("llama-cpp")
		if !running {
			fmt.Println("  Starting llama-cpp server...")
			if err := client.StartLlama(); err != nil {
				return err
			}
		}
	}

	fmt.Printf("  Starting pi agent in: %s\n", workDir)
	if len(cfg.Extensions) > 0 {
		fmt.Printf("  Extensions: %s\n", cfg.ExtensionsString())
	}

	return client.RunChat(workDir)
}

// expandTilde expands ~ to the user's home directory
func expandTilde(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
