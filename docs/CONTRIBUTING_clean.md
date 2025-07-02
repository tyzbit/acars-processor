# Contributing guide

## Overview

Thank you for your interest in contributing to ACARS processor! This guide provides information for developers who want to contribute code or documentation improvements.

## Getting started

### Prerequisites

**Required software**:
- Go 1.24 or later
- Git

**Optional but recommended**:
- Docker (for container testing)
- A code editor with Go support (VS Code recommended)

### First contribution workflow

1. **Fork and clone the repository**:
   ```bash
   # Fork the repository on GitHub
   git clone https://github.com/YOUR_USERNAME/acars-processor.git
   cd acars-processor
   git remote add upstream https://github.com/tyzbit/acars-processor.git
   ```

2. **Set up development environment**:
   ```bash
   # Install dependencies
   go mod download
   
   # Build the application
   go build -o acars-processor
   ```

3. **Make your changes**:
   ```bash
   # Create a feature branch
   git checkout -b feature/your-feature-name
   
   # Make your changes
   # Test your changes locally
   
   # Commit your changes
   git add .
   git commit -m "Add your feature description"
   ```

4. **Submit a pull request**:
   ```bash
   # Push to your fork
   git push origin feature/your-feature-name
   
   # Open a pull request on GitHub
   ```

## Development workflow

### Building and running

```bash
# Build the application
go build -o acars-processor

# Run with example configuration
cp config_example.yaml config.yaml
./acars-processor -c config.yaml

# Generate configuration schema
./acars-processor -s
```

### Code style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and concise

## Contribution process

### Submitting changes

1. Fork the repository
2. Create a feature branch from main
3. Make your changes
4. Test your changes locally
5. Submit a pull request with clear description

### Pull request guidelines

- Provide a clear description of the changes
- Include any relevant issue numbers
- Test changes locally before submitting
- Be responsive to feedback and review comments

## Issue reporting

When reporting bugs or requesting features, please include:

- Clear description of the issue or feature request
- Steps to reproduce (for bugs)
- Expected vs actual behavior
- System information (OS, Go version, etc.)
- Relevant log output or error messages
- Configuration files (with sensitive data removed)

Thank you for contributing to ACARS processor!
