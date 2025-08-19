package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

const (
	FlightAwareRoot          = "https://flightaware.com/live/flight/"
	FlightAwarePhotos        = "https://www.flightaware.com/photos/aircraft/"
	WebhookUserAgent         = "github.com/tyzbit/acars-processor"
	GoogleTranslateLink      = "https://translate.google.com/?sl=auto&tl=en&text=%s&op=translate"
	ACARSDramaTailNumberLink = "https://live.acarsdrama.com/tags/%s"
)

func (a AnnotateStep) Annotate(m APMessage) APMessage {
	annotators := []Annotator{
		a.ADSB,
		a.Ollama,
		a.Tar1090,
	}
	var err error
	for _, a := range annotators {
		if !a.Configured() {
			continue
		}
		m, err = a.Annotate(m)
		if err != nil {
			log.Warn(Attention(fmt.Sprintf("%s", err)))
		}
	}
	return m
}
