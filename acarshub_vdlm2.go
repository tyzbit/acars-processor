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
}

func (v *VDLM2) Name() string {
	return reflect.TypeOf(v).Name()
}

func (v *VDLM2) GetDefaultFields() APMessage {
	sap := FormatAsAPMessage(VDLM2Message{})
	c := FormatAsAPMessage(VDLM2Calculated{})
	return MergeAPMessages(sap, c)
}

// This is the format VDLM2Hub sends
type VDLM2Message struct {
	gorm.Model
	VDLM2Calculated
	ProcessingStartedAt      time.Time
	ProcessingFinishedAt     time.Time
	Processed                bool
	TrackingLink             string  `ap:"tracking_link"`
	PhotosLink               string  `ap:"photos_link"`
	ThumbnailLink            string  `ap:"thumbnail_link"`
	ImageLink                string  `ap:"image_link"`
	TranslateLink            string  `ap:"translate_link"`
	ACARSDramaTailNumberLink string  `ap:"acars_drama_tail_number_link"`
	FrequencyMHz             float64 `ap:"frequency_mhz"`
	From                     string  `ap:"from"`
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
			UnixTimestamp int64 `json:"sec" ap:"unix_timestamp"`
			Microseconds  int64 `json:"usec"`
		} `json:"t" gorm:"embedded"`
	} `json:"vdl2" gorm:"embedded"`
}

// Merged with VDLM2Message before being turned into an APMessage
type VDLM2Calculated struct {
}

func (v VDLM2Message) Prepare() APMessage {
	// Chop off leading periods
	v.VDL2.AVLC.ACARS.Registration, _ = strings.CutPrefix(v.VDL2.AVLC.ACARS.Registration, ".")
	var thumbnail, link string
	img := getImageByRegistration(v.VDL2.AVLC.ACARS.Registration)
	if img != nil {
		thumbnail = img.ThumbnailLarge.Src
		link = img.Link
	}
	v.TrackingLink = FlightAwareRoot + v.VDL2.AVLC.ACARS.Registration
	v.PhotosLink = FlightAwarePhotos + v.VDL2.AVLC.ACARS.Registration
	v.ThumbnailLink = thumbnail
	v.ImageLink = link
	v.TranslateLink = fmt.Sprintf(GoogleTranslateLink, url.QueryEscape(v.VDL2.AVLC.ACARS.MessageText))
	v.ACARSDramaTailNumberLink = fmt.Sprintf(ACARSDramaTailNumberLink, v.VDL2.AVLC.ACARS.Registration)
	v.FrequencyMHz = float64(v.VDL2.FrequencyHz) / 1000000
	v.From = AircraftOrTower(v.VDL2.AVLC.ACARS.FlightNumber)
	return FormatAsAPMessage(v)
}
