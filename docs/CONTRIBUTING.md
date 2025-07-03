# Contributing

## Getting started

You'll need Go 1.24+ and Git. Docker is optional for container testing.

### Setup

1. Fork and clone:
   ```bash
   git clone https://github.com/tyzbit/acars-processor.git
   cd acars-processor
   ```

2. Build and test:
   ```bash
   go mod download
   go build -o acars-processor
   cp config_example.yaml config.yaml
   ./acars-processor -c config.yaml
   ```

3. Make changes on a feature branch:
   ```bash
   git checkout -b feature/your-change
   # Make your changes
   git commit -am "Description of changes"
   git push origin feature/your-change
   ```

4. Open a pull request with a clear description.

## Code style

- Use `go fmt` for formatting
- Write clear variable and function names
- Comment complex logic
- Keep functions focused

## Bug reports

Include:
- Clear description and steps to reproduce
- Expected vs actual behavior  
- Log output (with sensitive data removed)
- Configuration (with secrets removed)
