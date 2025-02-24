package main

import (
	"strings"

	cfg "github.com/golobby/config/v3"
	"github.com/golobby/config/v3/pkg/feeder"
	log "github.com/sirupsen/logrus"
)

var (
	config            = Config{}
	enabledAnnotators = []ACARSAnnotator{}
	enabledReceivers  = []Receiver{}
	enabledFilters    = []string{}
)

// Set via ENV variables or a .env file
type Config struct {
	ACARSHubHost                     string  `env:"ACARSHUB_HOST"`
	ACARSHubPort                     int     `env:"ACARSHUB_PORT"`
	AnnotateACARS                    bool    `env:"ANNOTATE_ACARS"`
	ACARSAnnotatorSelectedFields     string  `env:"ACARS_ANNOTATOR_SELECTED_FIELDS"`
	ADSBExchangeAPIKey               string  `env:"ADBSEXCHANGE_APIKEY"`
	ADSBExchangeReferenceGeolocation string  `env:"ADBSEXCHANGE_REFERENCE_GEOLOCATION"`
	ADSBAnnotatorSelectedFields      string  `env:"ADSB_ANNOTATOR_SELECTED_FIELDS"`
	FilterCriteriaInclusive          bool    `env:"FILTER_CRITERIA_INCLUSIVE"`
	FilterCriteriaHasText            bool    `env:"FILTER_CRITERIA_HAS_TEXT"`
	FilterCriteriaMatchTailCode      string  `env:"FILTER_CRITERIA_MATCH_TAIL_CODE"`
	FilterCriteriaMatchFlightNumber  string  `env:"FILTER_CRITERIA_MATCH_FLIGHT_NUMBER"`
	FilterCriteriaMatchFrequency     float64 `env:"FILTER_CRITERIA_MATCH_FREQUENCY"`
	FilterCriteriaMatchASSStatus     string  `env:"FILTER_CRITERIA_MATCH_ASSSTATUS"`
	FilterCriteriaAboveSignaldBm     float64 `env:"FILTER_CRITERIA_ABOVE_SIGNAL_DBM"`
	FilterCriteriaBelowSignaldBm     float64 `env:"FILTER_CRITERIA_BELOW_SIGNAL_DBM"`
	FilterCriteriaMatchStationID     string  `env:"FILTER_CRITERIA_MATCH_STATION_ID"`
	LogLevel                         string  `env:"LOGLEVEL"`
	NewRelicLicenseKey               string  `env:"NEW_RELIC_LICENSE_KEY"`
	NewRelicLicenseCustomEventType   string  `env:"NEW_RELIC_CUSTOM_EVENT_TYPE"`
	WebhookURL                       string  `env:"WEBHOOK_URL"`
	WebhookMethod                    string  `env:"WEBHOOK_METHOD"`
	WebhookHeaders                   string  `env:"WEBHOOK_HEADERS"`
	DiscordWebhookURL                string  `env:"DISCORD_WEBHOOK_URL"`
}

// Set up Config, logging
func init() {
	// Read from .env and override from the local environment
	dotEnvFeeder := feeder.DotEnv{Path: ".env"}
	envFeeder := feeder.Env{}

	_ = cfg.New().AddFeeder(dotEnvFeeder).AddStruct(&config).Feed()
	_ = cfg.New().AddFeeder(envFeeder).AddStruct(&config).Feed()

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

func main() {
	// Add annotators based on what's enabled
	if config.AnnotateACARS {
		log.Info("ACARS handler enabled")
		enabledAnnotators = append(enabledAnnotators, ACARSHandlerAnnotator{})
	}
	if config.ADSBExchangeAPIKey != "" {
		log.Info("ADSB handler enabled")
		if config.ADSBExchangeAPIKey == "" {
			log.Error("ADSB API key not set")
		}
		enabledAnnotators = append(enabledAnnotators, ADSBHandlerAnnotator{})
	}
	if len(enabledAnnotators) == 0 {
		log.Warn("no annotators are enabled")
	}

	// -------------------------------------------------------------------------

	// Add receivers based on what's enabled
	if config.WebhookURL != "" {
		log.Info("Webhook receiver enabled")
		enabledReceivers = append(enabledReceivers, WebhookHandlerReciever{})
	}
	if config.NewRelicLicenseKey != "" {
		log.Info("New Relic reciever enabled")
		enabledReceivers = append(enabledReceivers, NewRelicHandlerReciever{})
	}
	if config.DiscordWebhookURL != "" {
		log.Info("Discord reciever enabled")
		enabledReceivers = append(enabledReceivers, DiscordHandlerReciever{})
	}
	if len(enabledReceivers) == 0 {
		log.Warn("no receivers are enabled")
	}

	// -------------------------------------------------------------------------

	// Add filters based on what's enabled
	if config.FilterCriteriaMatchTailCode != "" {
		enabledFilters = append(enabledFilters, "MatchesTailCode")
	}
	if config.FilterCriteriaHasText {
		enabledFilters = append(enabledFilters, "HasText")
	}
	if config.FilterCriteriaMatchFlightNumber != "" {
		enabledFilters = append(enabledFilters, "MatchesFlightNumber")
	}
	if config.FilterCriteriaMatchFrequency != 0.0 {
		enabledFilters = append(enabledFilters, "MatchesFrequency")
	}
	if config.FilterCriteriaMatchStationID != "" {
		enabledFilters = append(enabledFilters, "MatchesStationID")
	}
	if config.FilterCriteriaAboveSignaldBm != 0.0 {
		enabledFilters = append(enabledFilters, "AboveMinimumSignal")
	}
	if config.FilterCriteriaBelowSignaldBm != 0.0 {
		enabledFilters = append(enabledFilters, "BelowMaximumSignal")
	}
	if config.FilterCriteriaMatchASSStatus != "" {
		enabledFilters = append(enabledFilters, "MatchesASSStatus")
	}

	SubscribeToACARSHub()
}
