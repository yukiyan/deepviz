# deepviz

[![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://github.com/yourusername/deepviz)

A CLI tool for Gemini API Deep Research and NanoBanana (infographics generation).

## Features

**End-to-end pipeline from Deep Research to NanoBanana infographics generation**

- 🔍 **Deep Research**: Conduct comprehensive research and analysis using Google's Deep Research API
- 🎨 **Infographics Generation**: Automatically transform research results into visual infographics with NanoBanana
- ⚡ **Flexible Workflow**: Execute full pipeline, research-only, or image-only modes
- 🔧 **Highly Configurable**: Customize via command-line flags, environment variables, or config file
- 🌍 **Multi-language Support**: Generate infographics in any language (default: Japanese)
- 📦 **XDG Base Directory Compliant**: Follows standard directory conventions

## Installation

### From source

```bash
make build
make install  # Install to ~/.local/bin/
```

### Multi-platform builds

Build for all supported platforms:

```bash
make build-all
```

This creates binaries in `dist/` for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Quick Start

1. **Set your API key**:
   ```bash
   export GEMINI_API_KEY="your-api-key"
   ```

2. **Run end-to-end pipeline** (Research → Infographics):
   ```bash
   deepviz --prompt "Kubernetes best practices"
   ```

3. **Check the output**:
   ```bash
   ls ~/.local/share/deepviz/
   ```

## Authentication

Gemini API key is required. Set it via environment variable:

```bash
export GEMINI_API_KEY="your-api-key"
```

You can also set it in the configuration file (`~/.config/deepviz/config.yaml`):

```yaml
api_key: your-api-key-here
```

Get your API key from: https://aistudio.google.com/apikey

## Usage Examples

### Full pipeline (Research → Image generation)

```bash
deepviz --prompt "Kubernetes best practices"
```

### Research only

```bash
deepviz --research-only --prompt "PostgreSQL performance tuning"
```

### Image generation only

```bash
deepviz --image-only --prompt "Microservices architecture overview diagram"
```

### Using a prompt file

```bash
deepviz --file prompt.txt
```

### Generate infographics in English

```bash
export GEMINI_IMAGE_LANG="English"
deepviz --prompt "Docker container best practices"
```

### Custom aspect ratio and size

```bash
deepviz --prompt "System architecture" --aspect-ratio 1:1 --image-size 4K
```

### Verbose logging for debugging

```bash
deepviz --verbose --prompt "Cloud security"
```

### Trace HTTP requests/responses (for deep debugging)

```bash
deepviz --trace --prompt "API design patterns"
```

## Configuration Management

### Initialize configuration file

```bash
deepviz config init
```

### Show current configuration

```bash
deepviz config show
```

### Configuration file location

`$XDG_CONFIG_HOME/deepviz/config.yaml` (default: `~/.config/deepviz/config.yaml`)

### Configuration file example

```yaml
# Output directory
output_dir: ~/.local/share/deepviz

# API authentication
api_key: your-api-key-here

# Deep Research settings
deep_research_agent: deep-research-pro-preview-12-2025
poll_interval: 10
poll_timeout: 600

# Image generation settings
model: gemini-3-pro-image-preview
aspect_ratio: "16:9"
image_size: 2K
image_lang: Japanese
```

### Configuration priority (highest to lowest)

1. Command-line flags
2. Environment variables
3. Configuration file (`config.yaml`)
4. Default values

## Command-Line Options

### Basic Options

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--prompt` | `-p` | Inline prompt text | - |
| `--file` | `-f` | Read prompt from file | - |
| `--output` | `-o` | Output directory | `~/.local/share/deepviz` |
| `--verbose` | `-v` | Enable verbose logging | `false` |
| `--trace` | - | Enable trace logging (includes HTTP request/response bodies) | `false` |

### Workflow Control

| Option | Description | Default |
|--------|-------------|---------|
| `--research-only` | Execute research only (skip image generation) | `false` |
| `--no-image` | Alias for `--research-only` | `false` |
| `--image-only` | Execute image generation only (skip research) | `false` |

### Image Generation Options

| Option | Description | Default | Available Values |
|--------|-------------|---------|------------------|
| `--model` | Image generation model | `gemini-3-pro-image-preview` | `gemini-3-pro-image-preview`, `gemini-2.0-flash-exp` |
| `--aspect-ratio` | Image aspect ratio | `16:9` | `16:9`, `4:3`, `1:1`, `9:16`, `3:4` |
| `--image-size` | Image resolution | `2K` | `2K` (2048x1152), `4K` (3840x2160) |

### Subcommands

| Command | Description |
|---------|-------------|
| `config show` | Display current configuration |
| `config init` | Initialize configuration file |
| `completion [bash\|zsh\|fish\|powershell]` | Generate shell completion script |

## Environment Variables

### Basic Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `GEMINI_API_KEY` | Gemini API key (required) | - |
| `GEMINI_OUTPUT_DIR` | Output directory | `~/.local/share/deepviz` |
| `GEMINI_MODEL` | Image generation model | `gemini-3-pro-image-preview` |
| `GEMINI_ASPECT_RATIO` | Image aspect ratio | `16:9` |
| `GEMINI_IMAGE_SIZE` | Image resolution | `2K` |
| `GEMINI_IMAGE_LANG` | Language for image generation | `Japanese` |

### Advanced Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `GEMINI_DEEP_RESEARCH_AGENT` | Deep Research agent name | `deep-research-pro-preview-12-2025` |
| `GEMINI_POLL_INTERVAL` | Polling interval in seconds | `10` |
| `GEMINI_POLL_TIMEOUT` | Polling timeout in seconds | `600` |

## Output

### Output directory structure

Default: `$XDG_DATA_HOME/deepviz` (typically `~/.local/share/deepviz`)

```
~/.local/share/deepviz/
├── research/
│   └── 20251224_103045.md              # Research result (Markdown)
├── images/
│   └── 20251224_103045.png             # Generated infographics
└── responses/
    └── 20251224_103045_image.json      # Image generation API response (JSON)
```

### File naming

All output files use timestamp format: `YYYYMMDD_HHMMSS` (e.g., `20251224_103045`)

### Custom output directory

You can customize the output directory:

```bash
# Via command-line flag
deepviz --output ./my-output --prompt "Custom output location"

# Via environment variable
export GEMINI_OUTPUT_DIR="./my-output"
deepviz --prompt "Custom output location"
```

## Shell Completion

Generate shell completion scripts:

### Bash

```bash
source <(deepviz completion bash)
echo "source <(deepviz completion bash)" >> ~/.bashrc
```

### Zsh

```bash
source <(deepviz completion zsh)
echo "source <(deepviz completion zsh)" >> ~/.zshrc
```

### Fish

```bash
deepviz completion fish | source
deepviz completion fish > ~/.config/fish/completions/deepviz.fish
```

### PowerShell

```powershell
deepviz completion powershell | Out-String | Invoke-Expression
```

## Development

```bash
make fmt       # Format code
make lint      # Static analysis
make test      # Run tests
make coverage  # Coverage report
make help      # Show help
```
