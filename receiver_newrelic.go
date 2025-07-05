package main

import (
	"context"
	"time"

	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	log "github.com/sirupsen/logrus"
)

const ACARSCustomEventType = "CustomACARS"

type NewRelicHandlerReciever struct {
	Payload any
}

// Must satisfy Receiver interface
func (n NewRelicHandlerReciever) Name() string {
	return "newrelic"
}

// Must satisfy Receiver interface
func (n NewRelicHandlerReciever) SubmitACARSAnnotations(a Annotation) (err error) {
	if config.Receivers.NewRelic.APIKey == "" {
		log.Panic(Attention("New Relic API key not specified"))
	}
	// Create a new harvester for sending telemetry data.
	harvester, err := telemetry.NewHarvester(
		telemetry.ConfigAPIKey(config.Receivers.NewRelic.APIKey),
	)
	if err != nil {
		log.Error(Attention("Error creating harvester:", err))
	}

	// Allow overriding the custom event type if set
	eventType := ACARSCustomEventType
	if config.Receivers.NewRelic.CustomEventType != "" {
		eventType = config.Receivers.NewRelic.CustomEventType
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
	log.Info(Content("calling new relic"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	harvester.HarvestNow(ctx)

	err = ctx.Err()
	return err
}
