package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	ConfigName  = "picolo"
	ConfigType  = "yaml"
	DefaultEnv  = "cuda"
)

// LlamaEnvironment represents the GPU backend for llama.cpp
type LlamaEnvironment string

const (
	LlamaCUDA   LlamaEnvironment = "cuda"
	LlamaVulkan LlamaEnvironment = "vulkan"
	LlamaROCM   LlamaEnvironment = "rocm"
	LlamaMetal  LlamaEnvironment = "metal"
	LlamaCPU    LlamaEnvironment = "cpu"
)

// Config holds all picolo configuration
type Config struct {
	HomeDir      string           `mapstructure:"home_dir"`
	Env          LlamaEnvironment `mapstructure:"env"`
	SkipLlama    bool             `mapstructure:"skip_llama"`
	Extensions   []string         `mapstructure:"extensions"`
	PIAgentImage string           `mapstructure:"pi_agent_image"`
	TtydPort     int              `mapstructure:"ttyd_port"`
	LlamaPort    int              `mapstructure:"llama_port"`
}

// GetConfig returns the config directory path (~/.picolo)
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".picolo"), nil
}

// GetPIConfigDir returns the pi config directory (~/.picolo/pi)
func GetPIConfigDir() (string, error) {
	home, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "pi"), nil
}

// LoadConfig reads config from file and environment
func LoadConfig() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	viper.AddConfigPath(configDir)
	viper.SetConfigName(ConfigName)
	viper.SetConfigType(ConfigType)

	// Set defaults
	viper.SetDefault("env", string(LlamaCUDA))
	viper.SetDefault("skip_llama", false)
	viper.SetDefault("extensions", []string{"npm:pi-observability", "npm:pi-web-access"})
	viper.SetDefault("pi_agent_image", "ghcr.io/wchomik/pi-docker:latest")
	viper.SetDefault("ttyd_port", 7681)
	viper.SetDefault("llama_port", 8080)

	// Try to read config file (ignore if not found)
	_ = viper.ReadInConfig()

	cfg := &Config{
		HomeDir:      configDir,
		Env:          LlamaEnvironment(viper.GetString("env")),
		SkipLlama:    viper.GetBool("skip_llama"),
		Extensions:   viper.GetStringSlice("extensions"),
		PIAgentImage: viper.GetString("pi_agent_image"),
		TtydPort:     viper.GetInt("ttyd_port"),
		LlamaPort:    viper.GetInt("llama_port"),
	}

	return cfg, nil
}

// SaveConfig writes config to file
func SaveConfig(cfg *Config) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.Set("env", string(cfg.Env))
	viper.Set("skip_llama", cfg.SkipLlama)
	viper.Set("extensions", cfg.Extensions)
	viper.Set("pi_agent_image", cfg.PIAgentImage)
	viper.Set("ttyd_port", cfg.TtydPort)
	viper.Set("llama_port", cfg.LlamaPort)

	configPath := filepath.Join(configDir, ConfigName+"."+ConfigType)
	return viper.WriteConfigAs(configPath)
}

// ExtensionsString returns extensions as comma-separated string
func (c *Config) ExtensionsString() string {
	return strings.Join(c.Extensions, ",")
}

// LlamaImage returns the appropriate llama.cpp docker image for the configured environment
func (c *Config) LlamaImage() string {
	return fmt.Sprintf("ghcr.io/ggml-org/llama.cpp:server-%s", string(c.Env))
}

// IsLlamaEnabled returns whether llama.cpp server should be started
func (c *Config) IsLlamaEnabled() bool {
	return !c.SkipLlama
}
