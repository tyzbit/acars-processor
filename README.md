# ACARS Processor

A high-performance Go daemon that processes Aircraft Communication Addressing and Reporting System (ACARS) and VHF Data Link Mode 2 (VDLM2) messages in real-time. Connect to ACARSHub, filter and enrich messages with AI-powered analysis, and deliver results to Discord, monitoring systems, or custom endpoints.

## Features

- **Real-time processing** of ACARS/VDLM2 message streams from ACARSHub
- **Intelligent filtering** with AI classification (OpenAI, local Ollama) and configurable criteria
- **Data enrichment** from ADS-B Exchange, tar1090, and custom annotations
- **Flexible delivery** to Discord webhooks, New Relic, and custom endpoints
- **Production-ready** with Docker support, database options, and comprehensive monitoring

## Quick Start

### Docker (Recommended)

```bash
# Download example files
curl -o docker-compose.yml https://raw.githubusercontent.com/tyzbit/acars-processor/main/docker-compose.example.yml
curl -o .env https://raw.githubusercontent.com/tyzbit/acars-processor/main/.env.example

# Configure your environment
nano .env

# Start services
docker-compose up -d

# Monitor logs
docker-compose logs -f acars-processor
```

### Binary Installation

```bash
# Download and setup
wget https://github.com/tyzbit/acars-processor/releases/latest/download/acars-processor-linux-amd64
chmod +x acars-processor-linux-amd64
wget https://raw.githubusercontent.com/tyzbit/acars-processor/main/config_example.yaml -O config.yaml

# Configure and run
nano config.yaml
./acars-processor-linux-amd64 -c config.yaml
```

## Configuration

ACARS processor uses YAML configuration with environment variable substitution:

```yaml
ACARSProcessorSettings:
  ACARSHub:
    ACARS:
      Host: "${ACARSHUB_HOST}"
      Port: 15550
  Database:
    Type: sqlite
    SQLiteDatabasePath: ./messages.db
  LogLevel: info

Annotators:
  ACARS:
    Enabled: true
  ADSBExchange:
    Enabled: true
    APIKey: "${ADSB_API_KEY}"

Filters:
  Generic:
    HasText: true
    Emergency: true

Receivers:
  DiscordWebhook:
    Enabled: true
    URL: "${DISCORD_WEBHOOK_URL}"
```

**Environment variables**:
```bash
export ACARSHUB_HOST="your-acarshub-host"
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
export ADSB_API_KEY="your-adsb-exchange-key"
```

**Generate schema for IDE support**:
```bash
./acars-processor -s  # Creates schema.json
```

## Usage Examples

### Emergency Monitoring
```yaml
Filters:
  Generic:
    Emergency: true
  OpenAI:
    Enabled: true
    UserPrompt: "Is this an aircraft emergency?"
    Model: "gpt-4"
Receivers:
  DiscordWebhook:
    URL: "${EMERGENCY_WEBHOOK_URL}"
```

### Regional Aircraft Tracking
```yaml
Annotators:
  Tar1090:
    URL: "http://local-tar1090/"
    ReferenceGeolocation: "40.7128,-74.0060"
Filters:
  Generic:
    BelowDistanceNm: 100.0  # Within 100nm
    AboveSignaldBm: -60.0   # Strong signals only
```

### Maintenance Detection
```yaml
Filters:
  Ollama:
    Enabled: true
    Model: "gemma2:9b"
    UserPrompt: "Does this discuss aircraft maintenance or technical issues?"
```

## Documentation

- **[Architecture](docs/ARCHITECTURE.md)** - System design, components, and data flow
- **[API Reference](docs/API.md)** - Interfaces, data structures, and external APIs  
- **[Configuration](docs/CONFIGURATION.md)** - Complete configuration reference and examples
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Production deployment and operations
- **[Development Tools](docs/DEVELOPMENT.md)** - Development workflow and tooling
- **[Utilities](docs/UTILITIES.md)** - Template system and utility functions
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Diagnostic procedures and issue resolution
- **[Contributing](docs/CONTRIBUTING.md)** - Development setup and guidelines

## Support

- **Issues**: [GitHub Issues](https://github.com/tyzbit/acars-processor/issues) for bugs and feature requests
- **Discussions**: [GitHub Discussions](https://github.com/tyzbit/acars-processor/discussions) for questions and support

## License

GPL v3 - see [LICENSE](LICENSE) for details.

---

⚠️ **Security Note**: This software processes aviation communications which may contain sensitive operational information. Ensure appropriate security measures for your deployment environment.
