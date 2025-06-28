package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/gorm"
)

var (
	config                 = Config{}
	db                     = new(gorm.DB)
	configFilePath         = "config.yaml"
	schemaFilePath         = "schema.json"
	enabledACARSAnnotators = []ACARSAnnotator{}
	enabledVDLM2Annotators = []VDLM2Annotator{}
	enabledReceivers       = []Receiver{}
	enabledFilters         = []string{}
)

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

	LoadConfig()
	ConfigureLogging()
	if err := LoadSavedMessages(); err != nil {
		log.Fatal(Attention("unable to initialize database: %s", err))
	}
	ConfigureAnnotators()
	ConfigureReceivers()
	ConfigureFilters()

	go SubscribeToACARSHub()

	// Listen for signals from the OS
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
