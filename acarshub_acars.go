package main

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (a ACARSMessage) Name() string {
	return "ACARSMessage"
}

func (a ACARSMessage) GetDefaultFields() APMessage {
	return ACARSMessage{}.Prepare()
}

// This is the format ACARSHub sends for ACARS messages
type ACARSMessage struct {
	gorm.Model
	ProcessingStartedAt  time.Time
	ProcessingFinishedAt time.Time
	Processed            bool

	// The rest of the struct is the actual message from ACARSHub
	FrequencyMHz float64 `json:"freq" ap:"FrequencyMHz"`
	Channel      int     `json:"channel" ap:"Channel"`
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
	StationID        string `json:"station_id" ap:"StationId"`
	ASSStatus        string `json:"assstat"`
	Mode             string `json:"mode"`
	Label            string `json:"label"`
	BlockID          string `json:"block_id"`
	Acknowledge      any    `json:"ack" gorm:"type:string"` // Can be bool or string
	AircraftTailCode string `json:"tail" ap:"TailCode"`     // Can be string or float
	MessageText      string `json:"text" ap:"MessageText"`
	MessageNumber    string `json:"msgno"`
	FlightNumber     string `json:"flight" ap:"FlightNumber"`
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
	result = FormatAsAPMessage(a, a.Name())

	// Sometimes tail numbers lead with periods, chop them off
	result[ACARSProcessorPrefix+"TailCode"] = strings.TrimLeft(a.AircraftTailCode, ".")

	// Extra helper or common fields
	result[ACARSProcessorPrefix+"TrackingLink"] = FlightAwareRoot + a.AircraftTailCode
	result[ACARSProcessorPrefix+"PhotosLink"] = FlightAwarePhotos + a.AircraftTailCode
	result[ACARSProcessorPrefix+"ThumbnailLink"] = thumbnail
	result[ACARSProcessorPrefix+"ImageLink"] = link
	result[ACARSProcessorPrefix+"TranslateLink"] = fmt.Sprintf(GoogleTranslateLink, url.QueryEscape(a.MessageText))
	result[ACARSProcessorPrefix+"ACARSDramaTailNumberLink"] = fmt.Sprintf(ACARSDramaTailNumberLink, a.AircraftTailCode)
	result[ACARSProcessorPrefix+"UnixTimestamp"] = int64(a.Timestamp)
	result[ACARSProcessorPrefix+"FrequencyHz"] = int(a.FrequencyMHz * 1000000)
	result[ACARSProcessorPrefix+"From"] = AircraftOrTower(a.FlightNumber)

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
