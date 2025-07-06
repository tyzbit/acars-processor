package main

const (
	FlightAwareRoot     = "https://flightaware.com/live/flight/"
	FlightAwarePhotos   = "https://www.flightaware.com/photos/aircraft/"
	WebhookUserAgent    = "github.com/tyzbit/acars-processor"
	GoogleTranslateLink = "https://translate.google.com/?sl=auto&tl=en&text=%s&op=translate"
)

// ALL KEYS MUST BE UNIQUE AMONG ALL ANNOTATORS
type Annotation map[string]interface{}

type Receiver interface {
	SubmitACARSAnnotations(Annotation) error
	Name() string
}

type ACARSFilter interface {
	Filter(ACARSMessage) bool
}
