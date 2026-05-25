package docker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wchomik/picolo/config"
)

const (
	NetworkName   = "picolo-net"
	LlamaContainer = "picolo-llama-cpp"
	AgentContainer = "picolo-pi-agent"
)

// Client wraps docker CLI operations
type Client struct {
	config *config.Config
}

// New creates a new docker client
func New(cfg *config.Config) *Client {
	return &Client{config: cfg}
}

// runDocker executes a docker command and returns combined output
func (c *Client) runDocker(args ...string) (string, error) {
	cmd := exec.Command("docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		output := stderr.String()
		if output == "" {
			output = stdout.String()
		}
		return "", fmt.Errorf("docker %s failed: %w\nOutput: %s", strings.Join(args, " "), err, output)
	}
	return stdout.String(), nil
}

// runDockerInteractive executes a docker command with stdin/stdout/stderr passthrough
func (c *Client) runDockerInteractive(args ...string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// EnsureNetwork creates the picolo docker network if it doesn't exist
func (c *Client) EnsureNetwork() error {
	_, err := c.runDocker("network", "inspect", NetworkName)
	if err == nil {
		return nil // network already exists
	}
	// Network doesn't exist, create it
	fmt.Printf("  Creating docker network: %s\n", NetworkName)
	_, err = c.runDocker("network", "create", NetworkName)
	if err != nil {
		return fmt.Errorf("failed to create network %s: %w", NetworkName, err)
	}
	return nil
}

// PullImage pulls a docker image
func (c *Client) PullImage(image string) error {
	fmt.Printf("  Pulling image: %s\n", image)
	_, err := c.runDocker("pull", image)
	if err != nil {
		return fmt.Errorf("failed to pull %s: %w", image, err)
	}
	return nil
}

// PullAll pulls all required images
func (c *Client) PullAll() error {
	if c.config.IsLlamaEnabled() {
		if err := c.PullImage(c.config.LlamaImage()); err != nil {
			return err
		}
	}
	return c.PullImage(c.config.PIAgentImage)
}

// llamaRunArgs builds the docker run arguments for the llama-cpp server
func (c *Client) llamaRunArgs() []string {
	args := []string{
		"run", "-d",
		"--name", LlamaContainer,
		"--restart", "unless-stopped",
		"--network", NetworkName,
		"-p", fmt.Sprintf("%d:8080", c.config.LlamaPort),
		"-v", filepath.Join(os.Getenv("HOME"), ".models") + ":/models",
	}

	// GPU configuration
	switch c.config.Env {
	case config.LlamaCUDA:
		args = append(args, "--gpus", "all")
	case config.LlamaROCM:
		args = append(args,
			"--device", "/dev/kfd",
			"--device", "/dev/dri",
			"--group-add", "video",
			"-e", "HSA_OVERRIDE_GFX_VERSION=11.0.0",
		)
	case config.LlamaVulkan:
		// Vulkan uses host GPU directly, no special device flags needed
	case config.LlamaMetal:
		// Metal is macOS-only
	case config.LlamaCPU:
		// No GPU flags
	}

	// Image
	args = append(args, c.config.LlamaImage())

	// Command
	args = append(args,
		"--models-preset", "/models/config.ini",
		"--models-max", "1",
		"--host", "0.0.0.0",
	)

	return args
}

// piAgentRunArgs builds the base docker run arguments for the pi-agent container
func (c *Client) piAgentRunArgs(workDir string) []string {
	piConfigDir := filepath.Join(c.config.HomeDir, "pi")

	args := []string{
		"--name", AgentContainer,
		"--network", NetworkName,
		"-v", piConfigDir + ":/root/.pi",
		"-v", workDir + ":/app",
	}

	// LLM endpoint
	if c.config.IsLlamaEnabled() {
		args = append(args, "-e", "PI_LLM_ENDPOINT=http://picolo-llama-cpp:8080/v1")
	} else {
		args = append(args, "-e", "PI_LLM_ENDPOINT=http://host.docker.internal:8080/v1")
	}

	// Extensions
	args = append(args, "-e", "PI_EXTENSIONS="+c.config.ExtensionsString())

	return args
}

// StartLlama starts the llama-cpp server
func (c *Client) StartLlama() error {
	if !c.config.IsLlamaEnabled() {
		fmt.Println("  Llama.cpp server is disabled (skip_llama=true)")
		return nil
	}

	// Check if already running
	running, _ := c.IsContainerRunning(LlamaContainer)
	if running {
		fmt.Println("  llama-cpp server is already running")
		return nil
	}

	fmt.Println("  Starting llama-cpp server...")
	if _, err := c.runDocker(c.llamaRunArgs()...); err != nil {
		return fmt.Errorf("failed to start llama-cpp: %w", err)
	}
	fmt.Printf("  ✓ llama-cpp server started (port %d)\n", c.config.LlamaPort)
	return nil
}

// RunChat starts pi-agent interactively (runs `pi` directly, bypassing ttyd entrypoint)
func (c *Client) RunChat(workDir string) error {
	args := []string{"run", "-it", "--rm"}
	args = append(args, c.piAgentRunArgs(workDir)...)
	args = append(args, c.config.PIAgentImage)

	// Override entrypoint: run `pi` directly instead of ttyd
	// Extensions are already installed via PI_EXTENSIONS env var
	args = append(args, "pi")

	return c.runDockerInteractive(args...)
}

// RunServe starts pi-agent in ttyd mode (detached)
func (c *Client) RunServe(workDir string) error {
	// Stop existing pi-agent if running
	running, _ := c.IsContainerRunning(AgentContainer)
	if running {
		fmt.Println("  Stopping existing pi-agent...")
		_, _ = c.runDocker("stop", AgentContainer)
		_, _ = c.runDocker("rm", AgentContainer)
	}

	args := []string{"run", "-d", "--rm", "-p", fmt.Sprintf("%d:7681", c.config.TtydPort)}
	args = append(args, c.piAgentRunArgs(workDir)...)
	args = append(args, c.config.PIAgentImage)

	if _, err := c.runDocker(args...); err != nil {
		return fmt.Errorf("failed to start pi-agent: %w", err)
	}

	fmt.Printf("  ✓ pi-agent started in ttyd mode\n")
	fmt.Printf("  Open http://localhost:%d in your browser\n", c.config.TtydPort)
	return nil
}

// StopAll stops and removes all picolo containers
func (c *Client) StopAll() error {
	for _, name := range []string{AgentContainer, LlamaContainer} {
		running, _ := c.IsContainerRunning(name)
		if running {
			fmt.Printf("  Stopping %s...\n", name)
			if _, err := c.runDocker("stop", name); err != nil {
				return fmt.Errorf("failed to stop %s: %w", name, err)
			}
		}
		if _, err := c.runDocker("rm", "-f", name); err != nil {
			return fmt.Errorf("failed to remove %s: %w", name, err)
		}
	}
	return nil
}

// RestartContainer restarts a specific container
func (c *Client) RestartContainer(name string) error {
	if _, err := c.runDocker("stop", name); err != nil {
		return fmt.Errorf("failed to stop %s: %w", name, err)
	}
	if _, err := c.runDocker("rm", "-f", name); err != nil {
		return fmt.Errorf("failed to remove %s: %w", name, err)
	}

	switch name {
	case LlamaContainer:
		if !c.config.IsLlamaEnabled() {
			return nil
		}
		_, err := c.runDocker(c.llamaRunArgs()...)
		return err
	case AgentContainer:
		// pi-agent is not restarted in detached mode during update
		// (it's only started by chat/serve)
		return nil
	default:
		return fmt.Errorf("unknown container: %s", name)
	}
}

// IsContainerRunning checks if a container is running
func (c *Client) IsContainerRunning(name string) (bool, error) {
	output, err := c.runDocker("ps",
		"--filter", fmt.Sprintf("name=%s", name),
		"--filter", "status=running",
		"--format", "{{.Names}}",
	)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) == name, nil
}

// CheckDocker verifies docker is available
func CheckDocker() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker is not installed or not in PATH")
	}

	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker is not running: %w", err)
	}

	return nil
}
