package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wchomik/picolo/docker"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of picolo containers",
	Long:  `Show the current status of all picolo containers and configuration.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDocker()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

func runStatus() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	client := docker.New(cfg)

	fmt.Println("📊 Picolo status")
	fmt.Println(strings.Repeat("─", 40))

	// Show config
	fmt.Printf("  Config: %s/picolo.yaml\n", cfg.HomeDir)
	fmt.Printf("  Environment: %s\n", cfg.Env)
	fmt.Printf("  Llama server: %s\n", func() string {
		if cfg.SkipLlama {
			return "disabled"
		}
		return "enabled"
	}())
	if cfg.LastMode != "" {
		fmt.Printf("  Last mode: %s\n", cfg.LastMode)
	}
	fmt.Println()

	// Show container status
	llamaRunning, _ := client.IsContainerRunning(docker.LlamaContainer)
	agentRunning, _ := client.IsContainerRunning(docker.AgentContainer)

	if cfg.IsLlamaEnabled() {
		fmt.Printf("  %s  %s  %s\n",
			containerIcon(llamaRunning),
			docker.LlamaContainer,
			statusLabel(llamaRunning, cfg.LlamaPort, 8080),
		)
	}

	fmt.Printf("  %s  %s  %s\n",
		containerIcon(agentRunning),
		docker.AgentContainer,
		statusLabel(agentRunning, cfg.TtydPort, 7681),
	)

	return nil
}

func containerIcon(running bool) string {
	if running {
		return "🟢"
	}
	return "⚫"
}

func statusLabel(running bool, hostPort, containerPort int) string {
	if running {
		return fmt.Sprintf("running (:%d → :%d)", hostPort, containerPort)
	}
	return "stopped"
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
