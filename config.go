package main

import (
	"strings"

	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"gomodules.xyz/envsubst"
)

func ConfigureLogging() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:    true,
		ForceColors:      config.ACARSProcessorSettings.ColorOutput,
		DisableTimestamp: config.ACARSProcessorSettings.LogHideTimestamps,
	})
	loglevel := strings.ToLower(config.ACARSProcessorSettings.LogLevel)
	if l, err := log.ParseLevel(loglevel); err == nil {
		log.SetLevel(l)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	if config.ACARSProcessorSettings.ColorOutput {
		color.NoColor = false
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

// Special message format internal to ACARS-Processor. Not ACARS/VDLM2 specific.
// These are the fields and values that are given to receviers.
type APMessage map[string]any

// Main configuration for acars-processor. Have fun!
type Config struct {
	// These control acars-processor itself
	ACARSProcessorSettings ACARSProcessorSettings `jsonschema:"required"`
	// Actions to take on messages in the order they should be taken.
	Steps []ProcessingStep
}

type ACARSProcessorSettings struct {
	// Force whether or not color output is used.
	ColorOutput bool `json:",omitempty" jsonschema:"default=true" default:"true"`
	// Database configuration
	Database ACARSProcessorDatabaseConfig `json:",omitempty"`
	// Set logging verbosity.
	LogLevel string `json:",omitempty" jsonschema:"default=info" default:"info"`
	// Whether to refrain from printing timestamps in logs.
	LogHideTimestamps bool `json:",omitempty" jsonschema:"default=false" default:"false"`
	// ACARSHub connection settings.
	ACARSHub ACARSHubConfig `jsonschema:"required"`
}

type ACARSProcessorDatabaseConfig struct {
	// Whether or not to use a database to save messages.
	Enabled bool `json:",omitempty" jsonschema:"default=false" default:"false"`
	// Type of database to use
	Type string `json:"" jsonschema:"example=sqlite,example=mariadb" default:"sqlite"`
	// Connection string (if using an external database)
	ConnectionString string `json:",omitempty" jsonschema:"example=user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local" default:"user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"`
	// Path to the database file (if using SQLITE). If set to an empty string (""), database will be in-memory only.
	SQLiteDatabasePath string `json:",omitempty" jsonschema:"default=./messages.db" default:"./messages.db"`
}

type ACARSHubConfig struct {
	// ACARS-specific settings when connecting to ACARSHub.
	ACARS ACARSConnectionConfig
	// VDLM2-specific settings when connecting to ACARSHub.
	VDLM2 VDLM2ConnectionConfig
	// Maximum number of requests from ACARSHub to process at once.
	MaxConcurrentRequests int
}

type ACARSJSONConnection struct {
	// IP or DNS to your ACARSHub instance serving JSON data from a particular port.
	Host string `jsonschema:"required,default=acarshub" default:"acarshub"`
}

type ACARSConnectionConfig struct {
	Module
	ACARSJSONConnection
	// ACARS JSON port.
	Port int `jsonschema:"required,default=15550" default:"15550"`
	// Only provide these fields to configured steps.
	SelectedFields []string
}

type VDLM2ConnectionConfig struct {
	Module
	ACARSJSONConnection
	// VDLM2 JSON port.
	Port int `jsonschema:"required,default=15555" default:"15555"`
	// Only provide these fields to configured steps.
	SelectedFields []string
}

// A module is a standard component of a ProcessingStep. Source (internal only
// such as ACARS/VDLM2 feeders), Annotator, Filter, Receiver are all components
// of their ProcessingSteps.
type Module interface {
	// How the module should be named in logs and elsewhere
	Name() string
	// Returns whether or not the Filterer is configured for a given step.
	Configured() bool
	// Used to generate the exhaustive all-option config
	GetDefaultFields() []string
}

// Here we add some extra fields for convenience and to have common
// fields between ACARS and VDLM2 when each has a different measuring unit
// Only ACARS and VDLM2, there probably won't be more sources.
type Source interface {
	Prepare() APMessage
}

// An Annotator provides annotations to received messages. Annotations here just
// means additional fields and values.
type Annotator interface {
	Module
	Annotate(APMessage) (APMessage, error)
}

// Filterers evaluate fields provided by previous steps they are compatible
// with (like "ACARSMessage.AircraftGeolocation"). If the message is not
// filtered, it proceeds to the next step.
type Filterer interface {
	Module
	// The main entrypoint into a Filterer, returns true if the message
	// should be filtered.
	Filter(APMessage) (filter bool, reason string, err error)
}

// Receivers are destinations where annotated and filtered messages are sent.
// An example is a webhook.
type Receiver interface {
	Module
	Send(APMessage) error
}

type ProcessingStep struct {
	// Apply one or more filters in this step
	Filter FilterStep `json:",omitempty"`
	// Add annotations from one or more annotators in this step
	Annotate AnnotateStep `json:",omitempty"`
	// Send the message to one or more receivers in this step
	Send ReceiverStep `json:",omitempty"`
}

type FilterStep struct {
	// Built-in filters
	Builtin BuiltinFilter
	// Use Ollama (which can be self-hosted) to choose to filter messages based on plain-text criteria.
	Ollama OllamaFilterer
	// Use OpenAI to choose to filter messages based on plain-text criteria.
	OpenAI OpenAIFilterer
	// Remove all but these fields for this filter step. You can have a filter step that only selects fields.
	SelectedFields []string
}

type AnnotateStep struct {
	// Look up geolocation, including distance from a reference point to aircraft, from a tar1090 instance (which can be self-hosted)
	Tar1090 Tar1090Annotator
	// Use Ollama (which can be self-hosted) to annotate messages, such as to answer custom questions about the message ("Is this message about coffee makers?").
	Ollama OllamaAnnotator
	// 	// Look up geolocation, including distance from a reference point to aircraft, from ADSB-Exchange
	ADSB ADSBExchangeAnnotator
}

type ReceiverStep struct {
	// Send messages to a Discord channel using a webhook created from that channel.
	Discord DiscordReceiver
	// Create posts with messages using Mastodon.
	Mastodon MastodonReceiver
	// Send messages to NewRelic as a custom event type.
	NewRelic NewRelicReceiver
	// Generic webhook receiver. Please read README for how to use custom payloads.
	Webhook WebHookReceiver
}

type OllamaCommonConfig struct {
	// Model to use (you need to pull this in Ollama to use it).
	Model string `jsonschema:"required,default=llama3.2" default:"llama3.2"`
	// URL to the Ollama instance to use (include protocol and port). Use
	// 'ollama.com' if you're using Ollama Turbo and also set APIKey.
	URL string `jsonschema:"required,example=http://ollama-service:11434" default:"http://ollama-service:11434"`
	// API key to include in requests.
	APIKey string `jsonschema:",omitempty,example=1234d54321e" default:"your api key here"`
	// Override the system prompt (not usually necessary). This instructs Ollama how to behave with user prompts (ex: pretend you are a pirate. all answers must end in "arrr!"). This might make other options less effective.
	SystemPrompt string `default:"Answer like a pirate"`
	// Instructions for Ollama for processing messages. More detail produces better results.
	UserPrompt string `jsonschema:"required,example=Is there prose in this message?"  default:"Tell the LLM how to handle the message"`
	// Maximum number of retries to make against the Ollama URL.
	MaxRetryAttempts int `default:"5"`
	// How long to wait before retrying the Ollama API.
	MaxRetryDelaySeconds int `default:"5"`
	// How long to wait until giving up on any request to Ollama.
	Timeout int `default:"5"`
	// Options to pass to the model
	Options []OllamaOptionsConfig // The default for this is set in schema.go
}

type OllamaOptionsConfig struct {
	// Option name, specific to the model you are using.
	Name string `jsonschema:"required,default=example_value" default:"example_value"`
	// Value for this particular option, any value is allowed.
	Value any `jsonschema:"required,default=0.1" default:"0.1"`
}
