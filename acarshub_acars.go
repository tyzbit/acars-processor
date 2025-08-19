package main

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ACARS struct {
	ProcessingStep
	// Is this step locked for editing?
	Locked bool
	// Should this be enabled?
	Enabled bool `jsonschema:"default=true" default:"true"`
}

func (a *ACARS) Name() string {
	return reflect.TypeOf(a).Name()
}

func (a *ACARS) Lock() bool {
	if !a.Locked {
		a.Locked = true
		return true
	} else {
		return false
	}
}

func (a *ACARS) Unlock() bool {
	if !a.Locked {
		a.Locked = false
		return true
	} else {
		return false
	}
}

func (a *ACARS) Enable() error {
	if a.Lock() {
		a.Enabled = true
		a.Unlock()
	} else {
		return fmt.Errorf("unable to enable %s, it is locked", a.Name())
	}
	return nil
}

func (a *ACARS) Disable() error {
	if a.Lock() {
		a.Enabled = false
		a.Unlock()
	} else {
		return fmt.Errorf("unable to disable %s, it is locked", a.Name())
	}
	return nil
}

func (a *ACARS) IsEnabled() bool {
	return a.Enabled
}

func (a *ACARS) GetDefaultFields() APMessage {
	sap := FormatAsAPMessage(ACARSMessage{})
	c := FormatAsAPMessage(ADSBExchangeCalculated{})
	return MergeMaps(sap, c)
}

// This is the format ACARSHub sends for ACARS messages
type ACARSMessage struct {
	gorm.Model
	ProcessingStartedAt  time.Time
	ProcessingFinishedAt time.Time
	Processed            bool
	FrequencyMHz         float64 `json:"freq" acars:"frequency_mhz"`
	Channel              int     `json:"channel" acars:"channel"`
	ErrorCode            int     `json:"error"`
	SignaldBm            float64 `json:"level" acars:"signal"`
	Timestamp            float64 `json:"timestamp" acars:"timestamp"`
	App                  struct {
		Name               string `json:"name"`
		Version            string `json:"version"`
		Proxied            bool   `json:"proxied"`
		ProxiedBy          string `json:"proxied_by"`
		ACARSRouterVersion string `json:"acars_router_version"`
		ACARSRouterUUID    string `json:"acars_router_uuid"`
	} `json:"app" gorm:"embedded"`
	StationID        string `json:"station_id" acars:"station_id"`
	ASSStatus        string `json:"assstat"`
	Mode             string `json:"mode"`
	Label            string `json:"label"`
	BlockID          string `json:"block_id"`
	Acknowledge      any    `json:"ack" gorm:"type:string"` // Can be bool or string
	AircraftTailCode string `json:"tail" acars:"tail_code"` // Can be string or float
	MessageText      string `json:"text" acars:"message_text"`
	MessageNumber    string `json:"msgno"`
	FlightNumber     string `json:"flight" acars:"flight_number"`
}

type ACARSCalculated struct {
	TrackingLink             string
	PhotosLink               string
	ThumbnailLink            string
	ImageLink                string
	TranslateLink            string
	ACARSDramaTailNumberLink string
}

func (a *ACARS) Annotate(m ACARSMessage) (APMessage, error) {
	// Chop off leading periods
	m.AircraftTailCode, _ = strings.CutPrefix(m.AircraftTailCode, ".")
	var thumbnail, link string
	img := getImageByRegistration(m.AircraftTailCode)
	if img != nil {
		thumbnail = img.ThumbnailLarge.Src
		link = img.Link
	}
	c := ACARSCalculated{
		TrackingLink:             FlightAwareRoot + m.AircraftTailCode,
		PhotosLink:               FlightAwarePhotos + m.AircraftTailCode,
		ThumbnailLink:            thumbnail,
		ImageLink:                link,
		TranslateLink:            fmt.Sprintf(GoogleTranslateLink, url.QueryEscape(m.MessageText)),
		ACARSDramaTailNumberLink: fmt.Sprintf(ACARSDramaTailNumberLink, m.AircraftTailCode),
	}
	return MergeMaps(FormatAsAPMessage(c), FormatAsAPMessage(m)), nil
}
