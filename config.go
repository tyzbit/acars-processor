package main

// Set via ENV variables or a .env file
type Config struct {
	ACARSHubHost                                string  `env:"ACARSHUB_HOST"`
	ACARSHubPort                                int     `env:"ACARSHUB_PORT"`
	AnnotateACARS                               bool    `env:"ANNOTATE_ACARS"`
	ACARSHubVDLM2Host                           string  `env:"ACARSHUB_HOST"`
	ACARSHubVDLM2Port                           int     `env:"ACARSHUB_VDLM2_PORT"`
	AnnotateVDLM2                               bool    `env:"ANNOTATE_VDLM2"`
	TAR1090URL                                  string  `env:"TAR1090_URL"`
	TAR1090ReferenceGeolocation                 string  `env:"TAR1090_REFERENCE_GEOLOCATION"`
	ACARSAnnotatorSelectedFields                string  `env:"ACARS_ANNOTATOR_SELECTED_FIELDS"`
	ADSBExchangeAPIKey                          string  `env:"ADBSEXCHANGE_APIKEY"`
	ADSBExchangeReferenceGeolocation            string  `env:"ADBSEXCHANGE_REFERENCE_GEOLOCATION"`
	ADSBAnnotatorSelectedFields                 string  `env:"ADSB_ANNOTATOR_SELECTED_FIELDS"`
	VDLM2AnnotatorSelectedFields                string  `env:"VDLM2_ANNOTATOR_SELECTED_FIELDS"`
	TAR1090AnnotatorSelectedFields              string  `env:"TAR1090_ANNOTATOR_SELECTED_FIELDS"`
	FilterCriteriaHasText                       bool    `env:"FILTER_CRITERIA_HAS_TEXT"`
	FilterCriteriaMatchTailCode                 string  `env:"FILTER_CRITERIA_MATCH_TAIL_CODE"`
	FilterCriteriaMatchFlightNumber             string  `env:"FILTER_CRITERIA_MATCH_FLIGHT_NUMBER"`
	FilterCriteriaMatchFrequency                float64 `env:"FILTER_CRITERIA_MATCH_FREQUENCY"`
	FilterCriteriaMatchASSStatus                string  `env:"FILTER_CRITERIA_MATCH_ASSSTATUS"`
	FilterCriteriaAboveSignaldBm                float64 `env:"FILTER_CRITERIA_ABOVE_SIGNAL_DBM"`
	FilterCriteriaBelowSignaldBm                float64 `env:"FILTER_CRITERIA_BELOW_SIGNAL_DBM"`
	FilterCriteriaMatchStationID                string  `env:"FILTER_CRITERIA_MATCH_STATION_ID"`
	FilterCriteriaMore                          bool    `env:"FILTER_CRITERIA_MORE"`
	FilterCriteriaAboveDistanceNm               float64 `env:"FILTER_CRITERIA_ABOVE_DISTANCE_NM"`
	FilterCriteriaBelowDistanceNm               float64 `env:"FILTER_CRITERIA_Below_DISTANCE_NM"`
	FilterCriteriaEmergency                     bool    `env:"FILTER_CRITERIA_EMERGENCY"`
	FilterCriteriaDictionaryPhraseLengthMinimum int64   `env:"FILTER_CRITERIA_DICTIONARY_PHRASE_LENGTH_MINIMUM"`
	LogLevel                                    string  `env:"LOGLEVEL"`
	NewRelicLicenseKey                          string  `env:"NEW_RELIC_LICENSE_KEY"`
	NewRelicLicenseCustomEventType              string  `env:"NEW_RELIC_CUSTOM_EVENT_TYPE"`
	WebhookURL                                  string  `env:"WEBHOOK_URL"`
	WebhookMethod                               string  `env:"WEBHOOK_METHOD"`
	WebhookHeaders                              string  `env:"WEBHOOK_HEADERS"`
	DiscordWebhookURL                           string  `env:"DISCORD_WEBHOOK_URL"`
}
