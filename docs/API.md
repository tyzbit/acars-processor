# API reference

## Overview

This document provides detailed API reference information for ACARS processor interfaces, data structures, and external integrations.

## Core interfaces

### Receiver interface

All receivers must implement the `Receiver` interface to handle processed message delivery.

```go
type Receiver interface {
    SubmitACARSAnnotations(Annotation) error
    Name() string
}
```

#### Methods

##### SubmitACARSAnnotations(annotation Annotation) error

Delivers processed and annotated message data to the receiver endpoint.

**Parameters**:
- `annotation` (Annotation): Map containing all annotation data from enabled annotators

**Returns**:
- `error`: nil on success, error describing failure

**Error handling**:
- Network timeouts should return retriable errors
- Authentication failures should return non-retriable errors
- Rate limiting should implement exponential backoff

##### Name() string

Returns the human-readable name of the receiver for logging and debugging.

**Returns**:
- `string`: Receiver name (e.g., "Discord Webhook", "New Relic")

### Annotator interfaces

#### ACARS annotator interface

```go
type ACARSAnnotator interface {
    AnnotateACARSMessage(ACARSMessage) Annotation
    Name() string
    SelectFields(Annotation) Annotation
    DefaultFields() []string
}
```

#### VDLM2 annotator interface

```go
type VDLM2Annotator interface {
    AnnotateVDLM2Message(VDLM2Message) Annotation
    Name() string
    SelectFields(Annotation) Annotation
    DefaultFields() []string
}
```

#### Methods

##### AnnotateACARSMessage(message ACARSMessage) Annotation
##### AnnotateVDLM2Message(message VDLM2Message) Annotation

Processes a message and returns annotation data.

**Parameters**:
- `message`: Structured message data from ACARSHub

**Returns**:
- `Annotation`: Map of annotation fields, nil if no annotations produced

**Requirements**:
- All returned field keys must be unique across all annotators
- Must handle network failures gracefully
- Should implement appropriate timeouts and retry logic

##### SelectFields(annotation Annotation) Annotation

Filters annotation results to include only configured fields.

**Parameters**:
- `annotation`: Full annotation map from annotation processing

**Returns**:
- `Annotation`: Filtered map containing only selected fields

##### DefaultFields() []string

Returns the default set of fields this annotator provides.

**Returns**:
- `[]string`: Array of available field names

### Filter interfaces

#### ACARS filter interface

```go
type ACARSFilter interface {
    Filter(ACARSMessage) bool
}
```

#### VDLM2 filter interface

```go
type VDLM2Filter interface {
    Filter(VDLM2Message) bool
}
```

#### Methods

##### Filter(message ACARSMessage) bool
##### Filter(message VDLM2Message) bool

Evaluates whether a message should be filtered (excluded from processing).

**Parameters**:
- `message`: Message data to evaluate

**Returns**:
- `bool`: true if message should be filtered out, false if message should continue processing

**Fail-safe behavior**:
- On error, filters should return false (do not filter) to preserve important messages
- Errors should be logged at appropriate level for debugging

## Data structures

### Core message types

#### ACARSMessage

```go
type ACARSMessage struct {
    gorm.Model
    ProcessingStartedAt time.Time
    
    // ACARS protocol fields
    AircraftTailCode    string    `json:"tail"`
    FlightNumber        string    `json:"flight"`
    MessageText         string    `json:"text"`
    MessageLabel        string    `json:"label"`
    BlockID             string    `json:"block_id"`
    Acknowledge         string    `json:"ack"`
    Mode                string    `json:"mode"`
    
    // Radio transmission fields
    FrequencyMHz        float64   `json:"freq"`
    SignaldBm           float64   `json:"level"`
    ErrorCode           int       `json:"err"`
    Timestamp           time.Time `json:"timestamp"`
    
    // Ground station fields
    StationID           string    `json:"station"`
    
    // Application metadata
    AppName             string    `json:"app_name"`
    AppVersion          string    `json:"app_version"`
}
```

#### VDLM2Message

```go
type VDLM2Message struct {
    gorm.Model
    ProcessingStartedAt time.Time
    
    // VDL Mode 2 protocol fields
    FrequencyHz         int64     `json:"freq"`
    BurstLengthOctets   int       `json:"len"`
    SignalLeveldBm      float64   `json:"level"`
    NoiseLevel          float64   `json:"noise"`
    FrequencySkew       float64   `json:"freq_skew"`
    HDRBitsFixed        int       `json:"hdr_bits_corrected"`
    OctetsCorrectedByFEC int      `json:"octets_corrected_by_fec"`
    Index               int       `json:"idx"`
    Station             string    `json:"station"`
    Timestamp           time.Time `json:"timestamp"`
    TimestampMicroseconds int64   `json:"timestamp_us"`
    
    // Application metadata
    AppName             string    `json:"app_name"`
    AppVersion          string    `json:"app_version"`
    
    // AVLC layer
    AVLC struct {
        CR            string `json:"cr"`
        FrameType     string `json:"frame_type"`
        RSequence     int    `json:"rseq"`
        SSequence     int    `json:"sseq"`
        Poll          bool   `json:"poll"`
        
        Destination struct {
            Address string `json:"addr"`
            Type    string `json:"type"`
        } `json:"dst"`
        
        Source struct {
            Address string `json:"addr"`
            Type    string `json:"type"`
            Status  string `json:"status"`
        } `json:"src"`
    } `json:"avlc"`
}
```

#### Annotation

```go
type Annotation map[string]interface{}
```

Annotations are key-value maps containing enrichment data from annotators. Common annotation fields:

**ACARS annotator fields**:
- `acarsFlightNumber`: Flight identifier
- `acarsAircraftTailCode`: Aircraft registration
- `acarsMessageText`: Message content
- `acarsFrequencyMHz`: Communication frequency
- `acarsSignaldBm`: Signal strength

**VDLM2 annotator fields**:
- `vdlm2FrequencyHz`: VDL frequency
- `vdlm2SignalLeveldBm`: Signal level
- `vdlm2BurstLengthOctets`: Message length

**ADS-B Exchange annotator fields**:
- `adsbAircraftLatitude`: Current latitude
- `adsbAircraftLongitude`: Current longitude
- `adsbAircraftDistanceKm`: Distance from reference point
- `adsbOriginGeolocation`: Departure airport location

**tar1090 annotator fields**:
- `tar1090AircraftType`: Aircraft model
- `tar1090AircraftOwnerOperator`: Airline/operator
- `tar1090AircraftEmergency`: Emergency status
- `tar1090AircraftAltimeterBarometerFeet`: Altitude

**Ollama annotator fields**:
- `OllamaProcessedText`: AI-processed message content
- `OllamaQuestion`: Response to prompt question
- `OllamaEditActions`: Suggested message modifications

## External API integrations

### ADS-B Exchange API

#### Base URL
```
https://adsbexchange.com/api/aircraft/v2/
```

#### Authentication
All requests require API key in header:
```
X-API-Key: your-api-key-here
```

#### Get aircraft by registration

**Endpoint**: `GET /registration/{registration}/`

**Parameters**:
- `registration`: Aircraft tail number (normalized, lowercase, no separators)

**Response format**:
```json
{
  "ac": [
    {
      "hex": "a1b2c3",
      "reg": "n123ab",
      "t": "B738",
      "lat": 40.7128,
      "lon": -74.0060,
      "alt_baro": 35000,
      "gs": 450,
      "track": 90,
      "squawk": "1200",
      "emergency": "none",
      "rssi": -45.2
    }
  ],
  "msg": "success",
  "now": 1672531200,
  "total": 1,
  "ctime": 1672531200,
  "ptime": 15
}
```

**Error responses**:
- `401`: Invalid API key
- `429`: Rate limit exceeded
- `404`: Aircraft not found

### tar1090 API

#### Base URL
Configurable, typically: `http://tar1090-host/data/aircraft.json`

#### Authentication
None required for local instances

#### Get all aircraft data

**Endpoint**: `GET /data/aircraft.json`

**Response format**:
```json
{
  "aircraft": [
    {
      "hex": "a1b2c3",
      "r": "N123AB",
      "t": "B738",
      "lat": 40.7128,
      "lon": -74.0060,
      "alt_baro": 35000,
      "gs": 450,
      "track": 90,
      "squawk": "1200",
      "emergency": "none",
      "rssi": -45.2,
      "messages": 1234,
      "seen": 0.5
    }
  ],
  "now": 1672531200,
  "messages": 567890
}
```

### Ollama API

#### Base URL
Configurable, typically: `http://ollama:11434`

#### Authentication
None required for local instances

#### Generate completion

**Endpoint**: `POST /api/generate`

**Request format**:
```json
{
  "model": "gemma2:9b",
  "prompt": "Analyze this aviation message: ENGINE FAIL LEFT",
  "system": "You are an aviation expert. Respond with JSON only.",
  "format": {
    "type": "object",
    "properties": {
      "processed_text": {"type": "string"},
      "question": {"type": "boolean"},
      "edit_actions": {"type": "array"}
    }
  },
  "stream": false,
  "options": {
    "num_predict": 512,
    "temperature": 0.1
  }
}
```

**Response format**:
```json
{
  "model": "gemma2:9b",
  "created_at": "2025-07-02T10:30:00Z",
  "response": "{\"processed_text\": \"Critical engine failure on left engine\", \"question\": true, \"edit_actions\": []}",
  "done": true
}
```

### OpenAI API

#### Base URL
```
https://api.openai.com/v1/
```

#### Authentication
Bearer token in Authorization header:
```
Authorization: Bearer sk-your-api-key
```

#### Create completion

**Endpoint**: `POST /chat/completions`

**Request format**:
```json
{
  "model": "gpt-4",
  "messages": [
    {
      "role": "system",
      "content": "You are an aviation expert. Respond with YES or NO only."
    },
    {
      "role": "user",
      "content": "Is this an emergency situation: ENGINE FAIL LEFT"
    }
  ],
  "max_tokens": 10,
  "temperature": 0.1
}
```

**Response format**:
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1672531200,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "YES"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 1,
    "total_tokens": 26
  }
}
```

## Webhook payloads

### Discord webhook format

**Endpoint**: Discord webhook URL
**Method**: POST
**Content-Type**: application/json

```json
{
  "content": "**Flight AA1234** - Emergency reported\n```\nENGINE FAIL LEFT\n```",
  "embeds": [
    {
      "title": "Aircraft Communication",
      "color": 16711680,
      "fields": [
        {
          "name": "Flight",
          "value": "AA1234",
          "inline": true
        },
        {
          "name": "Aircraft",
          "value": "N123AB",
          "inline": true
        },
        {
          "name": "Signal",
          "value": "-45 dBm",
          "inline": true
        }
      ],
      "timestamp": "2025-07-02T10:30:00Z"
    }
  ]
}
```

### New Relic custom event format

**Endpoint**: `https://insights-collector.newrelic.com/v1/accounts/{account}/events`
**Method**: POST
**Headers**:
- `Content-Type: application/json`
- `X-Insert-Key: your-insert-key`

```json
{
  "eventType": "AircraftCommunications",
  "timestamp": 1672531200,
  "flight_number": "AA1234",
  "aircraft_registration": "N123AB",
  "message_text": "ENGINE FAIL LEFT",
  "frequency_mhz": 131.725,
  "signal_dbm": -45.2,
  "emergency": true,
  "aircraft_type": "B738",
  "distance_km": 125.5
}
```

### Custom webhook template variables

When using custom webhooks with `receiver_webhook.tpl`, the following variables are available:

#### ACARS message variables
- `{{ .acarsFlightNumber }}`
- `{{ .acarsAircraftTailCode }}`
- `{{ .acarsMessageText }}`
- `{{ .acarsFrequencyMHz }}`
- `{{ .acarsSignaldBm }}`
- `{{ .acarsTimestamp }}`
- `{{ .acarsLabel }}`
- `{{ .acarsStationID }}`

#### VDLM2 message variables
- `{{ .vdlm2FrequencyHz }}`
- `{{ .vdlm2SignalLeveldBm }}`
- `{{ .vdlm2BurstLengthOctets }}`
- `{{ .vdlm2Station }}`
- `{{ .vdlm2Timestamp }}`

#### ADS-B Exchange variables
- `{{ .adsbAircraftLatitude }}`
- `{{ .adsbAircraftLongitude }}`
- `{{ .adsbAircraftDistanceKm }}`
- `{{ .adsbOriginGeolocation }}`

#### tar1090 variables
- `{{ .tar1090AircraftType }}`
- `{{ .tar1090AircraftOwnerOperator }}`
- `{{ .tar1090AircraftEmergency }}`
- `{{ .tar1090AircraftAltimeterBarometerFeet }}`
- `{{ .tar1090AircraftDistanceKm }}`

#### Ollama variables
- `{{ .OllamaProcessedText }}`
- `{{ .OllamaQuestion }}`
- `{{ .OllamaEditActions }}`

#### Template functions

**Conditional rendering**:
```go
{{- if .tar1090AircraftEmergency }}
"emergency": true,
{{- end }}
```

**Number formatting**:
```go
"latitude": {{ printf "%.6f" .adsbAircraftLatitude }},
"signal": {{ printf "%.1f" .acarsSignaldBm }}
```

**String manipulation**:
```go
"message": "{{ .acarsMessageText | html }}",
"flight": "{{ .acarsFlightNumber | upper }}"
```

## Error codes and responses

### Common error patterns

#### Retriable errors
- Network timeouts: Implement exponential backoff
- Rate limiting (429): Wait for retry-after period
- Server errors (5xx): Retry with backoff

#### Non-retriable errors
- Authentication failures (401/403): Fix configuration
- Client errors (400): Fix request format
- Not found (404): Skip processing for this message

#### Error response format

Standard error responses include:
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Request rate limit exceeded",
    "retry_after": 60
  }
}
```

### Logging format

All API interactions are logged with structured data:

```
INFO[2025-07-02T10:30:00Z] API call successful component=adsb_exchange duration=150ms status=200
WARN[2025-07-02T10:30:00Z] API rate limit exceeded component=openai retry_after=60s
ERROR[2025-07-02T10:30:00Z] API authentication failed component=newrelic error="invalid API key"
```

## Performance considerations

### Rate limiting

**ADS-B Exchange**:
- 100 requests per minute for free tier
- 1000 requests per minute for paid tier
- Implement request queuing for high-volume scenarios

**OpenAI**:
- Model-specific rate limits (e.g., 3 RPM for GPT-4)
- Token-based billing affects cost optimization
- Implement intelligent caching for repeated queries

**Ollama**:
- Local processing eliminates external rate limits
- Memory and CPU bound by model size and hardware
- Concurrent request limits based on available resources

### Caching strategies

**Aircraft data caching**:
- Cache ADS-B Exchange responses for 30 seconds
- Cache tar1090 responses for 10 seconds
- Implement LRU eviction for memory management

**AI response caching**:
- Cache Ollama responses for identical message content
- Cache OpenAI responses with content hashing
- Configurable TTL based on use case requirements

### Optimization techniques

**Batch processing**:
- Group multiple messages for external API calls where possible
- Implement request coalescing for duplicate aircraft lookups
- Use connection pooling for HTTP clients

**Memory optimization**:
- Stream processing to avoid loading all messages in memory
- Configurable queue depths to balance latency and memory usage
- Garbage collection tuning for high-throughput scenarios
