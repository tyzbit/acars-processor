# Contributing guide

## Overview

Thank you for your interest in contributing to ACARS processor! This guide provides comprehensive information for developers who want to contribute code, documentation, or other improvements to the project.

## Table of contents

- [Getting started](#getting-started)
- [Development environment setup](#development-environment-setup)
- [Code organization](#code-organization)
- [Development workflow](#development-workflow)
- [Testing guidelines](#testing-guidelines)
- [Code style and standards](#code-style-and-standards)
- [Contribution process](#contribution-process)
- [Issue reporting](#issue-reporting)
- [Feature requests](#feature-requests)

## Getting started

### Prerequisites

**Required software**:
- Go 1.24 or later
- Git

**Optional but recommended**:
- Docker and Docker Compose (for integration testing)
- A code editor with Go support (VS Code recommended)
- golangci-lint for code quality checks
- pre-commit for automated code quality enforcement
- Air for live reloading during development
- Delve for debugging

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
   
   # Install development tools if available
   # Note: No automated tests exist yet, but manual testing is important
   ```

3. **Verify your build**:
   ```bash
   go build -o acars-processor
   ./acars-processor -s  # Test schema generation
   ```

4. **Make your changes and submit a pull request** (see detailed workflow below)

> **Note**: This project currently has limited test coverage. Contributors are encouraged to test manually and submit features without comprehensive test suites, though testing improvements are welcome.

## Development environment setup

### Manual setup

1. **Install pre-commit hooks**:
   ```bash
   # Install pre-commit
   pip install pre-commit
   
   # Install hooks
   pre-commit install
   
   # Test hooks
   pre-commit run --all-files
   ```

2. **Install Go development tools**:
   ```bash
   # Linting and formatting
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   
   # Live reloading for development
   go install github.com/cosmtrek/air@latest
   
   # Debugging
   go install github.com/go-delve/delve/cmd/dlv@latest
   
   # Testing tools
   go install github.com/onsi/ginkgo/v2/ginkgo@latest
   go install gotest.tools/gotestsum@latest
   ```

3. **VS Code setup**:
   
   Install recommended extensions:
   - Go (official Google extension)
   - YAML (Red Hat)
   - Docker
   - GitLens
   
   The repository includes VS Code configuration in `.vscode/`:
   ```json
   {
     "go.lintTool": "golangci-lint",
     "go.formatTool": "goimports",
     "go.useLanguageServer": true,
     "go.delveConfig": {
       "debugAdapter": "legacy"
     }
   }
   ```

### Docker development environment

1. **Build development container**:
   ```bash
   docker build -t acars-processor-dev -f Dockerfile.dev .
   ```

2. **Run with development mounts**:
   ```bash
   docker run -it \
     -v $(pwd):/workspace \
     -v go-mod-cache:/go/pkg/mod \
     -p 8080:8080 \
     acars-processor-dev
   ```

### Test data setup

1. **Mock ACARSHub for testing**:
   ```bash
   # Start mock services
   docker-compose -f docker-compose.test.yml up -d
   
   # This starts:
   # - Mock ACARSHub with sample data
   # - Test MySQL database
   # - Mock external APIs
   ```

2. **Sample configuration for development**:
   ```bash
   cp config_example_local.yaml config_local.yaml
   # Edit config_local.yaml with development-specific settings
   ```

## Code organization

### Project structure

```
acars-processor/
├── docs/                    # Documentation
├── .vscode/                 # VS Code configuration
├── .github/                 # GitHub workflows and templates
├── cmd/                     # Future: Command-line interfaces
├── internal/                # Future: Internal packages
├── pkg/                     # Future: Public packages
├── test/                    # Test utilities and fixtures
├── scripts/                 # Build and deployment scripts
├── config_example.yaml      # Example configuration
├── main.go                  # Application entry point
├── *.go                     # Core application files
└── go.mod                   # Go module definition
```

### Core components

**Message processing pipeline**:
- `acarshub.go`: ACARSHub integration and message ingestion
- `annotators.go`, `annotator_*.go`: Message enrichment
- `filters.go`, `filter_*.go`: Message filtering
- `receivers.go`, `receiver_*.go`: Message delivery

**Configuration and infrastructure**:
- `config.go`: Configuration management
- `schema.go`: JSON schema generation
- `db.go`: Database abstraction
- `types.go`: Core data structures
- `util.go`: Utility functions

**Supporting modules**:
- `logging.go`: Structured logging setup
- `main.go`: Application bootstrapping

### Design patterns

**Plugin architecture**:
- Annotators implement message-type-specific interfaces
- Receivers implement common `Receiver` interface
- Filters implement message-type-specific filter interfaces
- Configuration-driven component registration

**Error handling**:
- Fail-safe approach for filters (default to not filtering)
- Graceful degradation for annotator failures
- Retry logic with exponential backoff for external services

**Concurrency**:
- Channel-based message queuing
- Configurable worker pools
- Context-based cancellation

## Development workflow

### Development process

1. **Create feature branch**:
   ```bash
   git checkout -b feature/add-prometheus-metrics
   ```

2. **Make changes following code standards** (see below)

3. **Add/update tests**:
   ```bash
   # Run tests during development
   go test ./...
   
   # Run tests with coverage
   go test -cover ./...
   
   # Run specific test
   go test -run TestAnnotatorName ./...
   ```

4. **Update documentation**:
   - Update relevant markdown files in `docs/`
   - Update code comments and godoc
   - Update configuration examples if needed

5. **Commit changes**:
   ```bash
   # Stage changes
   git add .
   
   # Commit with descriptive message
   git commit -m "feat: add Prometheus metrics endpoint
   
   - Add /metrics endpoint for application metrics
   - Include message processing rate and error counters
   - Update configuration schema for metrics settings
   - Add unit tests for metrics collection"
   ```

6. **Keep branch updated**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

7. **Push and create pull request**:
   ```bash
   git push origin feature/add-prometheus-metrics
   # Create PR through GitHub UI
   ```

### Commit message conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

**Format**: `type(scope): subject`

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples**:
```
feat(annotators): add aircraft manufacturer lookup
fix(filters): handle nil pointer in emergency detection
docs(api): update webhook payload documentation
refactor(db): simplify connection management
test(receivers): add Discord webhook integration tests
```

### Live development

**Using Air for live reloading**:
```bash
# Create .air.toml configuration
cat > .air.toml << EOF
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/acars-processor ."
  bin = "./tmp/acars-processor -c config_local.yaml"
  include_ext = ["go", "yaml"]
  exclude_dir = ["tmp", "vendor", ".git"]

[log]
  time = true
EOF

# Start live development
air
```

**Manual development cycle**:
```bash
# Build and run
go build -o acars-processor .
./acars-processor -c config_local.yaml

# In another terminal, watch logs
tail -f acars-processor.log | jq .  # If using JSON logging
```

## Testing guidelines

### Test structure

**Unit tests**:
- Test individual functions and methods
- Mock external dependencies
- Focus on core functionality correctness

**Integration tests**:
- Test component interactions
- Use real database connections
- Mock external APIs

**End-to-end tests**:
- Test complete message processing pipeline
- Use Docker Compose test environment
- Verify external integrations

### Testing patterns

1. **Table-driven tests**:
   ```go
   func TestFilterACARSMessage(t *testing.T) {
       tests := []struct {
           name     string
           message  ACARSMessage
           config   FilterConfig
           expected bool
       }{
           {
               name: "emergency message should not be filtered",
               message: ACARSMessage{
                   MessageText: "EMERGENCY ENGINE FAILURE",
                   Emergency:   true,
               },
               config:   FilterConfig{Emergency: true},
               expected: false,
           },
           // ... more test cases
       }
   
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               result := FilterMessage(tt.message, tt.config)
               assert.Equal(t, tt.expected, result)
           })
       }
   }
   ```

2. **Mock external dependencies**:
   ```go
   type MockHTTPClient struct {
       responses map[string]*http.Response
   }
   
   func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
       if resp, ok := m.responses[req.URL.String()]; ok {
           return resp, nil
       }
       return nil, errors.New("not found")
   }
   ```

3. **Test fixtures**:
   ```go
   // test/fixtures/messages.go
   var (
       SampleACARSMessage = ACARSMessage{
           AircraftTailCode: "N123AB",
           FlightNumber:     "AA1234",
           MessageText:      "REQUEST GATE INFO",
       }
       
       EmergencyACARSMessage = ACARSMessage{
           AircraftTailCode: "N456CD",
           FlightNumber:     "UA5678",
           MessageText:      "EMERGENCY ENGINE FAILURE",
           Emergency:        true,
       }
   )
   ```

### Test execution

**Run all tests**:
```bash
go test ./...
```

**Run tests with coverage**:
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Run specific tests**:
```bash
# Run tests for specific package
go test ./annotators/

# Run specific test function
go test -run TestACARSAnnotator ./...

# Run tests matching pattern
go test -run "TestFilter.*Emergency" ./...
```

**Benchmark tests**:
```bash
go test -bench=. ./...
go test -bench=BenchmarkMessageProcessing -benchmem ./...
```

### Test environment

**Database testing**:
```go
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    err = db.AutoMigrate(&ACARSMessage{}, &VDLM2Message{})
    require.NoError(t, err)
    
    return db
}
```

**Configuration testing**:
```go
func TestConfig() *Config {
    return &Config{
        ACARSProcessorSettings: ACARSProcessorSettings{
            LogLevel: "debug",
            Database: DatabaseConfig{
                Type: "sqlite",
                SQLiteDatabasePath: ":memory:",
            },
        },
    }
}
```

## Code style and standards

### Go coding standards

**Follow effective Go principles**:
- Use `gofmt` for formatting
- Use `goimports` for import management
- Follow Go naming conventions
- Write idiomatic Go code

**Specific guidelines**:

1. **Package organization**:
   ```go
   // Good: focused functionality with clear responsibility  
   package annotators
   
   // Note: All code currently in main package
   ```

2. **Error handling**:
   ```go
   // Good: explicit error handling
   result, err := externalAPI.Call()
   if err != nil {
       log.Error("external API call failed", "error", err)
       return nil, fmt.Errorf("failed to get external data: %w", err)
   }
   
   // Avoid: ignoring errors
   result, _ := externalAPI.Call()
   ```

3. **Documentation**:
   ```go
   // Good: clear function documentation
   // NormalizeAircraftRegistration standardizes aircraft registration codes
   // by removing common separators and converting to lowercase
   func NormalizeAircraftRegistration(reg string) string {
       // Implementation...
   }
   ```

4. **Interface design**:
   ```go
   // Good: small, focused interfaces
   type MessageProcessor interface {
       ProcessMessage(Message) error
   }
   
   // Avoid: large interfaces
   type MessageHandler interface {
       ProcessMessage(Message) error
       ValidateMessage(Message) bool
       TransformMessage(Message) Message
       SaveMessage(Message) error
   }
   ```

### Documentation standards

**Package documentation**:
```go
// Package annotators provides message enrichment capabilities for ACARS processor.
//
// Annotators receive raw ACARS/VDLM2 messages and enrich them with additional
// data from external sources such as aircraft tracking APIs and AI services.
package annotators
```

**Function documentation**:
```go
// AnnotateACARSMessage enriches an ACARS message with additional data.
//
// The function calls external APIs to gather aircraft position, flight information,
// and other relevant data. If any external call fails, partial annotation data
// is returned to ensure message processing continues.
//
// Returns nil if no annotation data could be gathered.
func (a *Annotator) AnnotateACARSMessage(msg ACARSMessage) Annotation {
    // Implementation
}
```

**Configuration documentation**:
```go
type Config struct {
    // APIKey is the authentication key for external service access.
    // Required when annotator is enabled.
    APIKey string `jsonschema:"required" yaml:"api_key"`
    
    // Timeout specifies the maximum duration to wait for API responses.
    // Default: 10 seconds.
    Timeout time.Duration `default:"10s" yaml:"timeout"`
}
```

### Linting configuration

This file doesn't exist. Standard Go linting however devs want to do it is fine. If you want to keep this, please also create a `.golangci.yml`
```yaml
linters-settings:
  gocyclo:
    min-complexity: 15
  
  govet:
    check-shadowing: true
  
  lll:
    line-length: 120
  
  misspell:
    locale: US

linters:
  enable:
    - bodyclose
    - deadcode
    - depguard
    - dogsled
    - errcheck
    - gocyclo
    - gofmt
    - goimports
    - golint
    - gomnd
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - lll
    - misspell
    - nakedret
    - rowserrcheck
    - staticcheck
    - structcheck
    - stylecheck
    - typecheck
    - unconvert
    - unparam
    - unused
    - varcheck
    - whitespace

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - lll
```

### Security considerations

**Secure coding practices**:

1. **Input validation**:
   ```go
   func ValidateConfig(cfg *Config) error {
       if cfg.APIKey == "" {
           return errors.New("API key is required")
       }
       
       if cfg.Timeout <= 0 {
           return errors.New("timeout must be positive")
       }
       
       return nil
   }
   ```

2. **Secret handling**:
   ```go
   // Good: use environment variables
   apiKey := os.Getenv("EXTERNAL_API_KEY")
   
   // Avoid: hardcoded secrets
   const apiKey = "sk-1234567890abcdef"
   ```

3. **HTTP client security**:
   ```go
   client := &http.Client{
       Timeout: 30 * time.Second,
       Transport: &http.Transport{
           TLSClientConfig: &tls.Config{
               MinVersion: tls.VersionTLS12,
           },
       },
   }
   ```

### Accessibility guidelines

**Color usage in output**:
Follow the established color coding standards:

| Type of event | Color | Usage |
|---------------|-------|-------|
| Success | Green | Successful operations |
| Information | White | General status |
| Unusual but OK | Cyan | Non-critical warnings |
| Needs attention | Yellow | Issues requiring review |
| Raw data | Black | Verbose output |
| Additional info | Bold + Italic | Supplementary details |

**Implementation**:
```go
import "github.com/fatih/color"

var (
    Success     = color.New(color.FgGreen).SprintFunc()
    Information = color.New(color.FgWhite).SprintFunc()
    Warning     = color.New(color.FgYellow).SprintFunc()
    Error       = color.New(color.FgRed).SprintFunc()
)
```

## Contribution process

### Pull request process

1. **Before submitting**:
   - [ ] All tests pass locally
   - [ ] Code follows style guidelines
   - [ ] Documentation is updated
   - [ ] Commit messages follow conventions
   - [ ] Branch is up to date with main

2. **Pull request description template**:
   ```markdown
   ## Description
   Brief description of changes and motivation.
   
   ## Type of change
   - [ ] Bug fix (non-breaking change which fixes an issue)
   - [ ] New feature (non-breaking change which adds functionality)
   - [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
   - [ ] Documentation update
   
   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests added/updated
   - [ ] Manual testing performed
   
   ## Checklist
   - [ ] Code follows the style guidelines
   - [ ] Self-review of code completed
   - [ ] Code is commented, particularly in hard-to-understand areas
   - [ ] Corresponding changes to documentation made
   - [ ] Changes generate no new warnings
   - [ ] Tests added that prove fix/feature works
   - [ ] New and existing tests pass locally
   ```

3. **Review process**:
   - Automated checks must pass (CI/CD pipeline)
   - At least one maintainer review required
   - Address review feedback
   - Squash commits if requested

### Code review guidelines

**For reviewers**:

1. **Focus areas**:
   - Correctness and logic
   - Performance implications
   - Security considerations
   - Code maintainability
   - Test coverage

2. **Review checklist**:
   - [ ] Code solves the stated problem
   - [ ] Implementation follows Go best practices
   - [ ] Error handling is appropriate
   - [ ] Tests adequately cover new code
   - [ ] Documentation is clear and complete
   - [ ] No obvious security vulnerabilities
   - [ ] Performance impact is acceptable

3. **Feedback guidelines**:
   - Be constructive and specific
   - Suggest improvements, not just problems
   - Distinguish between must-fix and suggestions
   - Provide examples when helpful

**For contributors**:

1. **Responding to feedback**:
   - Address each review comment
   - Ask for clarification when needed
   - Mark conversations as resolved when addressed
   - Update PR description if scope changes

2. **Making changes**:
   - Make atomic commits for each review item
   - Avoid force-pushing during review
   - Keep discussion in PR comments, not commit messages

## Issue reporting

### Bug reports

**Use the bug report template**:
```markdown
## Bug description
A clear and concise description of what the bug is.

## To reproduce
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '...'
3. See error

## Expected behavior
A clear description of what you expected to happen.

## Environment
- OS: [e.g., Ubuntu 20.04]
- Go version: [e.g., 1.24.0]
- ACARS processor version: [e.g., v1.2.3]
- Deployment method: [e.g., Docker, native]

## Configuration
```yaml
# Relevant configuration sections (redact sensitive data)
```

## Logs
```
# Relevant log output
```

## Additional context
Add any other context about the problem here.
```

### Performance issues

**Include performance data**:
- CPU and memory usage profiles
- Message processing rates
- External API latency measurements
- Database query performance

**Example**:
```bash
# Generate CPU profile
go tool pprof http://localhost:8080/debug/pprof/profile

# Generate memory profile
go tool pprof http://localhost:8080/debug/pprof/heap

# Include output in issue
```

## Feature requests

### Feature request template

```markdown
## Is your feature request related to a problem?
A clear description of what the problem is. Ex. I'm always frustrated when [...]

## Describe the solution you'd like
A clear description of what you want to happen.

## Describe alternatives you've considered
A clear description of any alternative solutions or features you've considered.

## Additional context
Add any other context or screenshots about the feature request here.

## Implementation notes
If you have ideas about how this could be implemented, include them here.
```

### Feature development process

1. **Discussion phase**:
   - Create feature request issue
   - Discuss approach with maintainers
   - Agree on implementation plan

2. **Design phase**:
   - Create design document if complex
   - Update architecture documentation
   - Plan testing approach

3. **Implementation phase**:
   - Follow development workflow
   - Implement in iterations
   - Regular progress updates

4. **Review and integration**:
   - Comprehensive testing
   - Documentation updates
   - Performance validation

## Large language model policy

**Current policy**:
- LLM use is discouraged but not forbidden
- May change at any time
- No special leniency for AI-generated code issues
- Generated code must be thoroughly reviewed and tested
- Contributors responsible for code quality regardless of generation method

**Guidelines for LLM use**:
1. Use for learning and exploration, not production code
2. Always review and understand generated code
3. Test thoroughly - AI code often has subtle bugs
4. Follow all other contribution guidelines
5. Don't rely on AI for architectural decisions

**What works well with LLMs**:
- Learning Go patterns and idioms
- Understanding existing codebase
- Generating test cases and fixtures
- Writing documentation drafts

**What to avoid**:
- Generating complex business logic
- Security-sensitive code
- Performance-critical sections
- Integration code for external APIs

Thank you for contributing to ACARS processor! Your efforts help make aviation communication monitoring more accessible and powerful for the community.
