package filter

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tyzbit/acars-processor/acarshub"
	. "github.com/tyzbit/acars-processor/config"
	. "github.com/tyzbit/acars-processor/database"
	. "github.com/tyzbit/acars-processor/decorate"
)

type ACARSFilter interface {
	Filter(acarshub.ACARSMessage) bool
}

var EnabledFilters []string

func ConfigureFilters() {
	// -------------------------------------------------------------------------
	// Add filters based on what's enabled
	if Config.Filters.Generic.TailCode != "" {
		EnabledFilters = append(EnabledFilters, "MatchesTailCode")
	}
	if Config.Filters.Generic.HasText {
		EnabledFilters = append(EnabledFilters, "HasText")
	}
	if Config.Filters.Generic.FlightNumber != "" {
		EnabledFilters = append(EnabledFilters, "MatchesFlightNumber")
	}
	if Config.Filters.Generic.Frequency != 0.0 {
		EnabledFilters = append(EnabledFilters, "MatchesFrequency")
	}
	if Config.Filters.Generic.StationID != "" {
		EnabledFilters = append(EnabledFilters, "MatchesStationID")
	}
	if Config.Filters.Generic.AboveSignaldBm != 0.0 {
		EnabledFilters = append(EnabledFilters, "AboveMinimumSignal")
	}
	if Config.Filters.Generic.BelowSignaldBm != 0.0 {
		EnabledFilters = append(EnabledFilters, "BelowMaximumSignal")
	}
	if Config.Filters.Generic.AboveDistanceNm != 0.0 {
		EnabledFilters = append(EnabledFilters, "AboveMinimumSignal")
	}
	if Config.Filters.Generic.BelowDistanceNm != 0.0 {
		EnabledFilters = append(EnabledFilters, "BelowMaximumSignal")
	}
	if Config.Filters.Generic.FromTower {
		EnabledFilters = append(EnabledFilters, "FromTower")
	}
	if Config.Filters.Generic.FromAircraft {
		EnabledFilters = append(EnabledFilters, "FromAircraft")
	}
	if Config.Filters.Generic.ASSStatus != "" {
		EnabledFilters = append(EnabledFilters, "MatchesASSStatus")
	}
	if Config.Filters.Generic.More {
		EnabledFilters = append(EnabledFilters, "More")
	}
	if Config.Filters.Generic.Emergency {
		EnabledFilters = append(EnabledFilters, "Emergency")
	}
	if Config.Filters.Generic.DictionaryPhraseLengthMinimum > 0 {
		EnabledFilters = append(EnabledFilters, "DictionaryPhraseLengthMinimum")
	}
	if Config.Filters.OpenAI.Enabled && Config.Filters.OpenAI.APIKey != "" {
		EnabledFilters = append(EnabledFilters, "OpenAIPromptFilter")
	}
	if Config.Filters.Ollama.Enabled && Config.Filters.Ollama.URL != "" {
		EnabledFilters = append(EnabledFilters, "OllamaPromptFilter")
	}
	if Config.Filters.ACARS.DuplicateMessageSimilarity != 0.0 || Config.Filters.VDLM2.DuplicateMessageSimilarity != 0.0 {
		EnabledFilters = append(EnabledFilters, "MessageSimilarity")
	}
	log.Info(Content("enabled filters: "), Note(strings.Join(EnabledFilters, ",")))
}

func AutoMigrate() {
	// Ollama filter
	if err := DB.AutoMigrate(OllamaFilterResult{}); err != nil {
		log.Fatal(Attention("Unable to automigrate Ollama filter type: %s", err))
	}
}
