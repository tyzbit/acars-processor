package main

import "strings"

type ACARSAnnotator interface {
	Name() string
	AnnotateACARSMessage(ACARSMessage) Annotation
	SelectFields(Annotation) Annotation
}

type ACARSHandlerAnnotator struct {
}

func (a ACARSHandlerAnnotator) Name() string {
	return "acars"
}

func (a ACARSHandlerAnnotator) SelectFields(annotation Annotation) Annotation {
	if config.ACARSAnnotatorSelectedFields == "" {
		return annotation
	}
	selectedFields := Annotation{}
	for field, value := range annotation {
		if strings.Contains(config.ACARSAnnotatorSelectedFields, field) {
			selectedFields[field] = value
		}
	}
	return selectedFields
}

// This is the format ACARSHub sends for ACARS messages
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
		ACARSRouterUUID    string `json:"acars_router_uuid"`
	} `json:"app"`
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

// Interface function to satisfy ACARSHandler
func (a ACARSHandlerAnnotator) AnnotateACARSMessage(m ACARSMessage) (annotation Annotation) {
	// Chop off leading periods
	tailcode, cut := strings.CutPrefix(m.AircraftTailCode, ".")
	if !cut {
		tailcode = m.AircraftTailCode
	}
	annotation = Annotation{
		"acarsFrequencyMHz":     m.FrequencyMHz,
		"acarsChannel":          m.Channel,
		"acarsErrorCode":        m.ErrorCode,
		"acarsSignaldBm":        m.SignaldBm,
		"acarsTimestamp":        m.Timestamp,
		"acarsAppName":          m.App.Name,
		"acarsAppVersion":       m.App.Version,
		"acarsAppProxied":       m.App.Proxied,
		"acarsAppProxiedBy":     m.App.ProxiedBy,
		"acarsAppRouterVersion": m.App.ACARSRouterVersion,
		"acarsAppRouterUUID":    m.App.ACARSRouterUUID,
		"acarsStationID":        m.StationID,
		"acarsASSStatus":        m.ASSStatus,
		"acarsMode":             m.Mode,
		"acarsLabel":            m.Label,
		"acarsBlockID":          m.BlockID,
		"acarsAcknowledge":      m.Acknowledge,
		"acarsAircraftTailCode": tailcode,
		"acarsMessageText":      m.MessageText,
		"acarsMessageNumber":    m.MessageNumber,
		"acarsFlightNumber":     m.FlightNumber,
		"acarsExtraURL":         FlightAwareRoot + m.AircraftTailCode,
		"acarsExtraPhotos":      FlightAwarePhotos + m.AircraftTailCode,
	}
	return annotation
}
