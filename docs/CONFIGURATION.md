# Configuration reference

## Overview

ACARS processor uses a comprehensive YAML-based configuration system with automatic JSON schema generation and IDE support. The configuration supports environment variable substitution and provides extensive validation to ensure proper application startup.

## Configuration structure

### Schema generation

The application automatically generates configuration schemas using Go struct reflection:

**Generate schema and example configuration**:
```bash
./acars-processor -s
```

This creates:
- `schema.json` - JSON schema for IDE autocomplete and validation
- `config_example.yaml` - Complete example configuration with all options

### Environment variable substitution

All configuration values support environment variable substitution using the `${VARIABLE_NAME}` syntax:

```yaml
ACARSProcessorSettings:
  ACARSHub:
    ACARS:
      Host: "${ACARSHUB_HOST}"
      Port: 15550
  
Receivers:
  DiscordWebhook:
    URL: "${DISCORD_WEBHOOK_URL}"
```

**Environment variable handling**:
- Variables are resolved at application startup
- Missing variables result in empty strings (not startup failure)
- Use quotes around values containing variables to prevent YAML parsing issues

## Core configuration sections

### ACARSProcessorSettings

Central application configuration including database, logging, and ACARSHub connectivity.

```yaml
ACARSProcessorSettings:
  # Visual output configuration
  ColorOutput: true                    # Enable colored console output
  LogLevel: info                       # Logging verbosity: error, warn, info, debug
  LogHideTimestamps: false             # Hide timestamps in log output
  
  # Database configuration
  Database:
    Enabled: true                      # Enable database storage
    Type: sqlite                       # Database type: sqlite, mariadb
    SQLiteDatabasePath: ./messages.db  # SQLite database file path
    ConnectionString: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
  
  # ACARSHub connection settings
  ACARSHub:
    ACARS:
      Host: acarshub                   # ACARSHub hostname
      Port: 15550                      # ACARS message port
    VDLM2:
      Host: acarshub                   # ACARSHub hostname  
      Port: 15555                      # VDLM2 message port
    MaxConcurrentRequests: 10          # Concurrent message processing limit
```

### Annotators configuration

Services that enrich message data with additional information from external sources.

#### ACARS protocol annotator

Extracts structured data from ACARS messages:

```yaml
Annotators:
  ACARS:
    Enabled: true
    SelectedFields:
      - acarsFlightNumber              # Flight identifier
      - acarsAircraftTailCode          # Aircraft registration
      - acarsMessageText               # Message content
      - acarsFrequencyMHz              # Communication frequency
      - acarsSignaldBm                 # Signal strength
      - acarsTimestamp                 # Message timestamp
      - acarsLabel                     # ACARS message label
      - acarsBlockID                   # Block identifier
      - acarsMode                      # Transmission mode
      - acarsStationID                 # Ground station
```

#### VDLM2 protocol annotator

Processes VDL Mode 2 specific data:

```yaml
  VDLM2:
    Enabled: true
    SelectedFields:
      - acarsFlightNumber              # Flight identifier
      - acarsAircraftTailCode          # Aircraft registration
      - vdlm2FrequencyHz               # VDL frequency in Hz
      - vdlm2SignalLeveldBm            # Signal level
      - vdlm2BurstLengthOctets         # Message length
      - vdlm2Station                   # VDL ground station
      - vdlmFrameType                  # Frame type
      - vdlmSourceAddress              # Source address
      - vdlmDestinationAddress         # Destination address
```

#### External data annotators

**ADS-B Exchange integration**:
```yaml
  ADSBExchange:
    Enabled: true
    APIKey: "${ADSB_API_KEY}"                    # API key from ADS-B Exchange
    ReferenceGeolocation: "40.7128,-74.0060"    # Reference point (lat,lon)
    SelectedFields:
      - adsbAircraftLatitude                     # Current aircraft position
      - adsbAircraftLongitude
      - adsbAircraftDistanceKm                   # Distance from reference
      - adsbOriginGeolocation                    # Departure airport
```

**tar1090 local ADS-B**:
```yaml
  Tar1090:
    Enabled: true
    URL: "http://tar1090.local/"                # tar1090 instance URL
    ReferenceGeolocation: "40.7128,-74.0060"    # Reference point (lat,lon)
    SelectedFields:
      - tar1090AircraftType                      # Aircraft model
      - tar1090AircraftOwnerOperator             # Airline/operator
      - tar1090AircraftEmergency                 # Emergency status
      - tar1090AircraftAltimeterBarometerFeet    # Altitude
      - tar1090AircraftDirectionDegrees          # Heading
      - tar1090AircraftHexCode                   # ICAO hex code
```

**Ollama AI processing**:
```yaml
  Ollama:
    Enabled: true
    URL: "http://ollama:11434"                   # Ollama server URL
    Model: "gemma2:9b"                           # Model name
    SystemPrompt: "Analyze aviation messages"    # AI system instructions
    UserPrompt: "What is significant about this message?"  # Analysis question
    MaxRetryAttempts: 3                          # Retry attempts on failure
    MaxRetryDelaySeconds: 5                      # Delay between retries
    Timeout: 10                                  # Request timeout
    Options:                                     # Model-specific options
      - Name: temperature
        Value: 0.1
      - Name: num_predict
        Value: 100
```

### Filters configuration

Message filtering to reduce noise and focus on relevant communications.

#### Generic filters

Apply to both ACARS and VDLM2 messages:

```yaml
Filters:
  Generic:
    HasText: true                       # Require text content
    Emergency: true                     # Emergency flag set
    FlightNumber: "AA1234"              # Specific flight pattern
    TailCode: "N123AB"                  # Aircraft registration
    StationID: "KORD"                   # Ground station ID
    AboveSignaldBm: -50.0               # Minimum signal strength
    BelowSignaldBm: -20.0               # Maximum signal strength
    AboveDistanceNm: 100.0              # Minimum distance (nautical miles)
    BelowDistanceNm: 200.0              # Maximum distance
    Frequency: 136.95                   # Specific frequency
    FromAircraft: true                  # From aircraft
    FromTower: true                     # From control tower
    More: true                          # Has "More" flag
    DictionaryPhraseLengthMinimum: 5    # Minimum dictionary words
```

#### Protocol-specific filters

**ACARS message filtering**:
```yaml
  ACARS:
    Enabled: true
    DuplicateMessageSimilarity: 0.9     # Filter duplicates (90% similarity)
```

**VDLM2 message filtering**:
```yaml
  VDLM2:
    Enabled: true
    DuplicateMessageSimilarity: 0.9     # Filter duplicates (90% similarity)
```

#### AI-powered filters

**OpenAI classification**:
```yaml
  OpenAI:
    Enabled: true
    APIKey: "${OPENAI_API_KEY}"
    Model: "gpt-4"
    SystemPrompt: "You are an aviation expert. Respond only YES or NO."
    UserPrompt: "Is this message about emergencies or significant events?"
    Timeout: 15
```

**Ollama local AI filtering**:
```yaml
  Ollama:
    Enabled: true
    URL: "http://ollama:11434"
    Model: "gemma2:9b"
    SystemPrompt: "Classify aviation messages as significant or routine."
    UserPrompt: "Is this an emergency or maintenance issue? YES/NO"
    MaxPredictionTokens: 50
    FilterOnFailure: false              # Don't filter if AI fails
    Timeout: 10
```

### Receivers configuration

Output destinations for processed messages.

#### Discord webhook

```yaml
Receivers:
  DiscordWebhook:
    Enabled: true
    URL: "${DISCORD_WEBHOOK_URL}"       # Discord webhook URL
    FormatText: true                    # Apply Discord markdown formatting
    RequiredFields:                     # Only send messages with these fields
      - acarsMessageText
      - acarsFlightNumber
```

#### New Relic telemetry

```yaml
  NewRelic:
    Enabled: true
    APIKey: "${NEWRELIC_API_KEY}"
    CustomEventType: "AircraftCommunications"
```

#### Custom webhook

```yaml
  Webhook:
    Enabled: true
    URL: "https://api.example.com/acars"
    Method: POST                        # HTTP method
    Headers:                            # Custom headers
      - Name: "Authorization"
        Value: "Bearer ${API_TOKEN}"
      - Name: "Content-Type"
        Value: "application/json"
```

## Configuration validation

### Schema validation

The application generates JSON schema for configuration validation:

**Manual validation**:
```bash
# Generate schema
./acars-processor -s

# Validate configuration (requires jsonschema tool)
jsonschema -i config.yaml schema.json
```

**IDE integration**:
The generated `schema.json` provides:
- Real-time validation in VS Code, IntelliJ, and other editors
- Autocomplete for configuration options
- Inline documentation for all settings
- Type checking and format validation

### Error handling

**Configuration loading errors**:
- Invalid YAML syntax causes immediate startup failure
- Missing environment variables result in empty string values
- Invalid enum values (log levels, database types) cause startup failure
- Network connectivity issues to external services are logged as warnings

**Runtime validation**:
- API key validation occurs during first API call
- Database connectivity is verified at startup
- ACARSHub connection is established with automatic retry logic

## Best practices

### Security considerations

1. **Environment variables for secrets**:
   ```bash
   # Good - secrets in environment
   export OPENAI_API_KEY="sk-your-secret-key"
   export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
   
   # Bad - secrets in config file
   # APIKey: "sk-your-secret-key"  # Never do this
   ```

2. **File permissions**:
   ```bash
   # Restrict config file access
   chmod 600 config.yaml
   ```

3. **Database security**:
   ```yaml
   Database:
     ConnectionString: "${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:3306)/${DB_NAME}?charset=utf8mb4&parseTime=True&loc=Local&tls=true"
   ```

### Performance optimization

1. **Concurrent request tuning**:
   ```yaml
   ACARSProcessorSettings:
     ACARSHub:
       MaxConcurrentRequests: 20  # Tune based on CPU cores and network capacity
   ```

2. **AI service timeouts**:
   ```yaml
   Filters:
     OpenAI:
       Timeout: 10              # Balance between reliability and latency
     Ollama:
       Timeout: 5               # Local processing should be faster
   ```

3. **Database optimization**:
   ```yaml
   Database:
     ConnectionString: "user:pass@tcp(host:3306)/db?maxOpenConns=25&maxIdleConns=5"
   ```

### Development configuration

**Development settings**:
```yaml
ACARSProcessorSettings:
  LogLevel: debug                # Detailed logging for development
  ColorOutput: true              # Enhanced console output
  Database:
    Type: sqlite                 # Lightweight development database
    SQLiteDatabasePath: ./dev_messages.db
```

**Production settings**:
```yaml
ACARSProcessorSettings:
  LogLevel: info                 # Reduced logging for production
  ColorOutput: false             # Plain output for log aggregation
  Database:
    Type: mysql                  # Production database
    ConnectionString: "${DATABASE_URL}"
```

This configuration system provides comprehensive control over all aspects of ACARS processor operation while maintaining security best practices and development workflow efficiency.
