package main

const (
	FlightAwareRoot   = "https://flightaware.com/live/flight/"
	FlightAwarePhotos = "https://www.flightaware.com/photos/aircraft/"
	WebhookUserAgent  = "github.com/tyzbit/acars-annotator"
)

type ACARSAnnotator interface {
	AnnotateACARSMessage(ACARSMessage) Annotation
	Name() string
}

type Receiver interface {
	SubmitACARSAnnotations(Annotation) error
	Name() string
}

type ACARSAnnotation struct {
	Annotator string
	Annotation
}

// ALL KEYS MUST BE UNIQUE AMONG ALL ANNOTATORS
type Annotation map[string]interface{}

type AnnotatedACARSMessage struct {
	ACARSMessage
	Annotations map[string]any
}

// This is the format ACARSHub sends
type ACARSMessage struct {
	FrequencyMHz float64 `json:"freq"`
	Channel      int     `json:"channel"`
	ErrorCode    int     `json:"error"`
	SignaldBm    float64 `json:"level"`
	Timestamp    float64 `json:"timestamp"`
	App          struct {
		Name               string `json:"name"`
		Version            string `json:"version"`
		Proxied            bool   `json:"proxied"`
		ProxiedBy          string `json:"proxied_by"`
		ACARSRouterVersion string `json:"acars_router_version"`
		ACARSRouterUUID    string `json:"acars_router_UUID"`
	}
	StationID        string `json:"station_id"`
	ASSStatus        string `json:"assstat"`
	Mode             string `json:"mode"`
	Label            string `json:"label"`
	BlockID          string `json:"block_id"`
	Acknowledge      any    `json:"ack"` // Can be bool or string
	AircraftTailCode string `json:"tail"`
	MessageText      string `json:"text"`
	MessageNumber    string `json:"msgno"`
	FlightNumber     string `json:"flight"`
}
