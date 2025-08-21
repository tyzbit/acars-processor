package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	config            = Config{}
	db                = new(gorm.DB)
	configFilePath    = "config.yaml"
	schemaFilePath    = "schema.json"
	configExamplePath = "config_all_options.yaml"
)

func main() {
	var generateSchema bool
	// flags declaration using flag package
	flag.StringVar(&configFilePath, "c", configFilePath, "Config file path.")
	flag.BoolVar(&generateSchema, "s", false, "Generate schema.json, then exit.")
	flag.Parse()

	// Generate schema only and then exit
	if generateSchema {
		var updated bool
		updated = GenerateSchema(schemaFilePath) || updated
		updated = GenerateDocs(configExamplePath) || updated
		if updated {
			log.Info(Content("Files have changed, exiting with nonzero status"))
			os.Exit(100)
		}
		os.Exit(0)
	}

	LoadConfig()
	ConfigureLogging()
	if err := LoadSavedMessages(); err != nil {
		log.Fatal(Attention("unable to initialize database: %s", err))
	}

	go SubscribeToACARSHub()

	// Listen for signals from the OS
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
