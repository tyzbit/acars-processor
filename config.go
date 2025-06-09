package main

import (
	"strings"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"gomodules.xyz/envsubst"
)

func ConfigureLogging() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   config.ColorOutput,
	})
	loglevel := strings.ToLower(config.LogLevel)
	switch loglevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

func LoadConfig() {
	// Load config file, if present
	cb := ReadFile(configFilePath)

	// Replace environment variables, if present
	envEvalYaml, err := envsubst.EvalEnv(string(cb))
	if err != nil {
		log.Fatalf("there was a problem replacing environment variables: %s", err)
	}

	// Marshal the YAML config into the config struct
	if err := yaml.Unmarshal([]byte(envEvalYaml), &config); err != nil {
		log.Fatalf("unable to load config from %s, err: %s", configFilePath, err)
	}
}

// Main configuration for acars-processor. Have fun!
type Config struct {
	// Set logging verbosity.
	LogLevel string `jsonschema:"default=info" default:"info"`
	// Force whether or not color output is used
	ColorOutput bool `jsonschema:"default=true" default:"true"`
	// ACARSHub connection settings.
	ACARSHub ACARSHubConfig
	// Services that receive raw ACARS/VDLM2 messages and return more information, usually after a lookup or additional processing.
	Annotators AnnotatorsConfig
	// Filter messages out before being processed.
	Filters FiltersConfig
	// After messages are filtered and annotated, they're sent to all configured receivers. One example is Discord Webhooks, which allow you to post messages to a channel in Discord.
	Receivers ReceiversConfig
}

type ACARSHubConfig struct {
	// ACARS-specific settings when connecting to ACARSHub.
	ACARS ACARSConnectionConfig
	// VDLM2-specific settings when connecting to ACARSHub.
	VDLM2 VDLM2Connection
	// Maximum number of requests from ACARSHub to process at once.
	MaxConcurrentRequests int
}

type ACARSJSONConnection struct {
	// IP or DNS to your ACARSHub instance serving JSON data from a particular port.
	Host string `jsonschema:"required,default=acarshub" default:"acarshub"`
}

type ACARSConnectionConfig struct {
	ACARSJSONConnection
	// ACARS JSON port.
	Port int `jsonschema:"required,default=15550" default:"15550"`
}

type VDLM2Connection struct {
	ACARSJSONConnection
	// VDLM2 JSON port.
	Port int `jsonschema:"required,default=15555" default:"15555"`
}

type AnnotatorsConfig struct {
	// ACARS annotator, you probably want this enabled if you are ingesting ACARS messages into ACARSHub.
	ACARS ACARSAnnotatorConfig `jsonschema:"default"`
	// VDLM2 annotator, you probably want this enabled if you are ingesting VDLM2 messages into ACARSHub.
	VDLM2 VDLM2AnnotatorConfig `jsonschema:"default"`
	// Look up geolocation, including distance from a reference point to aircraft, from ADSB-Exchange.
	ADSBExchange ADSBExchangeAnnotatorConfig
	// Look up geolocation, including distance from a reference point to aircraft, from a tar1090 instance (which can be self-hosted)
	Tar1090 Tar1090AnnotatorConfig
	// Use Ollama (which can be self-hosted) to annotate messages, such as to answer custom questions about the message ("Is this message about coffee makers?").
	Ollama OllamaAnnotatorConfig
}

type ModuleCommonConfig struct {
	// Should ACARS message data from ACARSHub be sent to receivers?
	Enabled bool `jsonschema:",default=true" default:"true"`
}

type ACARSAnnotatorConfig struct {
	ModuleCommonConfig
	// Fields to provide to receivers from this annotator.
	SelectedFields []string `jsonschema:"example=acarsFrequencyMHz,example=acarsChannel,example=acarsErrorCode,example=acarsSignaldBm,example=acarsTimestamp,example=acarsAppName,example=acarsAppVersion,example=acarsAppProxied,example=acarsAppProxiedBy,example=acarsAppRouterVersion,example=acarsAppRouterUUID,example=acarsStationID,example=acarsASSStatus,example=acarsMode,example=acarsLabel,example=acarsBlockID,example=acarsAcknowledge,example=acarsAircraftTailCode,example=acarsMessageFrom,example=acarsMessageText,example=acarsMessageNumber,example=acarsFlightNumber,example=acarsExtraURL,example=acarsExtraPhotos" default:"[acarsFrequencyMHz]"`
}

type VDLM2AnnotatorConfig struct {
	ModuleCommonConfig
	// Fields to provide to receivers from this annotator.
	SelectedFields []string `jsonschema:"example=vdlm2AppName,example=vdlm2AppVersion,example=vdlm2AppProxied,example=vdlm2AppProxiedBy,example=vdlm2AppRouterVersion,example=vdlm2AppRouterUUID,example=vdlmCR,example=vdlmDestinationAddress,example=vdlmDestinationType,example=vdlmFrameType,example=vdlmSourceAddress,example=vdlmSourceType,example=vdlmSourceStatus,example=vdlmRSequence,example=vdlmSSequence,example=vdlmPoll,example=vdlm2BurstLengthOctets,example=vdlm2FrequencyHz,example=vdlm2Index,example=vdlm2FrequencySkew,example=vdlm2HDRBitsFixed,example=vdlm2NoiseLevel,example=vdlm2OctetsCorrectedByFEC,example=vdlm2SignalLeveldBm,example=vdlm2Station,example=vdlm2Timestamp,example=vdlm2TimestampMicroseconds,example=acarsErrorCode,example=acarsCRCOK,example=acarsMore,example=acarsAircraftTailCode,example=acarsMode,example=acarsLabel,example=acarsBlockID,example=acarsAcknowledge,example=acarsFlightNumber,example=acarsMessageFrom,example=acarsMessageNumber,example=acarsMessageNumberSequence,example=acarsMessageText,example=acarsExtraURL,example=acarsExtraPhotos" default:"vdlm2AppName"`
}

type ADSBExchangeAnnotatorConfig struct {
	ModuleCommonConfig
	// APIKey provided by signing up at ADSB-Exchange.
	APIKey string `jsonschema:"required" default:"example_key"`
	// Geolocation to use for distance calculations (LAT,LON).
	ReferenceGeolocation string `jsonschema:"example=35.6244416" default:"35.6244416,139.7753782"`
	// Fields to provide to receivers from this annotator.
	SelectedFields []string `jsonschema:"example=adsbOriginGeolocation,example=adsbOriginGeolocationLatitude,example=adsbOriginGeolocationLongitude,example=adsbAircraftGeolocation,example=adsbAircraftLatitude,example=adsbAircraftLongitude,example=adsbAircraftDistanceKm,example=adsbAircraftDistanceMi" default:"adsbOriginGeolocation"`
}

type Tar1090AnnotatorConfig struct {
	ModuleCommonConfig
	// URL to your tar1090 instance
	URL string `jsonschema:"required,example:http://tar1090/" default:"http://tar1090/"`
	// Geolocation to use for distance calculations (LAT,LON).
	ReferenceGeolocation string `jsonschema:"example=35.6244416" default:"35.6244416,139.7753782"`
	// Fields to provide to receivers from this annotator.
	SelectedFields []string `jsonschema:"example=tar1090ReferenceGeolocation,example=tar1090ReferenceGeolocationLatitude,example=tar1090ReferenceGeolocationLongitude,example=tar1090AircraftEmergency,example=tar1090AircraftGeolocation,example=tar1090AircraftLatitude,example=tar1090AircraftLongitude,example=tar1090AircraftDistanceKm,example=tar1090AircraftDistanceMi,example=tar1090AircraftDistanceNm,example=tar1090AircraftDirectionDegrees,example=tar1090AircraftAltimeterBarometerFeet,example=tar1090AircraftAltimeterGeometricFeet,example=tar1090AircraftAltimeterBarometerRateFeetPerSecond,example=tar1090AircraftOwnerOperator,example=tar1090AircraftFlightNumber,example=tar1090AircraftHexCode,example=tar1090AircraftType,example=tar1090AircraftDescription,example=tar1090AircraftYearOfManufacture,example=tar1090AircraftADSBMessageCount,example=tar1090AircraftRSSIdBm,example=tar1090AircraftNavModes" default:"tar1090AircraftDistanceMi"`
}

type OllamaCommonConfig struct {
	// Model to use (you need to pull this in Ollama to use it).
	Model string `jsonschema:"required,default=llama3.2" default:"llama3.2"`
	// URL to the Ollama instance to use (include protocol and port).
	URL string `jsonschema:"required,example=http://ollama-service:11434" default:"http://ollama-service:11434"`
	// Override the system prompt (not usually necessary). This instructs Ollama how to behave with user prompts (ex: pretend you are a pirate. all answers must end in "arrr!"). This might make other options less effective.
	SystemPrompt string `default:"Answer like a pirate"`
	// Maximum number of tokens to include in the answer. Lower values restrict response length but too low may clip the valid response short.
	MaxPredictionTokens int `jsonschema:"example=512" default:"512"`
	// Maximum number of retries to make against the Ollama URL.
	MaxRetryAttempts int `jsonschema:"example=5" default:"5"`
	// How long to wait before retrying the Ollama API.
	MaxRetryDelaySeconds int `jsonschema:"example=5" default:"5"`
	// How long to wait until giving up on any request to Ollama.
	Timeout int `jsonschema:"example=5" default:"5"`
	// Whether to surround the returned message field with backticks.
	Options []OllamaOptionsConfig
}

type OllamaAnnotatorConfig struct {
	ModuleCommonConfig
	OllamaCommonConfig
	// Instructions for Ollama for processing messages. More detail produces better results. You can include a question and Ollama will respond yes/no which can also be used to filter the message.
	UserPrompt string `jsonschema:"required,example=Is there prose in this message? If present\\, prose will be the last section of a message. Return any prose if found."`
	// If there is a question in the user prompt, this controls whether to use the answer to filter the message.
	FilterWithQuestion bool `jsonschema:"example=true" default:"true"`
	// Fields to provide to receivers from this annotator.
	SelectedFields []string `jsonschema:"example=ollamaProcessedText,example=ollamaEditActions,example=ollamaQuestion" default:"ollamaProcessedText"`
}

type OllamaOptionsConfig struct {
	// Option name, specific to the model you are using.
	Name string `jsonschema:"required,default=example_value" default:"example_value"`
	// Value for this particular option, any value is allowed.
	Value any `jsonschema:"required,default=0.1" default:"0.1"`
}

type FiltersConfig struct {
	// Simple built-in filters.
	Generic GenericFilterConfig
	// Built-in filters for ACARS messages.
	ACARS ACARSFilterConfig
	// Built-in filters for VDLM2 messages.
	VDLM2 VDLM2FilterConfig
	// Use Ollama (which can be self-hosted) to choose to filter messages based on plain-text criteria.
	Ollama OllamaFilterConfig
	// Use OpenAI to choose to filter messages based on plain-text criteria.
	OpenAI OpenAIFilterConfig
}

type GenericFilterConfig struct {
	// Only process messages with text included.
	HasText bool `jsonschema:"example=true" default:"true"`
	// Only process messages that have this tail code.
	TailCode string `default:"1234"`
	// Only process messages that have this flight number.
	FlightNumber string `default:"1234"`
	// Only process messages that have ASS Status.
	ASSStatus string `jsonschema:"example=anything" default:"anything"`
	// Only process messages that were received above this signal strength (in dBm).
	AboveSignaldBm float64 `jsonschema:"example=-10.0" default:"-9.9"`
	// Only process messages that were received below this signal strength (in dBm).
	BelowSignaldBm float64 `jsonschema:"example=-10.0" default:"-9.9"`
	// Only process messages received on this frequency.
	Frequency float64 `jsonschema:"example=136.950" default:"136.950"`
	// Only process messages with this station ID.
	StationID string `jsonschema:"example=N12346" default:"N12346"`
	// Only process messages that were from a ground-based transmitter - determined by the presence (from aircraft) or lack of (from ground) a flight number.
	FromTower bool `jsonschema:"example=true" default:"true"`
	// Only process messages that were from an aircraft - determined by the presence (from aircraft) or lack of (from ground) a flight number.
	FromAircraft bool `jsonschema:"example=true" default:"true"`
	// Only process messages that have the "More" flag set.
	More bool `jsonschema:"example=true" default:"true"`
	// Only process messages that came from aircraft further than this many nautical miles away (requires ADS-B or tar1090).
	AboveDistanceNm float64 `jsonschema:"example=15.5" default:"15.5"`
	// Only process messages that came from aircraft closer than this many nautical miles away (requires ADS-B or tar1090).
	BelowDistanceNm float64 `jsonschema:"example=15.5" default:"15.5"`
	// Only process messages that have the "Emergency" flag set.
	Emergency bool `jsonschema:"example=true" default:"true"`
	// Only process messages that have at least this many valid dictionary words in a row.
	DictionaryPhraseLengthMinimum int `jsonschema:"example=5" default:"5"`
}

type ACARSFilterConfig struct {
	ModuleCommonConfig
	// Only process ACARS messages that are at least this percent (ex: 0.8 for 80 percent) different than any other message received.
	DuplicateMessageSimilarity float64 `jsonschema:"example=0.9" default:"0.9"`
}

type VDLM2FilterConfig struct {
	ModuleCommonConfig
	// Only process VDLM2 messages that are at least this percent (ex: 0.8 for 80 percent) different than any other message received.
	DuplicateMessageSimilarity float64 `jsonschema:"example=0.9" default:"0.9"`
}

type OllamaFilterConfig struct {
	ModuleCommonConfig
	OllamaCommonConfig
	// Whether to filter messages where Ollama itself fails. Recommended if your ollama instance sometimes returns errors.
	FilterOnFailure bool `jsonschema:"example=true" default:"true"`
	// Model to use (you need to pull this in Ollama to use it).
	Model string `jsonschema:"required,default=llama3.2" default:"llama3.2"`
	// URL to the Ollama instance to use (include protocol and port).
	URL string `jsonschema:"required,example=http://ollama-service:11434" default:"http://ollama-service:11434"`
	// Instructions for Ollama for processing messages. More detail produces better results.
	UserPrompt   string `jsonschema:"required,example=Is there prose in this message?"` // Override the system prompt (not usually necessary). This instructs Ollama how to behave with user prompts (ex: pretend you are a pirate. all answers must end in "arrr!"). This might make other options less effective.
	SystemPrompt string `default:"Answer like a pirate"`
	// Maximum number of tokens to include in the answer. Lower values restrict response length but too low may clip the valid response short.
	MaxPredictionTokens int `default:"512"`
	// Maximum number of retries to make against the Ollama URL.
	MaxRetryAttempts int `jsonschema:"example=5" default:"5"`
	// How long to wait before retrying the Ollama API.
	MaxRetryDelaySeconds int `jsonschema:"example=5" default:"5"`
	// How long to wait until giving up on any request to Ollama.
	Timeout int `jsonschema:"example=5" default:"5"`
	// Additional options to provide to the model. This is specific to each model, so no defaults are provided
	Options []OllamaOptionsConfig
}

type OpenAIFilterConfig struct {
	ModuleCommonConfig
	APIKey string `jsonschema:"required" default:"example_key"`
	// Model to use.
	Model string `jsonschema:"required,default=gpt-4o" default:"gpt-4o"`
	// Instructions for OpenAI model to use when filtering messages. More detail is better.
	UserPrompt string `jsonschema:"required,example=Does this message talk about coffee makers or lavatories (shortand LAV is sometimes used)?" default:"Does this message talk about coffee makers or lavatories (shortand LAV is sometimes used)?"`
	// Override the built-in system prompt to instruct the model on how to behave for requests (not usually necessary).
	SystemPrompt string `default:"Answer like a pirate"`
	// How long to wait until giving up on any request to OpenAI.
	Timeout int `jsonschema:"example=5" default:"5"`
}

type ReceiversConfig struct {
	// Send messages to NewRelic as a custom event type.
	NewRelic NewRelicReceiverConfig
	// Generic webhook receiver. Please read README for how to use custom payloads.
	Webhook WebHookReceiverConfig
	// Send messages to a Discord channel using a webhook created from that channel.
	DiscordWebhook DiscordWebhookReceiverConfig
}

type NewRelicReceiverConfig struct {
	ModuleCommonConfig
	// API License key to use New Relic.
	APIKey string `jsonschema:"required" default:"api_key"`
	// Name for the custom event type to create (example if set to "MyCustomACARSEvents": `FROM MyCustomACARSEvents SELECT count(timestamp)`). If not provided, it will be `CustomACARS`.
	CustomEventType string `jsonschema:"example=CustomACARS" default:"CustomACARS"`
}

type WebHookReceiverConfig struct {
	ModuleCommonConfig
	// URL, including port and params, to the desired webhook.
	URL string `jsonschema:"required,example=https://webhook:8443/webhook/?enable_feature=yes" default:"https://webhook:8443/webhook/?enable_feature=yes"`
	// Method when calling webhook (GET,POST,PUT etc).
	Method string `jsonschema:"required,default=POST" default:"POST"`
	// Additional headers to send along with the request.
	Headers []WebHookReceiverConfigHeaders `default:"{{name:APIKey, value:1234abcdef}}"`
}

type DiscordWebhookReceiverConfig struct {
	ModuleCommonConfig
	// Full URL to the Discord webhook for a channel (edit a channel in the Discord UI for the option to create a webhook).
	URL string `jsonschema:"required" default:"https://discord.com/api/webhooks/1234321/unique_id1234"`
	// Surround fields with message content with backticks so they are monospaced and stand out.
	FormatText bool `jsonschema:"default=true" default:"true"`
	// Require a specific field to be populated or else no message will be sent
	RequiredFields []string `jsonschema:"example=acarsMessageText" default:"acarsMessageText"`
}

type WebHookReceiverConfigHeaders struct {
	// Header name.
	Name string `jsonschema:"required"` // NOTE: defaults are set in the function in schema.go
	// Header value.
	Value string `jsonschema:"required"` // NOTE: defaults are set in the function in schema.go
}
