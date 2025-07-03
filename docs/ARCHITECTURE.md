# Architecture documentation

## Overview

ACARS processor is a Go message processing pipeline that handles ACARS and VDL Mode 2 messages in real-time. It uses a modular architecture with configurable filters, annotators, and receivers.

## System architecture

### High-level architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   ACARSHub      │    │  ACARS Processor │    │   Receivers     │
│                 │    │                  │    │                 │
│ ┌─────────────┐ │    │ ┌──────────────┐ │    │ ┌─────────────┐ │
│ │ ACARS Port  │ │    │ |   Ingestion  │─┼────┼─│   Discord   │ │
│ │   :15550    │ │    │ │              │ │    │ │   Webhook   │ │
│ └─────────────┘ │    │ └──────┬───────┘ │    │ └─────────────┘ │
│                ─┼────┼─       │         │    │                 │
│ ┌─────────────┐ │    │ ┌──────▼───────┐ │    │ ┌─────────────┐ │
│ │ VDLM2 Port  │ │    │ │   Filters    │ │    │ │  New Relic  │ │
│ │   :15555    │ │    │ │              │ │    │ │             │ │
│ └─────────────┘ │    │ └──────┬───────┘ │    │ └─────────────┘ │
└─────────────────┘    │        │         │    │                 │
                       │ ┌──────▼───────┐ │    │ ┌─────────────┐ │
                       │ │  Annotators  │─┼────┼─│   Custom    │ │
                       │ │              │ │    │ │   Webhook   │ │
                       │ └──────┬───────┘ │    │ └─────────────┘ │
                       │        │         │    └─────────────────┘
                       │ ┌──────▼───────┐ │
                       │ │   Database   │ │
                       │ │  (SQLite/    │ │
                       │ │   MySQL)     │ │
                       │ └──────────────┘ │
                       └──────────────────┘
```

### Core components

#### 1. Message ingestion subsystem

**Location**: `acarshub.go`  
**Purpose**: Establishes TCP connections to ACARSHub for real-time message streaming

Key functions:
- `SubscribeToACARSHub()` - Connection setup
- `ReadACARSHubACARSMessages()` - ACARS message stream handling  
- `ReadACARSHubVDLM2Messages()` - VDLM2 message stream handling
- `HandleACARSJSONMessages()` - ACARS message processing
- `HandleVDLM2JSONMessages()` - VDLM2 message processing

**Design patterns**:
- **Producer-Consumer**: Message readers produce events, handlers consume from queues
- **Fan-Out**: Single connection distributes to multiple concurrent processors
- **Circuit Breaker**: Automatic reconnection on connection failure

#### 2. Filtering subsystem

**Location**: `filters.go`, `filter_*.go`
**Purpose**: Implements configurable message filtering to reduce noise and focus on relevant communications

**Filter types**:

1. **Generic filters** (`filter_*_criteria.go`):
   - Signal strength thresholds
   - Geographic distance filtering
   - Flight number and tail code matching
   - Emergency status detection
   - Message content validation

2. **Protocol-specific filters**:
   - **ACARS**: Message similarity detection, duplicate removal
   - **VDLM2**: Frame type filtering, sequence validation

3. **AI-powered filters**:
   - **Ollama filter** (`filter_ollama.go`): Local LLM-based classification
   - **OpenAI filter** (`filter_openai.go`): Cloud GPT-based classification

**Fail-safe design**: All filters implement fail-closed behavior - if a filter encounters an error, it defaults to not filtering the message, ensuring critical communications are preserved.

#### 3. Annotation subsystem

**Location**: `annotators.go`, `annotator_*.go`
**Purpose**: Enriches messages with additional contextual data from external sources

**Annotator types**:

1. **Protocol annotators**:
   - **ACARS annotator** (`annotator_acars.go`): Extracts structured data from ACARS messages
   - **VDLM2 annotator** (`annotator_vdlm2.go`): Extracts structured data from VDLM2 messages

2. **External data annotators**:
   - **ADS-B Exchange annotator** (`annotator_adsb_exchange.go`): Aircraft position and tracking data
   - **tar1090 annotator** (`annotator_tar1090.go`): Local ADS-B integration
   - **Ollama annotator** (`annotator_ollama.go`): AI-powered message analysis

**Key design principles**:
- **Unique field keys**: All annotators must use unique field names to prevent conflicts
- **Optional enrichment**: Annotation failures do not block message processing
- **Configurable field selection**: Users can choose which annotation fields to include

#### 4. Receiver subsystem

**Location**: `receivers.go`, `receiver_*.go`
**Purpose**: Delivers processed messages to external systems and APIs

**Receiver types**:
- **Discord webhook** (`receiver_discord.go`): Posts formatted messages to Discord channels
- **New Relic** (`receiver_newrelic.go`): Sends telemetry data for monitoring
- **Custom webhook** (`reciever_webhook.go`): Flexible HTTP endpoint integration with templating

**Template system**: Custom webhooks use Go's `text/template` package with `receiver_webhook.tpl` for payload customization.

#### 5. Data persistence layer

**Location**: `db.go`
**Purpose**: Provides message storage and retrieval capabilities

**Supported databases**:
- **SQLite**: Default for development and single-instance deployments
- **MySQL/MariaDB**: Recommended for production and multi-instance deployments

**Schema management**: Uses GORM for automatic schema migration and ORM capabilities.

## Message processing pipeline

### Pipeline stages

1. **Connection management**
   - Establishes TCP connections to ACARSHub
   - Implements automatic reconnection with exponential backoff
   - Maintains separate connections for ACARS and VDLM2 streams

2. **Message ingestion**
   - Decodes JSON messages from TCP streams
   - Validates message structure and format
   - Stores messages in database with processing timestamps

3. **Queue management**
   - Uses buffered channels for message queuing
   - Implements configurable concurrent processing
   - Provides backpressure handling for high-volume scenarios

4. **Filtering pipeline**
   - Applies filters in configured order
   - Short-circuits on first filter match (if configured to filter)
   - Logs filter decisions for debugging and monitoring

5. **Annotation pipeline**
   - Executes all enabled annotators in parallel
   - Merges annotation results into single annotation map
   - Handles annotator failures gracefully without blocking pipeline

6. **Delivery pipeline**
   - Sends annotated messages to all configured receivers
   - Implements retry logic for transient failures
   - Logs delivery success/failure for monitoring

### Concurrency model

**Connection handling**:
- Each ACARSHub connection runs in its own goroutine
- Configurable number of message processing workers per connection type
- Channel-based communication between ingestion and processing

**Message processing**:
- Configurable concurrency level via `MaxConcurrentRequests`
- Each message processed by single goroutine to maintain order
- Annotators can execute API calls concurrently within message processing

**Error handling**:
- Panic recovery at goroutine boundaries
- Structured logging for debugging and monitoring
- Graceful degradation on component failures

## Configuration system

### Configuration structure

**Main configuration** (`config.go`):
```go
type Config struct {
    ACARSProcessorSettings ACARSProcessorSettings
    Annotators            AnnotatorsConfig
    Filters               FiltersConfig
    Receivers             ReceiversConfig
}
```

**Schema generation** (`schema.go`):
- Automatic JSON schema generation from Go structs
- IDE autocomplete support through schema validation
- Example configuration generation with defaults

**Environment variable substitution**:
- Uses `gomodules.xyz/envsubst` for variable replacement
- Supports `${VARIABLE}` syntax in configuration values
- Enables secure secret management in containerized deployments

### Plugin architecture

**Interface definitions** (`types.go`):
```go
type Receiver interface {
    SubmitACARSAnnotations(Annotation) error
    Name() string
}

type ACARSFilter interface {
    Filter(ACARSMessage) bool
}
```

**Registration pattern**:
- Annotators implement message-type-specific interfaces
- Receivers implement common `Receiver` interface
- Filters implement message-type-specific filter interfaces

**Field selection**:
- Each annotator provides `DefaultFields()` method
- Configuration allows field selection per annotator
- Field selection reduces payload size and improves performance

## External integrations

### ACARSHub integration

**Protocol**: TCP with JSON message streaming
**Message format**: Structured JSON with nested objects for different protocol layers
**Connection management**: Persistent connections with automatic reconnection

### AI service integrations

**Ollama integration**:
- Uses official Ollama Go client library
- Supports model-specific configuration options
- Implements retry logic with exponential backoff
- JSON response format enforcement for structured output

**OpenAI integration**:
- Uses official OpenAI Go client library
- Supports multiple model types (GPT-3.5, GPT-4, etc.)
- Implements rate limiting and timeout handling
- Structured prompt engineering for consistent results

### External data sources

**ADS-B Exchange API**:
- RESTful API integration for aircraft position data
- API key authentication
- Rate limiting and error handling
- Geographic distance calculations using Vincenty formula

**tar1090 integration**:
- Custom HTTP client for tar1090 JSON API
- Local network optimization for co-located instances
- Aircraft data caching and deduplication

## Security considerations

### API key management

**Environment variable support**:
- All sensitive configuration supports environment variable substitution
- No default API keys in configuration files
- Secure container secret mounting support

### Network security

**TLS support**:
- HTTPS support for all external API calls
- Certificate validation for secure connections
- Configurable timeout and retry policies

### Data privacy

**Local processing options**:
- Ollama enables local AI processing without cloud dependencies
- SQLite enables local data storage without external databases
- Optional external service integration based on configuration

## Performance characteristics

### Throughput

**Message processing capacity**:
- Configurable concurrent workers per message type
- Typical throughput: 100-1000 messages/second depending on annotator configuration
- Bottlenecks typically in external API calls rather than core processing

### Memory usage

**Baseline requirements**:
- ~50MB for core application
- +8GB for local Ollama models (gemma2:9b)
- Variable based on message queue depth and annotation caching

### Scalability

**Horizontal scaling**:
- Stateless application design enables multiple instances
- Database can be shared across instances (MySQL recommended)
- Load balancing possible at ACARSHub level

**Vertical scaling**:
- CPU-bound workloads benefit from multiple cores
- Memory requirements scale with concurrent processing and AI model size
- I/O-bound workloads benefit from SSD storage for database

## Error handling and resilience

### Circuit breaker patterns

**Connection management**:
- Automatic reconnection to ACARSHub on connection loss
- Exponential backoff for failed connections
- Health check logging for monitoring

**External service integration**:
- Timeout configuration for all external API calls
- Retry logic with configurable attempts and delays
- Graceful degradation when services unavailable

### Logging and observability

**Structured logging**:
- Configurable log levels (error, warn, info, debug)
- Color-coded output for development
- JSON output option for production log aggregation

**Monitoring integration**:
- New Relic integration for application performance monitoring
- Custom metrics and events for message processing
- Health check endpoints for container orchestration

### Failure recovery

**Database resilience**:
- Automatic schema migration on startup
- Transaction support for message processing
- Backup and recovery through standard database tools

**Message durability**:
- Optional message persistence to prevent data loss
- Queue depth monitoring and alerting
- Replay capability for reprocessing scenarios
