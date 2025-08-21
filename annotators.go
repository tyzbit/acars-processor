package main

import (
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"
)

const (
	FlightAwareRoot          = "https://flightaware.com/live/flight/"
	FlightAwarePhotos        = "https://www.flightaware.com/photos/aircraft/"
	WebhookUserAgent         = "github.com/tyzbit/acars-processor"
	GoogleTranslateLink      = "https://translate.google.com/?sl=auto&tl=en&text=%s&op=translate"
	ACARSDramaTailNumberLink = "https://live.acarsdrama.com/tags/%s"
)

var (
	Annotators = []Annotator{
		AnnotateStep{}.ADSB,
		AnnotateStep{}.Ollama,
		AnnotateStep{}.Tar1090,
	}
)

func (as AnnotateStep) Annotate(m APMessage) APMessage {
	var err error
	for _, a := range Annotators {
		if !a.Configured() {
			continue
		}
		m, err = a.Annotate(m)
		if err != nil {
			log.Warn(Attention(fmt.Sprintf("%s annotator: %s", a.Name(), err)))
		}
	}
	// Only keep SelectedFields
	if len(as.SelectedFields) > 0 {
		for messageField := range m {
			if !slices.Contains(as.SelectedFields, messageField) {
				delete(m, messageField)
			}
		}
	}
	return m
}
