# Contributing

## Getting started

First off, thanks for your interest! All contributions are welcome. Here's some info to help you get a good start:

* Check go.mod for which Go version we're using. As of this writing, that's 1.24
* Install the pre-commit hook by installing pre-commit and then running pre-commit install in the root directory of the repo.
*  If you use VSCode, there's already an example launch config for debugging.
*  Although large language model use is strongly discouraged, it is not forbidden (but this could change at any time). you will not be given special leniency for counterfeit code that must be tediously tweaked after you make your PR due to inept AI. Weigh the efficacy of your tools against the new, additional work they make for you by using them.
 * Colors should follow the general guide below to help with accessibility.

| Function     | Color         | Purpose / Usage Description                             |
|--------------|---------------|----------------------------------------------------------|
| `Success`    | Green         | Indicates a successful operation or positive result.     |
| `Content`    | Magenta       | Used for general output or primary message content.      |
| `Note`       | Cyan          | Marks informational notes or helpful context.            |
| `Attention`  | Yellow        | Highlights warnings, required actions, or caution areas. |
| `Aside`      | Dark Grey     | Used for less important info or inline side notes.       |
| `Emphasised` | Bold + Italic | Emphasizes important text via formatting (not color).    |
| `Custom`     | Custom Color  | Allows specifying any custom color for special use.      |



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
