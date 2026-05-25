package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wchomik/picolo/docker"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all picolo containers",
	Long:  `Stop and remove all running picolo containers.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDocker()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStop()
	},
}

func runStop() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	client := docker.New(cfg)

	fmt.Println("🛑 Stopping picolo containers...")

	if err := client.StopAll(); err != nil {
		return fmt.Errorf("failed to stop containers: %w", err)
	}

	fmt.Println("✅ All containers stopped")
	return nil
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
