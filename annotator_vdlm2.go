package main

import (
	"slices"
	"strings"
)

type VDLM2Annotator interface {
	Name() string
	AnnotateVDLM2Message(VDLM2Message) Annotation
	SelectFields(Annotation) Annotation
}

type VDLM2HandlerAnnotator struct {
}

func (v VDLM2HandlerAnnotator) Name() string {
	return "vdlm2"
}

func (v VDLM2HandlerAnnotator) SelectFields(annotation Annotation) Annotation {
	if config.Annotators.VDLM2.SelectedFields == nil {
		return annotation
	}
	selectedFields := Annotation{}
	for field, value := range annotation {
		if slices.Contains(config.Annotators.VDLM2.SelectedFields, field) {
			selectedFields[field] = value
		}
	}
	return selectedFields
}

// This is the format ACARSHub sends
type VDLM2Message struct {
	VDL2 struct {
		App struct {
			Name               string `json:"name"`
			Version            string `json:"ver"`
			Proxied            bool   `json:"proxied"`
			ProxiedBy          string `json:"proxied_by"`
			ACARSRouterVersion string `json:"acars_router_version"`
			ACARSRouterUUID    string `json:"acars_router_uuid"`
		} `json:"app"`
		AVLC struct {
			CR          string `json:"cr"`
			Destination struct {
				Address string `json:"addr"`
				Type    string `json:"type"`
			} `json:"dst"`
			FrameType string `json:"frame_type"`
			Source    struct {
				Address string `json:"addr"`
				Type    string `json:"type"`
				Status  string `json:"status"`
			} `json:"src"`
			RSequence int  `json:"rseq"`
			SSequence int  `json:"sseq"`
			Poll      bool `json:"poll"`
			ACARS     struct {
				Error                 bool   `json:"err"`
				CRCOK                 bool   `json:"crc_ok"`
				More                  bool   `json:"more"`
				Registration          string `json:"reg"`
				Mode                  string `json:"mode"`
				Label                 string `json:"label"`
				BlockID               string `json:"blk_id"`
				Acknowledge           any    `json:"ack"`
				FlightNumber          string `json:"flight"`
				MessageNumber         string `json:"msg_num"`
				MessageNumberSequence string `json:"msg_num_seq"`
				MessageText           string `json:"msg_text"`
			} `json:"acars"`
		} `json:"avlc"`
		BurstLengthOctets    int     `json:"burst_len_octets"`
		FrequencyHz          int     `json:"freq"`
		Index                int     `json:"idx"`
		FrequencySkew        float64 `json:"freq_skew"`
		HDRBitsFixed         int     `json:"hdr_bits_fixed"`
		NoiseLevel           float64 `json:"noise_level"`
		OctetsCorrectedByFEC int     `json:"octets_corrected_by_fec"`
		SignalLevel          float64 `json:"sig_level"`
		Station              string  `json:"station"`
		Timestamp            struct {
			UnixTimestamp int `json:"sec"`
			Microseconds  int `json:"usec"`
		} `json:"t"`
	} `json:"vdl2"`
}

// Interface function to satisfy ACARSHandler
// Although this is the ACARS annotator, we must support ACARS and VLM2
// message types
func (v VDLM2HandlerAnnotator) AnnotateVDLM2Message(m VDLM2Message) (annotation Annotation) {
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
		"acarsMessageFrom":           AircraftOrTower(m.VDL2.AVLC.ACARS.FlightNumber),
		"acarsMessageNumber":         m.VDL2.AVLC.ACARS.MessageNumber,
		"acarsMessageNumberSequence": m.VDL2.AVLC.ACARS.MessageNumberSequence,
		"acarsMessageText":           text,
		"acarsExtraURL":              FlightAwareRoot + tailcode,
		"acarsExtraPhotos":           FlightAwarePhotos + tailcode,
	}
	return annotation
}
