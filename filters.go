package main

import (
	"strings"

	log "github.com/sirupsen/logrus"
	// English only
)

func ConfigureFilters() {
	// -------------------------------------------------------------------------
	// Add filters based on what's enabled
	if config.FilterCriteriaMatchTailCode != "" {
		enabledFilters = append(enabledFilters, "MatchesTailCode")
	}
	if config.FilterCriteriaHasText {
		enabledFilters = append(enabledFilters, "HasText")
	}
	if config.FilterCriteriaMatchFlightNumber != "" {
		enabledFilters = append(enabledFilters, "MatchesFlightNumber")
	}
	if config.FilterCriteriaMatchFrequency != 0.0 {
		enabledFilters = append(enabledFilters, "MatchesFrequency")
	}
	if config.FilterCriteriaMatchStationID != "" {
		enabledFilters = append(enabledFilters, "MatchesStationID")
	}
	if config.FilterCriteriaAboveSignaldBm != 0.0 {
		enabledFilters = append(enabledFilters, "AboveMinimumSignal")
	}
	if config.FilterCriteriaBelowSignaldBm != 0.0 {
		enabledFilters = append(enabledFilters, "BelowMaximumSignal")
	}
	if config.FilterCriteriaAboveDistanceNm != 0.0 {
		enabledFilters = append(enabledFilters, "AboveMinimumSignal")
	}
	if config.FilterCriteriaBelowDistanceNm != 0.0 {
		enabledFilters = append(enabledFilters, "BelowMaximumSignal")
	}
	if config.FilterCriteriaFromTower {
		enabledFilters = append(enabledFilters, "FromTower")
	}
	if config.FilterCriteriaFromAircraft {
		enabledFilters = append(enabledFilters, "FromAircraft")
	}
	if config.FilterCriteriaMatchASSStatus != "" {
		enabledFilters = append(enabledFilters, "MatchesASSStatus")
	}
	if config.FilterCriteriaMore {
		enabledFilters = append(enabledFilters, "More")
	}
	if config.FilterCriteriaEmergency {
		enabledFilters = append(enabledFilters, "Emergency")
	}
	if config.FilterCriteriaDictionaryPhraseLengthMinimum > 0 {
		enabledFilters = append(enabledFilters, "ConsecutiveDictionaryWordCount")
	}
	if config.OpenAIAPIKey != "" {
		enabledFilters = append(enabledFilters, "OpenAIPromptFilter")
	}
	if config.OllamaFilterURL != "" {
		enabledFilters = append(enabledFilters, "OllamaPromptFilter")
	}
	if config.FilterCriteriaACARSDuplicateMessageSimilarity != 0.0 || config.FilterCriteriaVDLM2DuplicateMessageSimilarity != 0.0 {
		enabledFilters = append(enabledFilters, "MessageSimilarity")
	}
	log.Infof("enabled filters: %s", strings.Join(enabledFilters, ","))
}
