package main

const (
	FlightAwareRoot   = "https://flightaware.com/live/flight/"
	FlightAwarePhotos = "https://www.flightaware.com/photos/aircraft/"
	WebhookUserAgent  = "github.com/tyzbit/acars-annotator"
)

type ACARSAnnotator interface {
	Name() string
	AnnotateACARSMessage(ACARSMessage) Annotation
	AnnotateVDLM2Message(VDLM2Message) Annotation
	SelectFields(Annotation) Annotation
}

// ALL KEYS MUST BE UNIQUE AMONG ALL ANNOTATORS
type Annotation map[string]interface{}

type Receiver interface {
	SubmitACARSAnnotations(Annotation) error
	Name() string
}

type ACARSFilter interface {
	Filter(ACARSMessage) bool
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
