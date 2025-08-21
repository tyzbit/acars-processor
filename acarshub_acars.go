package main

import (
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ACARS struct {
	ProcessingStep
}

func (a *ACARS) Name() string {
	return reflect.TypeOf(a).Name()
}

func (a *ACARS) GetDefaultFields() APMessage {
	sap := FormatAsAPMessage(ACARSMessage{})
	c := FormatAsAPMessage(ADSBExchangeCalculated{})
	return MergeAPMessages(sap, c)
}

// This is the format ACARSHub sends for ACARS messages
type ACARSMessage struct {
	gorm.Model
	ProcessingStartedAt      time.Time
	ProcessingFinishedAt     time.Time
	Processed                bool
	TrackingLink             string `ap:"TrackingLink"`
	PhotosLink               string `ap:"PhotosLink"`
	ThumbnailLink            string `ap:"ThumbnailLink"`
	ImageLink                string `ap:"ImageLink"`
	TranslateLink            string `ap:"TranslateLink"`
	ACARSDramaTailNumberLink string `ap:"ACARSDramaTailNumberLink"`
	UnixTimestamp            int64  `ap:"UnixTimestamp"`
	FrequencyHz              int    `ap:"frequency_hz"`
	From                     string `ap:"From"`
	// The rest of the struct is the actual message from ACARSHub
	FrequencyMHz float64 `json:"freq" ap:"FrequencyMhz"`
	Channel      int     `json:"channel" ap:"channel"`
	ErrorCode    int     `json:"error"`
	SignaldBm    float64 `json:"level" ap:"SignalLeveldBm"`
	Timestamp    float64 `json:"timestamp"`
	App          struct {
		Name               string `json:"name"`
		Version            string `json:"version"`
		Proxied            bool   `json:"proxied"`
		ProxiedBy          string `json:"proxied_by"`
		ACARSRouterVersion string `json:"acars_router_version"`
		ACARSRouterUUID    string `json:"acars_router_uuid"`
	} `json:"app" gorm:"embedded"`
	StationID        string `json:"station_id" ap:"station_id"`
	ASSStatus        string `json:"assstat"`
	Mode             string `json:"mode"`
	Label            string `json:"label"`
	BlockID          string `json:"block_id"`
	Acknowledge      any    `json:"ack" gorm:"type:string"` // Can be bool or string
	AircraftTailCode string `json:"tail" ap:"TailCode"`     // Can be string or float
	MessageText      string `json:"text" ap:"MessageText"`
	MessageNumber    string `json:"msgno"`
	FlightNumber     string `json:"flight" ap:"flight_number"`
}

func (a ACARSMessage) Prepare() (result APMessage) {
	// Chop off leading periods
	a.AircraftTailCode, _ = strings.CutPrefix(a.AircraftTailCode, ".")
	var thumbnail, link string
	img := getImageByRegistration(a.AircraftTailCode)
	if img != nil {
		thumbnail = img.ThumbnailLarge.Src
		link = img.Link
	}
	a.TrackingLink = FlightAwareRoot + a.AircraftTailCode
	a.PhotosLink = FlightAwarePhotos + a.AircraftTailCode
	a.ThumbnailLink = thumbnail
	a.ImageLink = link
	a.TranslateLink = fmt.Sprintf(GoogleTranslateLink, url.QueryEscape(a.MessageText))
	a.ACARSDramaTailNumberLink = fmt.Sprintf(ACARSDramaTailNumberLink, a.AircraftTailCode)
	a.UnixTimestamp = int64(a.Timestamp)
	a.FrequencyHz = int(a.FrequencyMHz * 1000000)
	a.From = AircraftOrTower(a.FlightNumber)

	result = FormatAsAPMessage(a)
	selectedFields := config.ACARSProcessorSettings.ACARSHub.ACARS.SelectedFields
	// Remove all but any selected fields
	if len(selectedFields) > 0 {
		for field := range result {
			if !slices.Contains(selectedFields, field) {
				delete(result, field)
			}
		}
	}
	return result
}
