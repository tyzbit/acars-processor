package main

import (
	"strings"

	cfg "github.com/golobby/config/v3"
	"github.com/golobby/config/v3/pkg/feeder"
	log "github.com/sirupsen/logrus"
)

var (
	config   = Config{}
	handlers = []ACARSHandler{}
)

type Config struct {
	ACARSHost                        string `env:"ACARS_HOST"`
	ACARSPort                        string `env:"ACARS_PORT"`
	ACARSTransport                   string `env:"ACARS_TRANSPORT"`
	ADSBExchangeEnabled              bool   `env:"ADSBEXCHANGE_ENABLED"`
	ADSBExchangeAPIKey               string `env:"ADBSEXCHANGE_APIKEY"`
	ADSBExchangeReferenceGeolocation string `env:"ADBSEXCHANGE_REFERENCE_GEOLOCATION"`
	LogLevel                         string `env:"LOGLEVEL"`
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
		log.Info("ADSB enabled")
		if config.ADSBExchangeAPIKey == "" {
			log.Error("ADSB API key not set")
		}
		handlers = append(handlers, ADSBHandler{})
	}
	SubscribeToACARSHub()
}
