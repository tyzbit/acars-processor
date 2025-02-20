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
)

type Config struct {
	ACARSHubHost                     string `env:"ACARSHUB_HOST"`
	ACARSHubPort                     int    `env:"ACARSHUB_PORT"`
	ADSBExchangeEnabled              bool   `env:"ADSBEXCHANGE_ENABLED"`
	ADSBExchangeAPIKey               string `env:"ADBSEXCHANGE_APIKEY"`
	ADSBExchangeReferenceGeolocation string `env:"ADBSEXCHANGE_REFERENCE_GEOLOCATION"`
	LogLevel                         string `env:"LOGLEVEL"`
	NewRelicLicenseKey               string `env:"NEW_RELIC_LICENSE_KEY"`
	NewRelicLicenseCustomEventType   string `env:"NEW_RELIC_CUSTOM_EVENT_TYPE"`
	WebhookURL                       string `env:"WEBHOOK_URL"`
	WebhookHeaders                   string `env:"WEBHOOK_HEADERS"`
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
	// Add handlers based on what's enabled
	if config.ADSBExchangeEnabled {
		log.Info("ADSB handler enabled")
		if config.ADSBExchangeAPIKey == "" {
			log.Error("ADSB API key not set")
		}
		enabledAnnotators = append(enabledAnnotators, ADSBHandlerAnnotator{})
	}
	if len(enabledAnnotators) == 0 {
		log.Warn("no annotators are enabled")
	}

	// Add receivers based on what's enabled
	if config.WebhookURL != "" {
		log.Info("Webhook receiver enabled")
		enabledReceivers = append(enabledReceivers, WebhookHandlerReciever{})
	}
	if config.NewRelicLicenseKey != "" {
		log.Info("New Relic reciever enabled")
		enabledReceivers = append(enabledReceivers, NewRelicHandlerReciever{})
	}
	if len(enabledReceivers) == 0 {
		log.Warn("no receivers are enabled")
	}

	SubscribeToACARSHub()
}
