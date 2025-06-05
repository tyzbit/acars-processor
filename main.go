package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
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
	ConfigureAnnotators()
	ConfigureReceivers()
	ConfigureFilters()

	go SubscribeToACARSHub()

	// Listen for signals from the OS
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
