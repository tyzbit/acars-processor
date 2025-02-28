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
	// Create a new harvester for sending telemetry data.
	harvester, err := telemetry.NewHarvester(
		telemetry.ConfigAPIKey(config.NewRelicLicenseKey),
	)
	if err != nil {
		log.Error("Error creating harvester:", err)
	}

	// Allow overriding the custom event type if set
	eventType := ACARSCustomEventType
	if config.NewRelicLicenseCustomEventType != "" {
		eventType = config.NewRelicLicenseCustomEventType
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
	log.Info("calling new relic")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	harvester.HarvestNow(ctx)

	return err
}
