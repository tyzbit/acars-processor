package main

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	log "github.com/sirupsen/logrus"
)

const ACARSCustomEventType = "CustomACARS"

type NewRelicReceiver struct {
	Module
	Receiver
	// API License key to use New Relic.
	APIKey string `jsonschema:"required" default:"api_key"`
	// Name for the custom event type to create (example if set to "MyCustomACARSEvents": `FROM MyCustomACARSEvents SELECT count(timestamp)`). If not provided, it will be `CustomACARS`.
	CustomEventType string `default:"CustomACARS"`
}

// Must satisfy Receiver interface
func (n NewRelicReceiver) Name() string {
	return reflect.TypeOf(n).Name()
}

// Must satisfy Receiver interface
func (n NewRelicReceiver) Send(a APMessage) (err error) {
	if n.APIKey == "" {
		return fmt.Errorf("New Relic API key not specified: %w", err)
	}
	// Create a new harvester for sending telemetry data.
	harvester, err := telemetry.NewHarvester(
		telemetry.ConfigAPIKey(n.APIKey),
	)
	if err != nil {
		return fmt.Errorf("Error creating harvester: %w", err)
	}

	// Allow overriding the custom event type if set
	eventType := ACARSCustomEventType
	if n.CustomEventType != "" {
		eventType = n.CustomEventType
	}

	event := telemetry.Event{
		EventType:  eventType,
		Attributes: a,
	}

	// Record the custom event.
	err = harvester.RecordEvent(event)
	if err != nil {
		return err
	}

	// Flush events to New Relic. HarvestNow sends any recorded events immediately.
	log.Debug(Content("calling new relic"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	harvester.HarvestNow(ctx)

	err = ctx.Err()
	return err
}

func (f NewRelicReceiver) Configured() bool {
	return !reflect.DeepEqual(f, NewRelicReceiver{})
}

func (f NewRelicReceiver) GetDefaultFields() (s []string) {
	for f := range FormatAsAPMessage(NewRelicReceiver{}) {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
}
