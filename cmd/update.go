package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wchomik/picolo/docker"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all containers to newest versions",
	Long: `Update all Docker containers to their newest versions.
Pulls the latest images from the registry and restarts running containers.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDocker()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate()
	},
}

func runUpdate() error {
	fmt.Println("🔄 Updating picolo environment...")

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	client := docker.New(cfg)

	// Pull latest images
	fmt.Println("📦 Pulling latest images...")
	if err := client.PullAll(); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}

	// Restart running containers
	fmt.Println("🔄 Restarting containers...")

	if cfg.IsLlamaEnabled() {
		running, _ := client.IsContainerRunning(docker.LlamaContainer)
		if running {
			fmt.Println("  Restarting llama-cpp...")
			if err := client.RestartContainer(docker.LlamaContainer); err != nil {
				return fmt.Errorf("failed to restart llama-cpp: %w", err)
			}
		}
	}

	running, _ := client.IsContainerRunning(docker.AgentContainer)
	if running {
		fmt.Println("  Restarting pi-agent...")
		if err := client.RestartContainer(docker.AgentContainer); err != nil {
			return fmt.Errorf("failed to restart pi-agent: %w", err)
		}
	}

	fmt.Println("✅ Update complete!")
	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
