# picolo

**Pi** + **Co**ntainer + **Lo**cal — a quick, easy, and safe sandboxed environment to play with the pi coding agent.

A standalone, portable Go executable that spins up pi in Docker containers. Local AI is optional but pretty cool, and picolo makes it easy to set up.

## Overview

Picolo gives you an isolated, reproducible environment to experiment with the pi coding agent. Everything runs in Docker containers on a private network:

- **pi-agent** — the pi coding agent, mounted to any directory you choose
- **llama.cpp server** _(optional)_ — local LLM inference on your GPU, served via an OpenAI-compatible API

No host pollution, no leftover files — just containers on a clean network.

## Installation

```bash
# Build from source
cd picolo
go build -o picolo .
sudo mv picolo /usr/local/bin/

# Or install directly
go install github.com/wchomik/picolo@latest
```

## Quick Start

```bash
# Initialize the environment (pulls images, starts llama server)
picolo init

# Start the agent interactively in a directory
picolo chat ~/my-project

# Serve the agent through a browser
picolo serve ~/my-project

# Check container status
picolo status

# Stop all containers
picolo stop

# Resume last session (chat or serve)
picolo start

# Update all containers to newest versions
picolo update
```

## Commands

| Command | Description |
|---|---|
| `picolo init` | Initialize environment (pull images, start llama server) |
| `picolo init --env vulkan` | Initialize with Vulkan GPU backend |
| `picolo init --skip-llama` | Skip local llama.cpp deployment |
| `picolo start` | Start containers and resume last session (chat/serve) |
| `picolo status` | Show status of all picolo containers |
| `picolo stop` | Stop all running picolo containers |
| `picolo update` | Update all containers to newest versions |
| `picolo chat [dir]` | Start pi agent interactively in a terminal |
| `picolo serve [dir]` | Start pi agent via ttyd (browser access) |

## Configuration

All configuration is stored in `~/.picolo/`:

```yaml
# ~/.picolo/picolo.yaml
env: cuda                    # GPU backend: cuda, vulkan, rocm, metal, cpu
skip_llama: false            # Skip local llama.cpp deployment
extensions:                  # Pi extensions to install
  - npm:pi-observability
  - npm:pi-web-access
pi_agent_image: ghcr.io/wchomik/pi-docker:latest
ttyd_port: 7681              # Browser terminal port
llama_port: 8080             # Llama.cpp API port
```

### GPU Environments

| Environment | Description |
|---|---|
| `cuda` | NVIDIA GPU (default) |
| `vulkan` | Vulkan-compatible GPU |
| `rocm` | AMD GPU |
| `metal` | Apple Silicon (macOS) |
| `cpu` | CPU-only inference |

### Directory Structure

```
~/.picolo/
├── picolo.yaml              # Configuration file
└── pi/                      # Pi agent config directory
```

## Docker Network

Picolo creates a `picolo-net` Docker network for inter-container communication.
Containers communicate by name (e.g., `http://picolo-llama-cpp:8080/v1`).

## Architecture

```
┌─────────────────────┐         ┌────────────────────────┐
│   llama-cpp server  │◄────────│     pi-agent           │
│   (GPU, port 8080)  │  HTTP   │  (pi coding agent)     │
│   ~/.models:/models │────────►│  PI_LLM_ENDPOINT       │
│   (optional)        │         │  http://picolo-llama-  │
└─────────────────────┘         │      cpp:8080/v1       │
                                │  HOST_PATH:/app        │
                                │  PI_EXTENSIONS (env)   │
                                └───────────┬────────────┘
                                            │
                                    ┌───────┴────────┐
                                    │  ttyd (port     │
                                    │  7681)          │
                                    │  (serve mode)   │
                                    └────────────────┘
```

## Requirements

### Software

- **Docker** — picolo runs everything in containers (no Docker Compose needed)
- **Go 1.24+** — only required if building from source

### Hardware

The pi-agent container runs on any machine. For local LLM inference (optional), typical hardware requirements for running GGUF models apply:

- **GPU** — NVIDIA (CUDA), AMD (ROCm), or Apple Silicon (Metal) recommended
- **RAM** — 8 GB minimum, 16 GB+ for larger models
- **Storage** — depends on model size (4–20 GB typical)

If you don't have suitable hardware, use `--skip-llama` and connect pi to any remote LLM endpoint.

## License

MIT — Copyright (c) 2026 Wiktor Chomik
