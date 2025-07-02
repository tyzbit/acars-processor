# Template system and utilities

## Overview

ACARS processor includes a flexible template system for webhook payload customization and a comprehensive utility library that provides core functionality for string processing, file operations, and data manipulation. This document covers the template system, utility functions, and logging framework.

## Webhook template system

### Template engine

ACARS processor uses Go's `text/template` package to provide flexible webhook payload customization through the `receiver_webhook.tpl` file.

### Default template

**Basic template** (`receiver_webhook.tpl`):
```json
{
  "flight_number": "{{ .acarsFlightNumber }}"
}
```

### Template variables

All annotation fields are available as template variables. The variable names correspond to the field names from enabled annotators:

#### ACARS annotator fields
```json
{
  "timestamp": "{{ .acarsTimestamp }}",
  "flight": "{{ .acarsFlightNumber }}",
  "aircraft": "{{ .acarsAircraftTailCode }}",
  "message": "{{ .acarsMessageText }}",
  "frequency_mhz": {{ .acarsFrequencyMHz }},
  "signal_dbm": {{ .acarsSignaldBm }},
  "label": "{{ .acarsLabel }}",
  "block_id": "{{ .acarsBlockID }}",
  "mode": "{{ .acarsMode }}",
  "station_id": "{{ .acarsStationID }}"
}
```

#### VDLM2 annotator fields
```json
{
  "vdl_frequency_hz": {{ .vdlm2FrequencyHz }},
  "vdl_signal_dbm": {{ .vdlm2SignalLeveldBm }},
  "vdl_burst_length": {{ .vdlm2BurstLengthOctets }},
  "vdl_station": "{{ .vdlm2Station }}",
  "frame_type": "{{ .vdlmFrameType }}",
  "source_address": "{{ .vdlmSourceAddress }}",
  "destination_address": "{{ .vdlmDestinationAddress }}"
}
```

#### ADS-B Exchange fields
```json
{
  "position": {
    "latitude": {{ .adsbAircraftLatitude }},
    "longitude": {{ .adsbAircraftLongitude }},
    "distance_km": {{ .adsbAircraftDistanceKm }},
    "distance_miles": {{ .adsbAircraftDistanceMi }}
  },
  "origin": "{{ .adsbOriginGeolocation }}",
  "geolocation": "{{ .adsbAircraftGeolocation }}"
}
```

#### tar1090 fields
```json
{
  "aircraft_info": {
    "type": "{{ .tar1090AircraftType }}",
    "operator": "{{ .tar1090AircraftOwnerOperator }}",
    "hex_code": "{{ .tar1090AircraftHexCode }}"
  },
  "position": {
    "latitude": {{ .tar1090AircraftLatitude }},
    "longitude": {{ .tar1090AircraftLongitude }},
    "altitude_feet": {{ .tar1090AircraftAltimeterBarometerFeet }},
    "heading_degrees": {{ .tar1090AircraftDirectionDegrees }},
    "distance_km": {{ .tar1090AircraftDistanceKm }}
  },
  "emergency": {{ .tar1090AircraftEmergency }},
  "flight_number": "{{ .tar1090AircraftFlightNumber }}"
}
```

#### Ollama AI processing fields
```json
{
  "ai_analysis": {
    "processed_text": "{{ .OllamaProcessedText }}",
    "question_response": "{{ .OllamaQuestion }}"
  }
}
```

### Advanced template examples

#### Comprehensive aviation data template
```json
{
  "timestamp": "{{ .acarsTimestamp }}",
  "aircraft": {
    "flight_number": "{{ .acarsFlightNumber }}",
    "tail_code": "{{ .acarsAircraftTailCode }}",
    {{- if .tar1090AircraftType }}
    "type": "{{ .tar1090AircraftType }}",
    "operator": "{{ .tar1090AircraftOwnerOperator }}",
    {{- end }}
    "hex_code": "{{ .tar1090AircraftHexCode }}"
  },
  "communication": {
    "message": "{{ .acarsMessageText }}",
    "frequency_mhz": {{ .acarsFrequencyMHz }},
    "signal_dbm": {{ .acarsSignaldBm }},
    "label": "{{ .acarsLabel }}",
    "station": "{{ .acarsStationID }}"
  },
  {{- if .tar1090AircraftLatitude }}
  "position": {
    "latitude": {{ .tar1090AircraftLatitude }},
    "longitude": {{ .tar1090AircraftLongitude }},
    "altitude_feet": {{ .tar1090AircraftAltimeterBarometerFeet }},
    "heading_degrees": {{ .tar1090AircraftDirectionDegrees }},
    "distance_km": {{ .tar1090AircraftDistanceKm }}
  },
  {{- end }}
  {{- if .tar1090AircraftEmergency }}
  "emergency": true,
  {{- end }}
  {{- if .OllamaProcessedText }}
  "ai_analysis": {
    "summary": "{{ .OllamaProcessedText }}",
    "classification": "{{ .OllamaQuestion }}"
  },
  {{- end }}
  "metadata": {
    "processed_by": "acars-processor",
    "webhook_version": "1.0"
  }
}
```

#### Slack integration template
```json
{
  "text": "✈️ Aircraft Communication Alert",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*Flight:* {{ .acarsFlightNumber }}\n*Aircraft:* {{ .acarsAircraftTailCode }}\n*Message:* {{ .acarsMessageText }}"
      }
    },
    {{- if .tar1090AircraftLatitude }}
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*Position:* {{ .tar1090AircraftLatitude }}, {{ .tar1090AircraftLongitude }}\n*Altitude:* {{ .tar1090AircraftAltimeterBarometerFeet }} ft\n*Distance:* {{ .tar1090AircraftDistanceKm }} km"
      }
    },
    {{- end }}
    {
      "type": "context",
      "elements": [
        {
          "type": "mrkdwn",
          "text": "Signal: {{ .acarsSignaldBm }} dBm | Frequency: {{ .acarsFrequencyMHz }} MHz | {{ .acarsTimestamp }}"
        }
      ]
    }
  ]
}
```

### Template functions and conditionals

**Conditional rendering**:
```json
{{- if .tar1090AircraftEmergency }}
"emergency_status": "EMERGENCY",
{{- else }}
"emergency_status": "normal",
{{- end }}
```

**Null value handling**:
```json
"altitude": {{ if .tar1090AircraftAltimeterBarometerFeet }}{{ .tar1090AircraftAltimeterBarometerFeet }}{{ else }}null{{ end }}
```

**String formatting**:
```json
"message_summary": "{{ printf "%.50s" .acarsMessageText }}{{ if gt (len .acarsMessageText) 50 }}...{{ end }}"
```

## Utility functions

### String processing utilities

#### Aircraft registration normalization
```go
func NormalizeAircraftRegistration(reg string) string
```

**Purpose**: Standardizes aircraft registration codes for consistent matching across different data sources.

**Features**:
- Removes dots, spaces, and hyphens
- Converts to lowercase
- Handles international registration format variations

**Usage examples**:
```go
// Various input formats normalized to consistent output
NormalizeAircraftRegistration("N123AB")     // "n123ab"
NormalizeAircraftRegistration("N-123-AB")   // "n123ab" 
NormalizeAircraftRegistration("N.123.AB")   // "n123ab"
NormalizeAircraftRegistration("N 123 AB")   // "n123ab"
```

#### Aircraft vs tower detection
```go
func AircraftOrTower(s string) string
```

**Purpose**: Determines message origin based on flight number presence.

**Logic**:
- Returns "Aircraft" if string contains printable characters (flight number)
- Returns "Tower" if string is empty or contains only non-printable characters

**Usage**:
```go
AircraftOrTower("AA1234")  // "Aircraft"
AircraftOrTower("")        // "Tower"
AircraftOrTower("   ")     // "Tower"
```

#### JSON string sanitization
```go
func SanitizeJSONString(s string) string
```

**Purpose**: Fixes common AI output formatting issues that break JSON parsing.

**Features**:
- Replaces smart quotes with standard quotes
- Normalizes apostrophes
- Handles Unicode quote variations

**Character replacements**:
```go
""" → "\""    // Left double quotation mark
""" → "\""    // Right double quotation mark  
"'" → "'"     // Left single quotation mark
"'" → "'"     // Right single quotation mark
```

#### Text truncation
```go
func Last20Characters(s string) string
```

**Purpose**: Creates consistent short representations of text content.

**Features**:
- Removes newlines and converts to spaces
- Trims leading whitespace
- Returns last 20 characters or full string if shorter

### File operations

#### Safe file reading
```go
func ReadFile(filePath string) []byte
```

**Features**:
- Handles file not found errors gracefully
- Logs errors for debugging
- Returns empty byte slice for missing files
- Exits on permission or I/O errors

#### File writing
```go
func WriteFile(filePath string, contents []byte)
```

**Features**:
- Creates files with 0644 permissions
- Logs errors for debugging
- Handles directory creation automatically

#### Atomic file updates
```go
func UpdateFile(filePath string, contents []byte) (changed bool)
```

**Purpose**: Updates files only when content changes.

**Features**:
- Compares existing content before writing
- Returns true if file was modified
- Useful for avoiding unnecessary file system operations
- Optimizes build and deployment processes

### Data manipulation

#### Map merging
```go
func MergeMaps(m1, m2 map[string]any) map[string]any
```

**Purpose**: Combines annotation data from multiple sources.

**Features**:
- Creates new map without modifying originals
- Later maps override earlier values for duplicate keys
- Handles nested map structures
- Essential for annotation pipeline processing

**Usage in annotation pipeline**:
```go
// Combine annotations from multiple annotators
acarsData := ACARSAnnotatorHandler{}.AnnotateACARSMessage(msg)
adsbData := ADSBAnnotatorHandler{}.AnnotateACARSMessage(msg)
combined := MergeMaps(acarsData, adsbData)
```

### Error handling

#### Retriable error type
```go
type RetriableError struct {
    Err        error
    RetryAfter time.Duration
}
```

**Purpose**: Implements intelligent retry logic for external API failures.

**Features**:
- Wraps underlying errors with retry timing information
- Supports exponential backoff strategies
- Integrates with external API rate limiting
- Enables graceful degradation during service outages

**Usage patterns**:
```go
// API rate limiting response
if rateLimited {
    return &RetriableError{
        Err:        errors.New("API rate limit exceeded"),
        RetryAfter: time.Minute * 5,
    }
}

// Temporary network failure
if networkError {
    return &RetriableError{
        Err:        err,
        RetryAfter: time.Second * 30,
    }
}
```

## Logging framework

### Color-coded logging system

ACARS processor implements a sophisticated color-coded logging system that enhances readability and provides visual distinction between different types of log messages.

#### Core logging functions

```go
func Success(s ...any) string    // Green - successful operations
func Content(s ...any) string    // Magenta - general content
func Note(s ...any) string       // Cyan - informational notes  
func Attention(s ...any) string  // Yellow - warnings and issues
func Aside(s ...any) string      // Gray - verbose/debug information
func Emphasised(s ...any) string // Bold+Italic - critical information
func Custom(c color.Color, s ...any) string // Custom color support
```

#### Color coding standards

**Message type classification**:
- **Success (Green)**: Successful connections, operations completed
- **Content (Magenta)**: General status information, counts, summaries
- **Note (Cyan)**: Non-critical informational messages
- **Attention (Yellow)**: Warnings, issues requiring attention
- **Aside (Gray)**: Debug information, verbose output
- **Emphasised (Bold+Italic)**: Critical alerts, important status changes

#### Configuration-aware color output

```go
func ColorSprintf(c color.Color, n ...any) string
```

**Features**:
- Respects `ColorOutput` configuration setting
- Automatically disables colors for log aggregation systems
- Handles both simple messages and formatted strings
- Includes color reset sequences to prevent bleeding

**Configuration control**:
```yaml
ACARSProcessorSettings:
  ColorOutput: true   # Enable colors for development
  ColorOutput: false  # Disable colors for production/logging
```

#### Usage examples

**Application startup logging**:
```go
log.Info(Success("Connected to ACARSHub successfully"))
log.Info(Content("%d annotators enabled", len(enabledAnnotators)))
log.Info(Note("Configuration loaded from %s", configPath))
log.Warn(Attention("No receivers configured"))
log.Debug(Aside("Debug mode enabled"))
```

**Error and warning patterns**:
```go
// Warning with suggested action
log.Warn(Attention("API rate limit exceeded, retrying in %v", retryAfter))

// Critical error requiring attention  
log.Error(Emphasised("Database connection failed: %v", err))

// Informational success
log.Info(Success("Message processed and sent to %d receivers", receiverCount))
```

**Conditional color output**:
```go
// Colors automatically disabled based on configuration
if config.ACARSProcessorSettings.ColorOutput {
    // Colors enabled for interactive use
} else {
    // Plain text for log aggregation systems
}
```

This template and utility system provides comprehensive support for webhook customization, robust string processing, efficient file operations, and enhanced logging capabilities that improve both development experience and production operational visibility.
