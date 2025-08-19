package main

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
)

type VDLM2 struct {
	ProcessingStep
	// Is this step locked for editing?
	Locked bool
	// Should this be enabled?
	Enabled bool `jsonschema:"default=true" default:"true"`
}

func (v *VDLM2) Name() string {
	return reflect.TypeOf(v).Name()
}

func (v *VDLM2) Lock() bool {
	if !v.Locked {
		v.Locked = true
		return true
	} else {
		return false
	}
}

func (v *VDLM2) Unlock() bool {
	if !v.Locked {
		v.Locked = false
		return true
	} else {
		return false
	}
}

func (v *VDLM2) Enable() error {
	if v.Lock() {
		v.Enabled = true
		v.Unlock()
	} else {
		return fmt.Errorf("unable to enable %s, it is locked", v.Name())
	}
	return nil
}

func (v *VDLM2) Disable() error {
	if v.Lock() {
		v.Enabled = false
		v.Unlock()
	} else {
		return fmt.Errorf("unable to disable %s, it is locked", v.Name())
	}
	return nil
}

func (v *VDLM2) IsEnabled() bool {
	return v.Enabled
}

func (v *VDLM2) GetDefaultFields() APMessage {
	sap := FormatAsAPMessage(VDLM2Message{})
	c := FormatAsAPMessage(VDLM2Calculated{})
	return MergeMaps(sap, c)
}

// This is the format VDLM2Hub sends
type VDLM2Message struct {
	gorm.Model
	ProcessingStartedAt  time.Time
	ProcessingFinishedAt time.Time
	Processed            bool
	VDL2                 struct {
		App struct {
			Name               string `json:"name"`
			Version            string `json:"ver"`
			Proxied            bool   `json:"proxied"`
			ProxiedBy          string `json:"proxied_by"`
			ACARSRouterVersion string `json:"acars_router_version"`
			ACARSRouterUUID    string `json:"acars_router_uuid"`
		} `json:"app" gorm:"embedded"`
		AVLC struct {
			CR          string `json:"cr"`
			Destination struct {
				Address string `json:"addr"`
				Type    string `json:"type"`
			} `json:"dst" gorm:"embedded"`
			FrameType string `json:"frame_type"`
			Source    struct {
				Address string `json:"addr"`
				Type    string `json:"type"`
				Status  string `json:"status"`
			} `json:"src" gorm:"embedded"`
			RSequence int  `json:"rseq"`
			SSequence int  `json:"sseq"`
			Poll      bool `json:"poll"`
			ACARS     struct {
				Error                 bool   `json:"err"`
				CRCOK                 bool   `json:"crc_ok"`
				More                  bool   `json:"more"`
				Registration          string `json:"reg" ap:"tail_code"`
				Mode                  string `json:"mode"`
				Label                 string `json:"label"`
				BlockID               string `json:"blk_id"`
				Acknowledge           any    `json:"ack" gorm:"type:string"`
				FlightNumber          string `json:"flight" ap:"flight_number"`
				MessageNumber         string `json:"msg_num"`
				MessageNumberSequence string `json:"msg_num_seq"`
				MessageText           string `json:"msg_text" ap:"message_text"`
			} `json:"acars" gorm:"embedded"`
		} `json:"avlc" gorm:"embedded"`
		BurstLengthOctets    int     `json:"burst_len_octets"`
		FrequencyHz          int     `json:"freq" ap:"frequency_hz"`
		Index                int     `json:"idx"`
		FrequencySkew        float64 `json:"freq_skew"`
		HDRBitsFixed         int     `json:"hdr_bits_fixed"`
		NoiseLevel           float64 `json:"noise_level"`
		OctetsCorrectedByFEC int     `json:"octets_corrected_by_fec"`
		SignalLevel          float64 `json:"sig_level" ap:"signal_level_dbm"`
		Station              string  `json:"station" ap:"station_id"`
		Timestamp            struct {
			UnixTimestamp int `json:"sec" ap:"unix_timestamp"`
			Microseconds  int `json:"usec"`
		} `json:"t" gorm:"embedded"`
	} `json:"vdl2" gorm:"embedded"`
}

type VDLM2Calculated struct {
	TrackingLink             string
	PhotosLink               string
	ThumbnailLink            string
	ImageLink                string
	TranslateLink            string
	ACARSDramaTailNumberLink string
}

func (v *VDLM2) Annotate(m VDLM2Message) (APMessage, error) {
	// Chop off leading periods
	m.VDL2.AVLC.ACARS.Registration, _ = strings.CutPrefix(m.VDL2.AVLC.ACARS.Registration, ".")
	var thumbnail, link string
	img := getImageByRegistration(m.VDL2.AVLC.ACARS.Registration)
	if img != nil {
		thumbnail = img.ThumbnailLarge.Src
		link = img.Link
	}
	c := VDLM2Calculated{
		TrackingLink:             FlightAwareRoot + m.VDL2.AVLC.ACARS.Registration,
		PhotosLink:               FlightAwarePhotos + m.VDL2.AVLC.ACARS.Registration,
		ThumbnailLink:            thumbnail,
		ImageLink:                link,
		TranslateLink:            fmt.Sprintf(GoogleTranslateLink, url.QueryEscape(m.VDL2.AVLC.ACARS.MessageText)),
		ACARSDramaTailNumberLink: fmt.Sprintf(ACARSDramaTailNumberLink, m.VDL2.AVLC.ACARS.Registration),
	}
	return MergeMaps(FormatAsAPMessage(c), FormatAsAPMessage(m)), nil
}
