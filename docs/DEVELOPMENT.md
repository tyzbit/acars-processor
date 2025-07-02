# Development tools and workflow

## Overview

ACARS processor includes comprehensive development tools and automation to ensure code quality, streamline development workflow, and support multiple deployment environments. This document covers the tooling ecosystem, IDE configuration, and automated processes.

## Development environment setup

### IDE configuration

#### Visual Studio Code

The project includes VS Code configuration for optimal development experience:

**Launch configurations** (`.vscode/launch.json.example`):
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Package",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${fileDirname}",
      "console": "integratedTerminal"
    },
    {
      "name": "Generate Schema",
      "type": "go",
      "request": "launch",
      "mode": "auto", 
      "program": "${fileDirname}",
      "args": ["-s"],
      "console": "integratedTerminal"
    }
  ]
}
```

**Setup instructions**:
1. Copy the example configuration:
   ```bash
   cp .vscode/launch.json.example .vscode/launch.json
   ```

2. Install recommended Go extension for VS Code

3. Configure environment variables in launch configuration:
   ```json
   {
     "name": "Launch Package",
     "env": {
       "DISCORD_WEBHOOK_URL": "your-webhook-url",
       "OPENAI_API_KEY": "your-api-key"
     }
   }
   ```

**Debugging features**:
- **Launch Package**: Run application with full debugging support
- **Generate Schema**: Execute schema generation with debugging
- Breakpoint support for all Go source files
- Variable inspection and call stack analysis
- Integrated terminal for command execution

### Code quality tools

#### Pre-commit hooks

Automated code quality checks using pre-commit framework:

**Installation**:
```bash
# Install pre-commit (requires Python)
pip install pre-commit

# Install hooks in repository
pre-commit install
```

**Configuration** (`.pre-commit-config.yaml`):
```yaml
repos:
  - repo: local
    hooks:
      - id: generate-schema
        name: Generate schema
        description: Auto-generates schema.json and config_example.yaml
        language: script
        entry: .pre-commit/generate-schema
```

**Schema generation hook** (`.pre-commit/generate-schema`):
```bash
#!/usr/bin/env bash
set -e
go run ./*.go -s
```

**Workflow integration**:
- Automatically runs on every git commit
- Generates updated schema and configuration files
- Prevents commits with outdated generated files
- Ensures consistency between code and configuration

### Build and containerization

#### Docker configuration

**Multi-stage Dockerfile**:
```dockerfile
FROM golang:1.24-alpine as build

LABEL org.opencontainers.image.source="https://github.com/tyzbit/acars-processor"

WORKDIR /
COPY . ./

RUN apk add \
    build-base \
    git \
    &&  go build -ldflags="-s -w"

FROM alpine

COPY --from=build /acars-processor /

CMD ["/acars-processor"]
```

**Build features**:
- **Multi-stage build**: Reduces final image size by excluding build dependencies
- **Static linking**: Creates self-contained binary with `-ldflags="-s -w"`
- **Alpine base**: Minimal runtime environment for security and size
- **Build dependencies**: Includes git and build-base for CGO compilation

**Local development build**:
```bash
# Build development binary
go build -o acars-processor

# Build optimized release binary
go build -ldflags="-s -w" -o acars-processor

# Build Docker image
docker build -t acars-processor:dev .

# Run with Docker Compose
docker-compose up --build
```

## Continuous Integration/Continuous Deployment

### GitHub Actions workflows

#### Build and push workflow

**Trigger conditions** (`.github/workflows/build.yaml`):
- Push to main branch
- Git tag creation
- Pull request validation

**Workflow features**:
```yaml
name: Build and Push Docker Images with Version Tag

on:
  push:
    branches: [main]
    tags: ["*"]

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - name: Check out code
        uses: actions/checkout@v4
        
      - name: Login to GitHub Container Registry
        run: echo ${{ secrets.GITHUB_TOKEN }} | docker login ghcr.io -u ${{ github.repository_owner }} --password-stdin
        
      - name: Build and push images
        run: |
          if [[ "$GITHUB_REF" == "refs/heads/main" ]]; then
            docker build -t ghcr.io/${{ github.repository }}:latest .
            docker push ghcr.io/${{ github.repository }}:latest
          else
            docker build -t ghcr.io/${{ github.repository }}:${{ env.TAG_NAME }} .
            docker push ghcr.io/${{ github.repository }}:${{ env.TAG_NAME }}
          fi
```

**Image publishing**:
- **Latest tag**: Published on main branch commits
- **Version tags**: Published when Git tags are created
- **Registry**: GitHub Container Registry (ghcr.io)
- **Automatic builds**: No manual intervention required

#### Release workflow

**Binary release automation** (`.github/workflows/release.yaml`):
```yaml
on:
  release:
    types: [created]

jobs:
  releases-matrix:
    strategy:
      matrix:
        goos: [linux, windows, darwin]
        goarch: ["386", amd64, arm64]
        exclude:
          - goarch: "386"
            goos: darwin
          - goarch: arm64  
            goos: windows
```

**Release features**:
- **Multi-platform builds**: Linux, Windows, macOS
- **Multiple architectures**: 386, amd64, arm64
- **Automated asset naming**: Includes platform and architecture
- **Checksum generation**: SHA256 checksums for all binaries
- **Asset bundling**: Includes LICENSE and README.md
- **Go version pinning**: Uses Go 1.23.6 for reproducible builds

### Schema generation automation

#### Automatic schema updates

**Pre-commit integration**:
The schema generation is integrated into the development workflow:

1. **Developer workflow**:
   ```bash
   # Make configuration changes
   vim config.go
   
   # Commit triggers pre-commit hook
   git add config.go
   git commit -m "Update configuration structure"
   
   # Pre-commit automatically:
   # - Generates new schema.json
   # - Updates config_example.yaml  
   # - Stages generated files
   ```

2. **Generated artifacts**:
   - **`schema.json`**: JSON schema for IDE validation and autocomplete
   - **`config_example.yaml`**: Complete example configuration with defaults
   - **YAML language server integration**: Schema URL for IDE support

3. **Validation workflow**:
   ```bash
   # Manual schema generation during development
   go run ./*.go -s
   
   # Verify schema in IDE (VS Code, IntelliJ)
   # Edit config.yaml with autocomplete support
   ```

## Testing and validation

### Configuration testing

**Schema validation testing**:
```bash
# Install JSON schema validator
npm install -g ajv-cli

# Validate example configuration
ajv validate -s schema.json -d config_example.yaml
```

**Integration testing with Docker**:
```bash
# Test complete stack with Docker Compose
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Test specific configurations
docker run --rm -v $(pwd)/config_test.yaml:/config.yaml acars-processor:dev -c /config.yaml
```

### Manual testing procedures

**Local development testing**:
1. **Configuration validation**:
   ```bash
   # Test configuration parsing
   ./acars-processor -c config.yaml --dry-run
   ```

2. **Connection testing**:
   ```bash
   # Test ACARSHub connectivity
   telnet acarshub-host 15550
   
   # Test external API access
   curl -H "X-API-Key: $ADSB_API_KEY" "https://adsbexchange.com/api/aircraft/json/"
   ```

3. **End-to-end validation**:
   ```bash
   # Run with debug logging
   LOG_LEVEL=debug ./acars-processor -c config.yaml
   
   # Monitor message processing
   docker-compose logs -f acars-processor | grep "Processing message"
   ```

## Deployment automation

### Container registry integration

**GitHub Container Registry**:
- **Automatic publishing**: Triggered by Git tags and main branch updates
- **Multi-architecture support**: ARM64 and AMD64 images available
- **Public registry**: No authentication required for pulling images
- **Version management**: Git tags create corresponding container tags

**Usage examples**:
```bash
# Pull latest development image
docker pull ghcr.io/tyzbit/acars-processor:latest

# Pull specific version
docker pull ghcr.io/tyzbit/acars-processor:v1.2.3

# Use in Docker Compose
services:
  acars-processor:
    image: ghcr.io/tyzbit/acars-processor:latest
```

### Release management

**Semantic versioning**:
- Git tags follow semantic versioning (v1.2.3)
- Pre-release tags supported (v1.2.3-rc1)
- Automatic changelog generation from commit messages

**Release checklist**:
1. Update version documentation
2. Create Git tag: `git tag v1.2.3`
3. Push tag: `git push origin v1.2.3`
4. GitHub Actions automatically:
   - Builds multi-platform binaries
   - Publishes container images
   - Creates GitHub release with assets

## Development workflow best practices

### Local development setup

1. **Environment preparation**:
   ```bash
   # Clone repository
   git clone https://github.com/tyzbit/acars-processor.git
   cd acars-processor
   
   # Install dependencies
   go mod download
   
   # Setup development tools
   pip install pre-commit
   pre-commit install
   
   # Copy IDE configuration
   cp .vscode/launch.json.example .vscode/launch.json
   ```

2. **Configuration management**:
   ```bash
   # Create development configuration
   cp config_example.yaml config_dev.yaml
   
   # Set development environment variables
   export ACARSHUB_HOST=localhost
   export LOG_LEVEL=debug
   ```

3. **Development cycle**:
   ```bash
   # Make changes
   vim main.go
   
   # Generate schema if configuration changed
   go run ./*.go -s
   
   # Test locally
   go run . -c config_dev.yaml
   
   # Commit (triggers pre-commit hooks)
   git add .
   git commit -m "Feature: Add new annotation capability"
   ```

### Code quality standards

**Automated quality checks**:
- **Pre-commit hooks**: Schema generation and validation
- **GitHub Actions**: Build verification and testing
- **Container scanning**: Automated security scanning of published images

**Manual quality checks**:
- **Code review**: Required for all pull requests
- **Integration testing**: Manual validation in development environment
- **Documentation updates**: Keep documentation synchronized with code changes

This development tooling ecosystem ensures consistent code quality, streamlined workflows, and reliable deployment processes while maintaining comprehensive automation and validation capabilities.
