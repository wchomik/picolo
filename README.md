# picolo

Orchestrate a local, sandboxed Docker-based pi agent environment.

A standalone, portable Go executable that replaces the pi-local-sandbox Makefile with a clean CLI for managing llama.cpp servers and pi agent containers.

## Overview

Picolo manages two Docker services:

- **llama.cpp server** — serves local LLM models on a GPU via an OpenAI-compatible API
- **pi-agent** — the pi coding agent, configured to use the local LLM endpoint

Everything runs in Docker, making it easy to spin up a private coding assistant.

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

# Update all containers to newest versions
picolo update
```

## Commands

| Command | Description |
|---|---|
| `picolo init` | Initialize environment (pull images, start llama server) |
| `picolo init --env vulkan` | Initialize with Vulkan GPU backend |
| `picolo init --skip-llama` | Skip local llama.cpp deployment |
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
├── docker-compose.yaml      # Generated compose file
└── pi/                      # Pi agent config directory
```

## Architecture

```
┌─────────────────────┐         ┌────────────────────────┐
│   llama-cpp server  │◄────────│     pi-agent           │
│   (GPU, port 8080)  │  HTTP   │  (pi coding agent)     │
│   ~/.models:/models │────────►│  PI_LLM_ENDPOINT       │
└─────────────────────┘         │  http://llama-cpp:8080 │
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

- Go 1.24+ (for building)
- Docker & Docker Compose
- GPU (for llama.cpp acceleration, optional with `--skip-llama`)
- Local LLM models in `~/.models`

## License

MIT — Copyright (c) 2026 Wiktor Chomik
