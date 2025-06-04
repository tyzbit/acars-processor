package main

import (
	"strings"

	log "github.com/sirupsen/logrus"
	// English only
)

func ConfigureFilters() {
	// -------------------------------------------------------------------------
	// Add filters based on what's enabled
	if config.Filters.Generic.TailCode != "" {
		enabledFilters = append(enabledFilters, "MatchesTailCode")
	}
	if config.Filters.Generic.HasText {
		enabledFilters = append(enabledFilters, "HasText")
	}
	if config.Filters.Generic.FlightNumber != "" {
		enabledFilters = append(enabledFilters, "MatchesFlightNumber")
	}
	if config.Filters.Generic.Frequency != 0.0 {
		enabledFilters = append(enabledFilters, "MatchesFrequency")
	}
	if config.Filters.Generic.StationID != "" {
		enabledFilters = append(enabledFilters, "MatchesStationID")
	}
	if config.Filters.Generic.AboveSignaldBm != 0.0 {
		enabledFilters = append(enabledFilters, "AboveMinimumSignal")
	}
	if config.Filters.Generic.BelowSignaldBm != 0.0 {
		enabledFilters = append(enabledFilters, "BelowMaximumSignal")
	}
	if config.Filters.Generic.AboveDistanceNm != 0.0 {
		enabledFilters = append(enabledFilters, "AboveMinimumSignal")
	}
	if config.Filters.Generic.BelowDistanceNm != 0.0 {
		enabledFilters = append(enabledFilters, "BelowMaximumSignal")
	}
	if config.Filters.Generic.FromTower {
		enabledFilters = append(enabledFilters, "FromTower")
	}
	if config.Filters.Generic.FromAircraft {
		enabledFilters = append(enabledFilters, "FromAircraft")
	}
	if config.Filters.Generic.ASSStatus != "" {
		enabledFilters = append(enabledFilters, "MatchesASSStatus")
	}
	if config.Filters.Generic.More {
		enabledFilters = append(enabledFilters, "More")
	}
	if config.Filters.Generic.Emergency {
		enabledFilters = append(enabledFilters, "Emergency")
	}
	if config.Filters.Generic.DictionaryPhraseLengthMinimum > 0 {
		enabledFilters = append(enabledFilters, "DictionaryPhraseLengthMinimum")
	}
	if config.Filters.OpenAI.APIKey != "" {
		enabledFilters = append(enabledFilters, "OpenAIPromptFilter")
	}
	if config.Filters.Ollama.URL != "" {
		enabledFilters = append(enabledFilters, "OllamaPromptFilter")
	}
	if config.Filters.ACARS.DuplicateMessageSimilarity != 0.0 || config.Filters.VDLM2.DuplicateMessageSimilarity != 0.0 {
		enabledFilters = append(enabledFilters, "MessageSimilarity")
	}
	log.Infof("enabled filters: %s", strings.Join(enabledFilters, ","))
}
