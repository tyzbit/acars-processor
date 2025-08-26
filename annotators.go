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

func (as AnnotateStep) Annotate(m APMessage) APMessage {
	annotators := []Annotator{
		as.ADSB,
		as.Ollama,
		as.Tar1090,
	}
	for _, a := range annotators {
		if !a.Configured() {
			continue
		}
		nm, err := a.Annotate(m)
		if err != nil {
			log.Warn(Attention(fmt.Sprintf("%s: %s", a.Name(), err)))
		}
		m = MergeAPMessages(m, nm)
	}
	return m
}
