# ACARS Processor

A Go app/daemon that processes ACARS and VDL Mode 2 messages in real-time. Connects to ACARSHub, filters and enriches messages with (optional) AI analysis, then delivers results to Discord or custom endpoints. 

## Features

- Real-time ACARS/VDLM2 message processing from ACARSHub
- **Optional** AI-powered filtering (OpenAI, local Ollama) and configurable criteria 
- Data enrichment from ADS-B Exchange, tar1090, and custom annotations (ADSBx API required for enrichment -- works just fine without it) 
- Output to Discord webhooks, New Relic, and custom endpoints
- Docker support with SQLite and MariaDB database options

## Quick start

### Docker

```bash
# Get configuration template
curl -o config.yaml https://raw.githubusercontent.com/tyzbit/acars-processor/main/config_example.yaml

# Edit configuration (Nano is what is used, but feel free to use whatever edit you like obviously) 
nano config.yaml

# Build and run
docker build -t acars-processor .
docker run -d --name acars-processor \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/data:/app/data \
  acars-processor

# View logs
docker logs -f acars-processor
```

### Binary

```bash
# Download binary and config
wget https://github.com/tyzbit/acars-processor/releases/latest/download/acars-processor-linux-amd64
chmod +x acars-processor-linux-amd64
wget https://raw.githubusercontent.com/tyzbit/acars-processor/main/config_example.yaml -O config.yaml

# Configure and run
nano config.yaml
./acars-processor-linux-amd64 -c config.yaml
```

## Configuration

ACARS processor uses .yaml configuration with environment variable substitution:

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

Set environment variables:
```bash
export ACARSHUB_HOST="your-acarshub-host"
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
export ADSB_API_KEY="your-adsb-exchange-key"
```

Generate schema for IDE support:
```bash
./acars-processor -s  # Creates schema.json
```

## Examples

Emergency monitoring:
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

Regional tracking:
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

Maintenance detection:
```yaml
Filters:
  Ollama:
    Enabled: true
    Model: "gemma2:9b"
    UserPrompt: "Does this discuss aircraft maintenance or technical issues?"
```

## Documentation

- [Configuration](docs/CONFIGURATION.md) - Complete configuration reference  
- [Architecture](docs/ARCHITECTURE.md) - System design and data flow
- [Deployment](docs/DEPLOYMENT.md) - Installation and deployment
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues and fixes
- [Contributing](docs/CONTRIBUTING.md) - Bug reports and contributions

## Support

- [GitHub Issues](https://github.com/tyzbit/acars-processor/issues) for bugs and feature requests
- [GitHub Discussions](https://github.com/tyzbit/acars-processor/discussions) for questions and support

## License

GPL v3 - see [LICENSE](LICENSE) for details.

---
