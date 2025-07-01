package annotator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/tyzbit/acars-processor/acarshub"
	. "github.com/tyzbit/acars-processor/config"
	"github.com/tyzbit/acars-processor/util"
)

const (
	FlightAwareRoot   = "https://flightaware.com/live/flight/"
	FlightAwarePhotos = "https://www.flightaware.com/photos/aircraft/"
)

type ACARSAnnotator interface {
	Name() string
	AnnotateACARSMessage(acarshub.ACARSMessage) Annotation
}

type ACARSAnnotatorHandler struct {
}

type VDLM2Annotator interface {
	Name() string
	AnnotateVDLM2Message(acarshub.VDLM2Message) Annotation
}

type VDLM2AnnotatorHandler struct {
}

func (a ACARSAnnotatorHandler) Name() string {
	return "acars"
}

func (a ACARSAnnotatorHandler) DefaultFields() []string {
	// ACARS
	fields := []string{}
	for field := range a.AnnotateACARSMessage(acarshub.ACARSMessage{}) {
		fields = append(fields, field)
	}
	slices.Sort(fields)
	return fields
}

// Interface function to satisfy ACARSHandler
func (a ACARSAnnotatorHandler) AnnotateACARSMessage(m acarshub.ACARSMessage) (annotation Annotation) {
	// Chop off leading periods
	tailcode, _ := strings.CutPrefix(m.AircraftTailCode, ".")
	text := m.MessageText
	// Please update Config example values if changed
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
		"acarsAcknowledge":      fmt.Sprint(m.Acknowledge),
		"acarsAircraftTailCode": tailcode,
		"acarsMessageFrom":      util.AircraftOrTower(m.FlightNumber),
		"acarsMessageText":      text,
		"acarsMessageNumber":    m.MessageNumber,
		"acarsFlightNumber":     m.FlightNumber,
		"acarsExtraURL":         FlightAwareRoot + tailcode,
		"acarsExtraPhotos":      FlightAwarePhotos + tailcode,
	}
	return annotation
}

func (v VDLM2AnnotatorHandler) Name() string {
	return "vdlm2"
}

func (v VDLM2AnnotatorHandler) SelectFields(annotation Annotation) Annotation {
	if Config.Annotators.VDLM2.SelectedFields == nil {
		return annotation
	}
	selectedFields := Annotation{}
	for field, value := range annotation {
		if slices.Contains(Config.Annotators.VDLM2.SelectedFields, field) {
			selectedFields[field] = value
		}
	}
	return selectedFields
}

func (v VDLM2AnnotatorHandler) DefaultFields() []string {
	// ACARS
	fields := []string{}
	for field := range v.AnnotateVDLM2Message(acarshub.VDLM2Message{}) {
		fields = append(fields, field)
	}
	slices.Sort(fields)
	return fields
}

// Interface function to satisfy ACARSHandler
// Although this is the ACARS annotator, we must support ACARS and VLM2
// message types
func (v VDLM2AnnotatorHandler) AnnotateVDLM2Message(m acarshub.VDLM2Message) (annotation Annotation) {
	tailcode, _ := strings.CutPrefix(m.VDL2.AVLC.ACARS.Registration, ".")
	text := m.VDL2.AVLC.ACARS.MessageText
	// Please update config example values if changed
	annotation = Annotation{
		"vdlm2AppName":               m.VDL2.App.Name,
		"vdlm2AppVersion":            m.VDL2.App.Version,
		"vdlm2AppProxied":            m.VDL2.App.Proxied,
		"vdlm2AppProxiedBy":          m.VDL2.App.ProxiedBy,
		"vdlm2AppRouterVersion":      m.VDL2.App.ACARSRouterVersion,
		"vdlm2AppRouterUUID":         m.VDL2.App.ACARSRouterUUID,
		"vdlmCR":                     m.VDL2.AVLC.CR,
		"vdlmDestinationAddress":     m.VDL2.AVLC.Destination.Address,
		"vdlmDestinationType":        m.VDL2.AVLC.Destination.Type,
		"vdlmFrameType":              m.VDL2.AVLC.FrameType,
		"vdlmSourceAddress":          m.VDL2.AVLC.Source.Address,
		"vdlmSourceType":             m.VDL2.AVLC.Source.Type,
		"vdlmSourceStatus":           m.VDL2.AVLC.Source.Status,
		"vdlmRSequence":              m.VDL2.AVLC.RSequence,
		"vdlmSSequence":              m.VDL2.AVLC.SSequence,
		"vdlmPoll":                   m.VDL2.AVLC.Poll,
		"vdlm2BurstLengthOctets":     m.VDL2.BurstLengthOctets,
		"vdlm2FrequencyHz":           m.VDL2.FrequencyHz,
		"vdlm2Index":                 m.VDL2.Index,
		"vdlm2FrequencySkew":         m.VDL2.FrequencySkew,
		"vdlm2HDRBitsFixed":          m.VDL2.HDRBitsFixed,
		"vdlm2NoiseLevel":            m.VDL2.NoiseLevel,
		"vdlm2OctetsCorrectedByFEC":  m.VDL2.OctetsCorrectedByFEC,
		"vdlm2SignalLeveldBm":        m.VDL2.SignalLevel,
		"vdlm2Station":               m.VDL2.Station,
		"vdlm2Timestamp":             m.VDL2.Timestamp.UnixTimestamp,
		"vdlm2TimestampMicroseconds": m.VDL2.Timestamp.Microseconds,
		// These fields are identical to ACARS, so they will have the ACARS prefix
		"acarsErrorCode":             m.VDL2.AVLC.ACARS.Error,
		"acarsCRCOK":                 m.VDL2.AVLC.ACARS.CRCOK,
		"acarsMore":                  m.VDL2.AVLC.ACARS.More,
		"acarsAircraftTailCode":      tailcode,
		"acarsMode":                  m.VDL2.AVLC.ACARS.Mode,
		"acarsLabel":                 m.VDL2.AVLC.ACARS.Mode,
		"acarsBlockID":               m.VDL2.AVLC.ACARS.BlockID,
		"acarsAcknowledge":           m.VDL2.AVLC.ACARS.Acknowledge,
		"acarsFlightNumber":          m.VDL2.AVLC.ACARS.FlightNumber,
		"acarsMessageFrom":           util.AircraftOrTower(m.VDL2.AVLC.ACARS.FlightNumber),
		"acarsMessageNumber":         m.VDL2.AVLC.ACARS.MessageNumber,
		"acarsMessageNumberSequence": m.VDL2.AVLC.ACARS.MessageNumberSequence,
		"acarsMessageText":           text,
		"acarsExtraURL":              FlightAwareRoot + tailcode,
		"acarsExtraPhotos":           FlightAwarePhotos + tailcode,
	}
	return annotation
}
