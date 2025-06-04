package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	// "gopkg.in/yaml.v3"
)

var (
	config                 = Config{}
	configFilePath         = "config.yaml"
	schemaFilePath         = "schema.json"
	enabledACARSAnnotators = []ACARSAnnotator{}
	enabledVDLM2Annotators = []VDLM2Annotator{}
	enabledReceivers       = []Receiver{}
	enabledFilters         = []string{}
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

func main() {
	var generateSchema bool
	// flags declaration using flag package
	flag.StringVar(&configFilePath, "c", configFilePath, "Config file path.")
	flag.BoolVar(&generateSchema, "s", false, "Generate schema.json, then exit.")
	flag.Parse()

	// Generate schema only and then exit
	if generateSchema {
		GenerateSchema(schemaFilePath)
		return
	}

	// Load config file, if present
	cb := ReadFile(configFilePath)
	if err := yaml.Unmarshal(cb, &config); err != nil {
		log.Fatalf("unable to load config from %s", configFilePath)
	}

	ConfigureLogging()
	ConfigureAnnotators()
	ConfigureReceivers()
	ConfigureFilters()

	go SubscribeToACARSHub()

	// Listen for signals from the OS
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
