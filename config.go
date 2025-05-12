package main

// Set via ENV variables or a .env file
type Config struct {
	// General
	LogLevel string `env:"LOGLEVEL"`
	// Annotators
	ACARSHubHost                       string `env:"ACARSHUB_HOST"`
	ACARSHubPort                       int    `env:"ACARSHUB_PORT"`
	ACARSHubMaxConcurrentRequests      int    `env:"ACARSHUB_MAX_CONCURRENT_REQUESTS_PER_SUBSCRIBER"`
	AnnotateACARS                      bool   `env:"ANNOTATE_ACARS"`
	ACARSHubVDLM2Host                  string `env:"ACARSHUB_VDLM2_HOST"`
	ACARSHubVDLM2Port                  int    `env:"ACARSHUB_VDLM2_PORT"`
	AnnotateVDLM2                      bool   `env:"ANNOTATE_VDLM2"`
	TAR1090URL                         string `env:"TAR1090_URL"`
	TAR1090ReferenceGeolocation        string `env:"TAR1090_REFERENCE_GEOLOCATION"`
	ACARSAnnotatorSelectedFields       string `env:"ACARS_ANNOTATOR_SELECTED_FIELDS"`
	ADSBExchangeAPIKey                 string `env:"ADBSEXCHANGE_APIKEY"`
	ADSBExchangeReferenceGeolocation   string `env:"ADBSEXCHANGE_REFERENCE_GEOLOCATION"`
	OllamaAnnotatorModel               string `env:"OLLAMA_ANNOTATOR_MODEL"`
	OllamaAnnotatorURL                 string `env:"OLLAMA_ANNOTATOR_URL"`
	OllamaAnnotatorSelectedFields      string `env:"OLLAMA_ANNOTATOR_SELECTED_FIELDS"`
	OllamaAnnotatorUserPrompt          string `env:"OLLAMA_ANNOTATOR_PROMPT"`
	OllamaAnnotatorFilterWithQuestion  bool   `env:"OLLAMA_ANNOTATOR_FILTER_WITH_QUESTION"`
	OllamaAnnotatorMaxPredictionTokens int    `env:"OLLAMA_ANNOTATOR_MAX_PREDICTION_TOKENS"`
	OllamaAnnotatorMaxRetryAttempts    int    `env:"OLLAMA_ANNOTATOR_MAX_RETRY_ATTEMPTS"`
	OllamaAnnotatorRetryDelaySeconds   int    `env:"OLLAMA_ANNOTATOR_RETRY_DELAY"`
	OllamaAnnotatorSystemPrompt        string `env:"OLLAMA_ANNOTATOR_SYSTEM_PROMPT"`
	OllamaAnnotatorTimeout             int    `env:"OLLAMA_ANNOTATOR_TIMEOUT"`
	// Filters
	OllamaFilterOnFailure                         bool    `env:"FILTER_OLLAMA_FILTER_ON_FAILURE"`
	OllamaFilterModel                             string  `env:"FILTER_OLLAMA_MODEL"`
	OllamaFilterURL                               string  `env:"FILTER_OLLAMA_URL"`
	OllamaFilterUserPrompt                        string  `env:"FILTER_OLLAMA_PROMPT"`
	OllamaFilterMaxPredictionTokens               int     `env:"FILTER_OLLAMA_MAX_PREDICTION_TOKENS"`
	OllamaFilterMaxRetryAttempts                  int     `env:"FILTER_OLLAMA_MAX_RETRY_ATTEMPTS"`
	OllamaFilterRetryDelaySeconds                 int     `env:"FILTER_OLLAMA_RETRY_DELAY"`
	OllamaFilterSystemPrompt                      string  `env:"FILTER_OLLAMA_SYSTEM_PROMPT"`
	OllamaFilterTimeout                           int     `env:"FILTER_OLLAMA_TIMEOUT"`
	OpenAIAPIKey                                  string  `env:"FILTER_OPENAI_APIKEY"`
	OpenAIModel                                   string  `env:"FILTER_OPENAI_MODEL"`
	OpenAIUserPrompt                              string  `env:"FILTER_OPENAI_PROMPT"`
	OpenAISystemPrompt                            string  `env:"FILTER_OPENAI_SYSTEM_PROMPT"`
	ADSBAnnotatorSelectedFields                   string  `env:"ADSB_ANNOTATOR_SELECTED_FIELDS"`
	VDLM2AnnotatorSelectedFields                  string  `env:"VDLM2_ANNOTATOR_SELECTED_FIELDS"`
	TAR1090AnnotatorSelectedFields                string  `env:"TAR1090_ANNOTATOR_SELECTED_FIELDS"`
	FilterCriteriaHasText                         bool    `env:"FILTER_CRITERIA_HAS_TEXT"`
	FilterCriteriaMatchTailCode                   string  `env:"FILTER_CRITERIA_MATCH_TAIL_CODE"`
	FilterCriteriaMatchFlightNumber               string  `env:"FILTER_CRITERIA_MATCH_FLIGHT_NUMBER"`
	FilterCriteriaMatchFrequency                  float64 `env:"FILTER_CRITERIA_MATCH_FREQUENCY"`
	FilterCriteriaMatchASSStatus                  string  `env:"FILTER_CRITERIA_MATCH_ASSSTATUS"`
	FilterCriteriaAboveSignaldBm                  float64 `env:"FILTER_CRITERIA_ABOVE_SIGNAL_DBM"`
	FilterCriteriaBelowSignaldBm                  float64 `env:"FILTER_CRITERIA_BELOW_SIGNAL_DBM"`
	FilterCriteriaMatchStationID                  string  `env:"FILTER_CRITERIA_MATCH_STATION_ID"`
	FilterCriteriaFromTower                       bool    `env:"FILTER_CRITERIA_FROM_TOWER"`
	FilterCriteriaFromAircraft                    bool    `env:"FILTER_CRITERIA_FROM_AIRCRAFT"`
	FilterCriteriaMore                            bool    `env:"FILTER_CRITERIA_MORE"`
	FilterCriteriaAboveDistanceNm                 float64 `env:"FILTER_CRITERIA_ABOVE_DISTANCE_NM"`
	FilterCriteriaBelowDistanceNm                 float64 `env:"FILTER_CRITERIA_Below_DISTANCE_NM"`
	FilterCriteriaEmergency                       bool    `env:"FILTER_CRITERIA_EMERGENCY"`
	FilterCriteriaDictionaryPhraseLengthMinimum   int64   `env:"FILTER_CRITERIA_DICTIONARY_PHRASE_LENGTH_MINIMUM"`
	FilterCriteriaACARSDuplicateMessageSimilarity float64 `env:"FILTER_CRITERIA_ACARS_DUPLICATE_MESSAGE_SIMILARITY"`
	FilterCriteriaVDLM2DuplicateMessageSimilarity float64 `env:"FILTER_CRITERIA_VDLM2_DUPLICATE_MESSAGE_SIMILARITY"`
	// Receivers
	NewRelicLicenseKey             string `env:"NEW_RELIC_LICENSE_KEY"`
	NewRelicLicenseCustomEventType string `env:"NEW_RELIC_CUSTOM_EVENT_TYPE"`
	WebhookURL                     string `env:"WEBHOOK_URL"`
	WebhookMethod                  string `env:"WEBHOOK_METHOD"`
	WebhookHeaders                 string `env:"WEBHOOK_HEADERS"`
	DiscordWebhookURL              string `env:"DISCORD_WEBHOOK_URL"`
}
