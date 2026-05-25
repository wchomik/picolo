package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wchomik/picolo/config"
	"github.com/wchomik/picolo/docker"
)

var (
	initEnv     string
	initSkipLlama bool
	initExtensions string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the picolo environment",
	Long: `Initialize the picolo environment:
  - Create config directory (~/.picolo)
  - Pull all required Docker images
  - Start the llama.cpp server (unless --skip-llama)
  - Generate docker-compose.yaml

The environment parameter determines which GPU backend to use for llama.cpp.
Supported values: cuda, vulkan, rocm, metal, cpu`,
	Example: `  picolo init
  picolo init --env vulkan
  picolo init --env rocm --skip-llama
  picolo init --extensions npm:pi-observability,npm:pi-web-access,npm:pi-github`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckDocker()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func runInit() error {
	fmt.Println("🏗️  Initializing picolo environment...")

	// Load existing config (may not exist yet)
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// Override with flags if provided
	if initEnv != "" {
		cfg.Env = config.LlamaEnvironment(initEnv)
	}
	if initSkipLlama {
		cfg.SkipLlama = true
	}
	if initExtensions != "" {
		cfg.Extensions = strings.Split(initExtensions, ",")
	}

	// Validate environment
	validEnvs := []string{"cuda", "vulkan", "rocm", "metal", "cpu"}
	if !contains(validEnvs, string(cfg.Env)) {
		return fmt.Errorf("invalid environment: %s (valid: %s)", cfg.Env, strings.Join(validEnvs, ", "))
	}

	// Create config directory
	if err := os.MkdirAll(cfg.HomeDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create pi config directory
	piConfigDir := filepath.Join(cfg.HomeDir, "pi")
	if err := os.MkdirAll(piConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create pi config directory: %w", err)
	}

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("  ✓ Config directory: %s\n", cfg.HomeDir)
	fmt.Printf("  ✓ Pi config directory: %s\n", piConfigDir)
	fmt.Printf("  ✓ Environment: %s\n", cfg.Env)

	// Generate docker-compose file
	client := docker.New(cfg)
	if err := client.GenerateComposeFile(); err != nil {
		return fmt.Errorf("failed to generate docker-compose.yaml: %w", err)
	}
	fmt.Printf("  ✓ Docker compose file: %s\n", client.ComposeFile())

	// Pull images
	fmt.Println("\n📦 Pulling Docker images...")
	if err := client.PullAll(); err != nil {
		return err
	}

	// Start llama server
	if cfg.IsLlamaEnabled() {
		fmt.Println("\n🚀 Starting llama.cpp server...")
		if err := client.StartLlama(); err != nil {
			return err
		}
		fmt.Printf("  Llama.cpp server available at http://localhost:%d\n", cfg.LlamaPort)
	} else {
		fmt.Println("\n  ⏭️  Skipping llama.cpp server (--skip-llama)")
	}

	// Check for models directory
	modelsDir := filepath.Join(os.Getenv("HOME"), ".models")
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		fmt.Println("\n  ⚠️  Models directory not found at ~/.models")
		fmt.Println("     Place your GGUF models there and create a config.ini preset.")
		fmt.Println("     See: https://huggingface.co/blog/ggml-org/model-management-in-llamacpp")
	}

	fmt.Println("\n✅ Environment initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  picolo chat ~/my-project    # Start agent in a directory")
	fmt.Println("  picolo serve ~/my-project   # Serve agent via browser")

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	initCmd.Flags().StringVar(&initEnv, "env", "", "llama.cpp environment (cuda, vulkan, rocm, metal, cpu)")
	initCmd.Flags().BoolVar(&initSkipLlama, "skip-llama", false, "Skip setting up local llama.cpp deployment")
	initCmd.Flags().StringVar(&initExtensions, "extensions", "", "Comma-separated list of pi extensions")

	rootCmd.AddCommand(initCmd)
}
