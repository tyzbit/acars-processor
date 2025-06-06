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

// Main configuration for acars-processor. Have fun!.
type Config struct {
	// Set logging verbosity.
	LogLevel   string           `json:",omitempty" jsonschema:"default=info" default:"info"`
	ACARSHub   ACARSHubConfig   `jsonschema:"example=acars"`
	Annotators AnnotatorsConfig `json:",omitempty"`
	Filters    FiltersConfig    `json:",omitempty"`
	Receivers  ReceiversConfig  `json:",omitempty"`
}

// ACARSHub connection settings.
type ACARSHubConfig struct {
	// ACARS-specific settings when connecting to ACARSHub.
	ACARS ACARSConnectionConfig `json:",omitempty"`
	// VDLM2-specific settings when connecting to ACARSHub.
	VDLM2 VDLM2Connection `json:",omitempty"`
	// Maximum number of requests from ACARSHub to process at once.
	MaxConcurrentRequests int `json:",omitempty"`
}

// ACARSHub settings for ACARS JSON feed.
type ACARSConnectionConfig struct {
	// IP or DNS to your ACARSHub instance serving ACARS JSON data from a particular port.
	Host string `json:"" jsonschema:"default=acarshub" default:"acarshub"`
	// ACARS JSON port.
	Port int `json:"" jsonschema:"default=15550" default:"15550"`
}

// ACARSHub settings for VDLM2 JSON feed.
type VDLM2Connection struct {
	// IP or DNS to your ACARSHub instance serving ACARS JSON data from a particular port.
	Host string `json:"" jsonschema:"default=acarshub" default:"acarshub"`
	// VDLM2 JSON port.
	Port int `json:"" jsonschema:"default=15555" default:"15555"`
}

// Look up geolocation, including distance from a reference point to aircraft, from ADSB-Exchange.
type ADSBExchangeAnnotatorConfig struct {
	// APIKey provided by signing up at ADSB-Exchange.
	APIKey string `json:"" default:"example_key"`
	// Geolocation to use for distance calculations (LAT,LON).
	ReferenceGeolocation string `json:",omitempty" jsonschema:"example=35.6244416\\,139.7753782" default:"35.6244416,139.7753782"`
	// Fields to provide to receivers from this annotator. Any separator will do.
	SelectedFields string `json:",omitempty" jsonschema:"example=adsbOriginGeolocation\\,adsbOriginGeolocationLatitude\\,adsbOriginGeolocationLongitude\\,adsbAircraftGeolocation\\,adsbAircraftLatitude\\,adsbAircraftLongitude\\,adsbAircraftDistanceKm\\,adsbAircraftDistanceMi" default:"adsbOriginGeolocation,adsbOriginGeolocationLatitude,adsbOriginGeolocationLongitude,adsbAircraftGeolocation,adsbAircraftLatitude,adsbAircraftLongitude,adsbAircraftDistanceKm,adsbAircraftDistanceMi"`
}

// Look up geolocation, including distance from a reference point to aircraft, from a tar1090 instance (which can be self-hosted)
type Tar1090AnnotatorConfig struct {
	// URL to your tar1090 instance
	URL string `json:"" jsonschema:"example:http://tar1090/" default:"http://tar1090/"`
	// Geolocation to use for distance calculations (LAT,LON).
	ReferenceGeolocation string `json:",omitempty" jsonschema:"example=35.6244416\\,139.7753782" default:"35.6244416,139.7753782"`
	// Fields to provide to receivers from this annotator. Any separator will do.
	SelectedFields string `json:",omitempty" jsonschema:"example=tar1090ReferenceGeolocation\\,tar1090ReferenceGeolocationLatitude\\,tar1090ReferenceGeolocationLongitude\\,tar1090AircraftEmergency\\,tar1090AircraftGeolocation\\,tar1090AircraftLatitude\\,tar1090AircraftLongitude\\,tar1090AircraftDistanceKm\\,tar1090AircraftDistanceMi\\,tar1090AircraftDistanceNm\\,tar1090AircraftDirectionDegrees\\,tar1090AircraftAltimeterBarometerFeet\\,tar1090AircraftAltimeterGeometricFeet\\,tar1090AircraftAltimeterBarometerRateFeetPerSecond\\,tar1090AircraftOwnerOperator\\,tar1090AircraftFlightNumber\\,tar1090AircraftHexCode\\,tar1090AircraftType\\,tar1090AircraftDescription\\,tar1090AircraftYearOfManufacture\\,tar1090AircraftADSBMessageCount\\,tar1090AircraftRSSIdBm\\,tar1090AircraftNavModes" `
}

// Services that receive raw ACARS/VDLM2 messages and return more information, usually after a lookup or additional processing.
type AnnotatorsConfig struct {
	ACARS        ACARSAnnotatorConfig        `json:",omitempty" jsonschema:"default"`
	VDLM2        VDLM2AnnotatorConfig        `json:",omitempty" jsonschema:"default"`
	ADSBExchange ADSBExchangeAnnotatorConfig `json:",omitempty"`
	Tar1090      Tar1090AnnotatorConfig      `json:",omitempty"`
	Ollama       OllamaAnnotatorConfig       `json:",omitempty"`
}

// ACARS annotator, you probably want this enabled if you are ingesting ACARS messages into ACARSHub.
type ACARSAnnotatorConfig struct {
	// Should ACARS message data from ACARSHub be sent to receivers?
	Enabled bool `json:"" jsonschema:"default=true" default:"true"`
	// Fields to provide to receivers from this annotator. Any separator will do.
	SelectedFields string `json:",omitempty" jsonschema:"example=acarsFrequencyMHz\\,acarsChannel\\,acarsErrorCode\\,acarsSignaldBm\\,acarsTimestamp\\,acarsAppName\\,acarsAppVersion\\,acarsAppProxied\\,acarsAppProxiedBy\\,acarsAppRouterVersion\\,acarsAppRouterUUID\\,acarsStationID\\,acarsASSStatus\\,acarsMode\\,acarsLabel\\,acarsBlockID\\,acarsAcknowledge\\,acarsAircraftTailCode\\,acarsMessageFrom\\,acarsMessageText\\,acarsMessageNumber\\,acarsFlightNumber\\,acarsExtraURL\\,acarsExtraPhotos" default:"acarsFrequencyMHz,acarsChannel,acarsErrorCode,acarsSignaldBm,acarsTimestamp,acarsAppName,acarsAppVersion,acarsAppProxied,acarsAppProxiedBy,acarsAppRouterVersion,acarsAppRouterUUID,acarsStationID,acarsASSStatus,acarsMode,acarsLabel,acarsBlockID,acarsAcknowledge,acarsAircraftTailCode,acarsMessageFrom,acarsMessageText,acarsMessageNumber,acarsFlightNumber,acarsExtraURL,acarsExtraPhotos"`
}

// VDLM2 annotator, you probably want this enabled if you are ingesting VDLM2 messages into ACARSHub.
type VDLM2AnnotatorConfig struct {
	// Should VDLM2 message data from ACARSHub be sent to receivers?
	Enabled bool `json:"" jsonschema:"default=true" default:"true"`
	// Fields to provide to receivers from this annotator. Any separator will do.
	SelectedFields string `json:",omitempty" jsonschema:"example=vdlm2AppName\\,vdlm2AppVersion\\,vdlm2AppProxied\\,vdlm2AppProxiedBy\\,vdlm2AppRouterVersion\\,vdlm2AppRouterUUID\\,vdlmCR\\,vdlmDestinationAddress\\,vdlmDestinationType\\,vdlmFrameType\\,vdlmSourceAddress\\,vdlmSourceType\\,vdlmSourceStatus\\,vdlmRSequence\\,vdlmSSequence\\,vdlmPoll\\,vdlm2BurstLengthOctets\\,vdlm2FrequencyHz\\,vdlm2Index\\,vdlm2FrequencySkew\\,vdlm2HDRBitsFixed\\,vdlm2NoiseLevel\\,vdlm2OctetsCorrectedByFEC\\,vdlm2SignalLeveldBm\\,vdlm2Station\\,vdlm2Timestamp\\,vdlm2TimestampMicroseconds\\,acarsErrorCode\\,acarsCRCOK\\,acarsMore\\,acarsAircraftTailCode\\,acarsMode\\,acarsLabel\\,acarsBlockID\\,acarsAcknowledge\\,acarsFlightNumber\\,acarsMessageFrom\\,acarsMessageNumber\\,acarsMessageNumberSequence\\,acarsMessageText\\,acarsExtraURL\\,acarsExtraPhotos" default:"vdlm2AppName,vdlm2AppVersion,vdlm2AppProxied,vdlm2AppProxiedBy,vdlm2AppRouterVersion,vdlm2AppRouterUUID,vdlmCR\,vdlmDestinationAddress,vdlmDestinationType,vdlmFrameType,vdlmSourceAddress,vdlmSourceType,vdlmSourceStatus,vdlmRSequence,vdlmSSequence,vdlmPoll,vdlm2BurstLengthOctets,vdlm2FrequencyHz,vdlm2Index,vdlm2FrequencySkew,vdlm2HDRBitsFixed,vdlm2NoiseLevel,vdlm2OctetsCorrectedByFEC,vdlm2SignalLeveldBm,vdlm2Station,vdlm2Timestamp,vdlm2TimestampMicroseconds,acarsErrorCode,acarsCRCOK,acarsMore,acarsAircraftTailCode,acarsMode,acarsLabel,acarsBlockID,acarsAcknowledge,acarsFlightNumber,acarsMessageFrom,acarsMessageNumber,acarsMessageNumberSequence,acarsMessageText,acarsExtraURL,acarsExtraPhotos"`
}

// Use Ollama (which can be self-hosted) to annotate messages, such as to answer custom questions about the message ("Is this message about coffee makers?").
type OllamaAnnotatorConfig struct {
	// Model to use (you need to pull this in Ollama to use it).
	Model string `json:"" jsonschema:"default=llama3.2" default:"llama3.2"`
	// URL to the Ollama instance to use (include protocol and port).
	URL string `json:"" jsonschema:"example=http://ollama-service:11434" default:"http://ollama-service:11434"`
	// Instructions for Ollama for processing messages. More detail produces better results.
	// You can include a question and Ollama will respond yes/no which can also be used to filter the message.
	UserPrompt string `json:"" jsonschema:"example=Is there prose in this message? If present\\, prose will be the last section of a message. Return any prose if found. Surround it with triple backticks." default:"Is there prose in this message? If present, prose will be the last section of a message. Return any prose if found. Surround it with triple backticks."`
	// Override the system prompt (not usually necessary). This instructs Ollama how to behave with user prompts (ex: pretend you are a pirate. all answers must end in "arrr!"). This might make other options less effective.
	SystemPrompt string `json:",omitempty" default:"Answer like a pirate"`
	// Fields to provide to receivers from this annotator. Any separator will do.
	SelectedFields string `json:",omitempty" jsonschema:"example=ollamaProcessedText\\,ollamaEditActions\\,ollamaQuestion" default:"ollamaProcessedText,ollamaEditActions,ollamaQuestion"`
	// If there is a question in the user prompt, this controls whether to use the answer to filter the message.
	FilterWithQuestion bool `json:",omitempty" jsonschema:"example=true" default:"true"`
	// Maximum number of tokens to include in the answer. Lower values restrict response length but too low may clip the valid response short.
	MaxPredictionTokens int `json:",omitempty" jsonschema:"example=512" default:"512"`
	// Maximum number of retries to make against the Ollama URL.
	MaxRetryAttempts int `json:",omitempty" jsonschema:"example=5" default:"5"`
	// How long to wait before retrying the Ollama API.
	MaxRetryDelaySeconds int `json:",omitempty" jsonschema:"example=5" default:"5"`
	// How long to wait until giving up on any request to Ollama.
	Timeout int                   `json:",omitempty" jsonschema:"example=5" default:"5"`
	Options []OllamaOptionsConfig `json:",omitempty"`
}

// Additional models to provide to the model. This is specific to each model, so no defaults are provided
type OllamaOptionsConfig struct {
	// Option name, specific to the model you are using.
	Name string `json:"" jsonschema:"default=example_value" default:"example_value"`
	// Value for this particular option, any value is allowed.
	Value any `json:"" jsonschema:"default=0.1" default:"0.1"`
}

// Filter messages out before being processed.
type FiltersConfig struct {
	Generic GenericFilterConfig `json:",omitempty"`
	ACARS   ACARSFilterConfig   `json:",omitempty"`
	VDLM2   VDLM2FilterConfig   `json:",omitempty"`
	Ollama  OllamaFilterConfig  `json:",omitempty"`
	OpenAI  OpenAIFilterConfig  `json:",omitempty"`
}

// Simple built-in filters.
type GenericFilterConfig struct {
	// Only process messages with text included.
	HasText bool `json:",omitempty" jsonschema:"example=true" default:"true"`
	// Only process messages that have this tail code.
	TailCode string `json:",omitempty" default:"1234"`
	// Only process messages that have this flight number.
	FlightNumber string `json:",omitempty" default:"1234"`
	// Only process messages that have ASS Status.
	ASSStatus string `json:",omitempty" jsonschema:"example=anything" default:"anything"`
	// Only process messages that were received above this signal strength (in dBm).
	AboveSignaldBm float64 `json:",omitempty" jsonschema:"example=-10.0" default:"-9.9"`
	// Only process messages that were received below this signal strength (in dBm).
	BelowSignaldBm float64 `json:",omitempty" jsonschema:"example=-10.0" default:"-9.9"`
	// Only process messages received on this frequency.
	Frequency float64 `json:",omitempty" jsonschema:"example=136.950" default:"136.950"`
	// Only process messages with this station ID.
	StationID string `json:",omitempty" jsonschema:"example=N12346" default:"N12346"`
	// Only process messages that were from a ground-based transmitter - determined by the presence (from aircraft) or lack of (from ground) a flight number.
	FromTower bool `json:",omitempty" jsonschema:"example=true" default:"true"`
	// Only process messages that were from an aircraft - determined by the presence (from aircraft) or lack of (from ground) a flight number.
	FromAircraft bool `json:",omitempty" jsonschema:"example=true" default:"true"`
	// Only process messages that have the "More" flag set.
	More bool `json:",omitempty" jsonschema:"example=true" default:"true"`
	// Only process messages that came from aircraft further than this many nautical miles away (requires ADS-B or tar1090).
	AboveDistanceNm float64 `json:",omitempty" jsonschema:"example=15.5" default:"15.5"`
	// Only process messages that came from aircraft closer than this many nautical miles away (requires ADS-B or tar1090).
	BelowDistanceNm float64 `json:",omitempty" jsonschema:"example=15.5" default:"15.5"`
	// Only process messages that have the "Emergency" flag set.
	Emergency bool `json:",omitempty" jsonschema:"example=true" default:"true"`
	// Only process messages that have at least this many valid dictionary words in a row.
	DictionaryPhraseLengthMinimum int `json:",omitempty" jsonschema:"example=5" default:"5"`
}

// Built-in filters for ACARS messages.
type ACARSFilterConfig struct {
	// Only process ACARS messages that are at least this percent (ex: 0.8 for 80 percent) different than any other message received.
	DuplicateMessageSimilarity float64 `json:",omitempty" jsonschema:"example=0.9" default:"0.9"`
}

// Built-in filters for VDLM2 messages.
type VDLM2FilterConfig struct {
	// Only process VDLM2 messages that are at least this percent (ex: 0.8 for 80 percent) different than any other message received.
	DuplicateMessageSimilarity float64 `json:",omitempty" jsonschema:"example=0.9" default:"0.9"`
}

// Use Ollama (which can be self-hosted) to choose to filter messages based on plain-text criteria.
type OllamaFilterConfig struct {
	// Whether to filter messages where Ollama itself fails. Recommended if your ollama instance sometimes returns errors.
	FilterOnFailure bool `json:",omitempty" jsonschema:"example=true" default:"true"`
	// Model to use (you need to pull this in Ollama to use it).
	Model string `json:"" jsonschema:"default=llama3.2" default:"llama3.2"`
	// URL to the Ollama instance to use (include protocol and port).
	URL string `json:"" jsonschema:"example=http://ollama-service:11434" default:"http://ollama-service:11434"`
	// Instructions for Ollama for processing messages. More detail produces better results.
	// You can include a question and Ollama will respond yes/no which can also be used to filter the message.
	UserPrompt string `json:"" jsonschema:"example=Is there prose in this message?" default:"Is there prose in this message?"`
	// Override the system prompt (not usually necessary). This instructs Ollama how to behave with user prompts (ex: pretend you are a pirate. all answers must end in "arrr!"). This might make other options less effective.
	SystemPrompt string `json:",omitempty" default:"Answer like a pirate"`
	// Maximum number of tokens to include in the answer. Lower values restrict response length but too low may clip the valid response short.
	MaxPredictionTokens int `json:",omitempty" default:"512"`
	// Maximum number of retries to make against the Ollama URL.
	MaxRetryAttempts int `json:",omitempty" jsonschema:"example=5" default:"5"`
	// How long to wait before retrying the Ollama API.
	MaxRetryDelaySeconds int `json:",omitempty" jsonschema:"example=5" default:"5"`
	// How long to wait until giving up on any request to Ollama.
	Timeout int                   `json:",omitempty" jsonschema:"example=5" default:"5"`
	Options []OllamaOptionsConfig `json:",omitempty"`
}

// Use OpenAI to choose to filter messages based on plain-text criteria.
type OpenAIFilterConfig struct {
	APIKey string `json:"" default:"example_key"`
	// Model to use.
	Model string `json:"" jsonschema:"default=gpt-4o" default:"gpt-4o"`
	// Instructions for OpenAI model to use when filtering messages. More detail is better.
	UserPrompt string `json:"" jsonschema:"example=Does this message talk about coffee makers or lavatories (shortand LAV is sometimes used)?" default:"Does this message talk about coffee makers or lavatories (shortand LAV is sometimes used)?"`
	// Override the built-in system prompt to instruct the model on how to behave for requests (not usually necessary).
	SystemPrompt string `json:",omitempty" default:"Answer like a pirate"`
	// How long to wait until giving up on any request to OpenAI.
	Timeout int `json:",omitempty" jsonschema:"example=5" default:"5"`
}

// After messages are filtered and annotated, they're sent to all configured receivers. One example is Discord Webhooks, which allow you to post messages to a channel in Discord.
type ReceiversConfig struct {
	NewRelic       NewRelicReceiverConfig       `json:",omitempty"`
	Webhook        WebHookReceiverConfig        `json:",omitempty"`
	DiscordWebhook DiscordWebhookReceiverConfig `json:",omitempty"`
}

// Send messages to NewRelic as a custom event type.
type NewRelicReceiverConfig struct {
	// API License key to use New Relic.
	APIKey string `json:"" default:"api_key"`
	// Name for the custom event type to create (example if set to "MyCustomACARSEvents": `FROM MyCustomACARSEvents SELECT count(timestamp)`). If not provided, it will be `CustomACARS`.
	CustomEventType string `json:",omitempty" jsonschema:"example=CustomACARS" default:"CustomACARS"`
}

// Generic webhook receiver. Please read README for how to use custom payloads.
type WebHookReceiverConfig struct {
	// URL, including port and params, to the desired webhook.
	URL string `json:"" jsonschema:"example=https://webhook:8443/webhook/?enable_feature=yes" default:"https://webhook:8443/webhook/?enable_feature=yes"`
	// Method when calling webhook (GET,POST,PUT etc).
	Method string `json:"" jsonschema:"default=POST" default:"POST"`
	// Additional headers to send along with the request.
	Headers []WebHookReceiverConfigHeaders `json:",omitempty" default:"{{name:APIKey, value:1234abcdef}}"`
}

// Send messages to a Discord channel using a webhook created from that channel.
type DiscordWebhookReceiverConfig struct {
	// Full URL to the Discord webhook for a channel (edit a channel in the Discord UI for the option to create a webhook).
	URL string `json:"" default:"https://discord.com/api/webhooks/1234321/unique_id1234"`
}

type WebHookReceiverConfigHeaders struct {
	// Header name.
	Name string `json:""` // NOTE: defaults are set in the function in schema.go
	// Header value.
	Value string `json:""` // NOTE: defaults are set in the function in schema.go
}
