package main

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ACARSAnnotator interface {
	Name() string
	AnnotateACARSMessage(ACARSMessage) Annotation
	SelectFields(Annotation) Annotation
}

type ACARSAnnotatorHandler struct {
}

func (a ACARSAnnotatorHandler) Name() string {
	return "acars"
}

func (a ACARSAnnotatorHandler) SelectFields(annotation Annotation) Annotation {
	if config.Annotators.ACARS.SelectedFields == nil {
		return annotation
	}
	selectedFields := Annotation{}
	for field, value := range annotation {
		if slices.Contains(config.Annotators.ACARS.SelectedFields, field) {
			selectedFields[field] = value
		}
	}
	return selectedFields
}

func (a ACARSAnnotatorHandler) DefaultFields() []string {
	// ACARS
	fields := []string{}
	for field := range a.AnnotateACARSMessage(ACARSMessage{}) {
		fields = append(fields, field)
	}
	slices.Sort(fields)
	return fields
}

// This is the format ACARSHub sends for ACARS messages
type ACARSMessage struct {
	gorm.Model
	ProcessingStartedAt time.Time
	FrequencyMHz        float64 `json:"freq"`
	Channel             int     `json:"channel"`
	ErrorCode           int     `json:"error"`
	SignaldBm           float64 `json:"level"`
	Timestamp           float64 `json:"timestamp"`
	App                 struct {
		Name               string `json:"name"`
		Version            string `json:"version"`
		Proxied            bool   `json:"proxied"`
		ProxiedBy          string `json:"proxied_by"`
		ACARSRouterVersion string `json:"acars_router_version"`
		ACARSRouterUUID    string `json:"acars_router_uuid"`
	} `json:"app" gorm:"embedded"`
	StationID        string `json:"station_id"`
	ASSStatus        string `json:"assstat"`
	Mode             string `json:"mode"`
	Label            string `json:"label"`
	BlockID          string `json:"block_id"`
	Acknowledge      any    `json:"ack" gorm:"type:string"` // Can be bool or string
	AircraftTailCode string `json:"tail"`                   // Can be string or float
	MessageText      string `json:"text"`
	MessageNumber    string `json:"msgno"`
	FlightNumber     string `json:"flight"`
}

// Interface function to satisfy ACARSHandler
func (a ACARSAnnotatorHandler) AnnotateACARSMessage(m ACARSMessage) (annotation Annotation) {
	// Chop off leading periods
	tailcode, _ := strings.CutPrefix(m.AircraftTailCode, ".")
	text := m.MessageText
	var thumbnail, link string
	img := getImageByRegistration(tailcode)
	if img != nil {
		thumbnail = img.ThumbnailLarge.Src
		link = img.Link
	}
	// Please update config example values if changed
	annotation = Annotation{
		"acarsFrequencyMHz":                  m.FrequencyMHz,
		"acarsChannel":                       m.Channel,
		"acarsErrorCode":                     m.ErrorCode,
		"acarsSignaldBm":                     m.SignaldBm,
		"acarsTimestamp":                     m.Timestamp,
		"acarsAppName":                       m.App.Name,
		"acarsAppVersion":                    m.App.Version,
		"acarsAppProxied":                    m.App.Proxied,
		"acarsAppProxiedBy":                  m.App.ProxiedBy,
		"acarsAppRouterVersion":              m.App.ACARSRouterVersion,
		"acarsAppRouterUUID":                 m.App.ACARSRouterUUID,
		"acarsStationID":                     m.StationID,
		"acarsASSStatus":                     m.ASSStatus,
		"acarsMode":                          m.Mode,
		"acarsLabel":                         m.Label,
		"acarsBlockID":                       m.BlockID,
		"acarsAcknowledge":                   fmt.Sprint(m.Acknowledge),
		"acarsAircraftTailCode":              tailcode,
		"acarsMessageFrom":                   AircraftOrTower(m.FlightNumber),
		"acarsMessageText":                   text,
		"acarsMessageNumber":                 m.MessageNumber,
		"acarsFlightNumber":                  m.FlightNumber,
		"acarsExtraTrackingLink":             FlightAwareRoot + tailcode,
		"acarsExtraPhotosLink":               FlightAwarePhotos + tailcode,
		"acarsExtraThumbnailLink":            thumbnail,
		"acarsExtraImageLink":                link,
		"acarsExtraTranslateLink":            fmt.Sprintf(GoogleTranslateLink, url.QueryEscape(m.MessageText)),
		"acarsExtraACARSDramaTailNumberLink": fmt.Sprintf(ACARSDramaTailNumberLink, m.AircraftTailCode),
	}
	return annotation
}
