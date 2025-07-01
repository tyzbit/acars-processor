package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	acarshub "github.com/tyzbit/acars-processor/acarshub"
	"github.com/tyzbit/acars-processor/annotator"
	config "github.com/tyzbit/acars-processor/config"
	"github.com/tyzbit/acars-processor/database"
	. "github.com/tyzbit/acars-processor/decorate"
	filter "github.com/tyzbit/acars-processor/filter"
	"github.com/tyzbit/acars-processor/handler"
	receiver "github.com/tyzbit/acars-processor/receiver"
)

var (
	schemaFilePath = "schema.json"
)

func main() {
	var generateSchema bool
	// flags declaration using flag package
	flag.StringVar(&config.ConfigFilePath, "c", config.ConfigFilePath, "Config file path.")
	flag.BoolVar(&generateSchema, "s", false, "Generate schema.json, then exit.")
	flag.Parse()

	// Generate schema only and then exit
	if generateSchema {
		config.GenerateSchema(schemaFilePath)
		return
	}

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("error loading config: %s", err)
	}
	config.ConfigureLogging()
	if err := database.LoadSavedMessages(); err != nil {
		log.Fatal(Attention("unable to initialize database: %s", err))
	}
	// Migrate types native to these packages
	acarshub.AutoMigrate()
	filter.AutoMigrate()
	annotator.ConfigureAnnotators()
	receiver.ConfigureReceivers()
	filter.ConfigureFilters()

	go handler.HandleACARSHub()
	go acarshub.SubscribeToACARSHub()

	// Listen for signals from the OS
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
