package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wchomik/picolo/config"
	"github.com/wchomik/picolo/docker"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start containers and resume last session",
	Long: `Start all picolo containers and resume the last session mode.

If the last session was 'chat', starts the agent interactively.
If the last session was 'serve', starts the agent in browser mode.
If no previous session, defaults to chat mode.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDocker()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStart()
	},
}

func runStart() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	client := docker.New(cfg)

	// Ensure network exists
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

	// Determine mode: last mode or default to chat
	mode := cfg.LastMode
	if mode == "" {
		mode = "chat"
	}

	// Resolve work directory (default to current dir)
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Expand ~ if present
	home, _ := os.UserHomeDir()
	if len(workDir) > 0 && workDir[0] == '~' {
		workDir = filepath.Join(home, workDir[1:])
	}

	switch mode {
	case "chat":
		config.SaveLastMode("chat")
		fmt.Printf("  Resuming chat session in: %s\n", workDir)
		return client.RunChat(workDir)
	case "serve":
		config.SaveLastMode("serve")
		fmt.Printf("  Resuming serve session in: %s\n", workDir)
		return client.RunServe(workDir)
	default:
		return fmt.Errorf("unknown last mode: %s", mode)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
}
