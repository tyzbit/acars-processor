package main

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
)

type VDLM2 struct {
}

func (v VDLM2Message) Name() string {
	return "VDLM2Message"
}

func (v VDLM2Message) GetDefaultFields() APMessage {
	return VDLM2Message{}.Prepare()
}

// This is the format VDLM2Hub sends
type VDLM2Message struct {
	gorm.Model
	ProcessingStartedAt      time.Time
	ProcessingFinishedAt     time.Time
	Processed                bool
	TrackingLink             string  `ap:"TrackingLink"`
	PhotosLink               string  `ap:"PhotosLink"`
	ThumbnailLink            string  `ap:"ThumbnailLink"`
	ImageLink                string  `ap:"ImageLink"`
	TranslateLink            string  `ap:"TranslateLink"`
	ACARSDramaTailNumberLink string  `ap:"ACARSDramaTailNumberLink"`
	FrequencyMHz             float64 `ap:"FrequencyMhz"`
	From                     string  `ap:"From"`
	// The rest of the struct is the actual message from ACARSHub
	VDL2 struct {
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
				Registration          string `json:"reg" ap:"TailCode"`
				Mode                  string `json:"mode"`
				Label                 string `json:"label"`
				BlockID               string `json:"blk_id"`
				Acknowledge           any    `json:"ack" gorm:"type:string"`
				FlightNumber          string `json:"flight" ap:"FlightNumber"`
				MessageNumber         string `json:"msg_num"`
				MessageNumberSequence string `json:"msg_num_seq"`
				MessageText           string `json:"msg_text" ap:"MessageText"`
			} `json:"acars" gorm:"embedded"`
		} `json:"avlc" gorm:"embedded"`
		BurstLengthOctets    int     `json:"burst_len_octets"`
		FrequencyHz          int     `json:"freq" ap:"FrequencyHz"`
		Index                int     `json:"idx"`
		FrequencySkew        float64 `json:"freq_skew"`
		HDRBitsFixed         int     `json:"hdr_bits_fixed"`
		NoiseLevel           float64 `json:"noise_level"`
		OctetsCorrectedByFEC int     `json:"octets_corrected_by_fec"`
		SignalLevel          float64 `json:"sig_level" ap:"SignalLeveldBm"`
		Station              string  `json:"station" ap:"StationId"`
		Timestamp            struct {
			UnixTimestamp int64 `json:"sec" ap:"UnixTimestamp"`
			Microseconds  int64 `json:"usec"`
		} `json:"t" gorm:"embedded"`
	} `json:"vdl2" gorm:"embedded"`
}

func (v VDLM2Message) Prepare() (result APMessage) {
	// Chop off leading periods
	v.VDL2.AVLC.ACARS.Registration, _ = strings.CutPrefix(v.VDL2.AVLC.ACARS.Registration, ".")
	var thumbnail, link string
	img := getImageByRegistration(v.VDL2.AVLC.ACARS.Registration)
	if img != nil {
		thumbnail = img.ThumbnailLarge.Src
		link = img.Link
	}
	v.VDL2.AVLC.ACARS.Registration = strings.TrimPrefix(v.VDL2.AVLC.ACARS.Registration, ".")
	v.TrackingLink = FlightAwareRoot + v.VDL2.AVLC.ACARS.Registration
	v.PhotosLink = FlightAwarePhotos + v.VDL2.AVLC.ACARS.Registration
	v.ThumbnailLink = thumbnail
	v.ImageLink = link
	v.TranslateLink = fmt.Sprintf(GoogleTranslateLink, url.QueryEscape(v.VDL2.AVLC.ACARS.MessageText))
	v.ACARSDramaTailNumberLink = fmt.Sprintf(ACARSDramaTailNumberLink, v.VDL2.AVLC.ACARS.Registration)
	v.FrequencyMHz = float64(v.VDL2.FrequencyHz) / 1000000
	v.From = AircraftOrTower(v.VDL2.AVLC.ACARS.FlightNumber)

	result = FormatAsAPMessage(v, v.Name())
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
