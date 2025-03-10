package main

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/words" // English only
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
}

// Unoptomized asf
func LongestDictionaryWordPhraseLength(messageText string) (wc int64) {
	var consecutiveWordSlice, maxConsecutiveWordSlice []string
	wordSlice := strings.Split(messageText, " ")
	for _, word := range wordSlice {
		for _, dictWord := range words.Words {
			if strings.EqualFold(word, dictWord) {
				consecutiveWordSlice = append(consecutiveWordSlice, word)
			} else {
				if len(maxConsecutiveWordSlice) < len(consecutiveWordSlice) {
					maxConsecutiveWordSlice = consecutiveWordSlice
				}
			}
		}
	}

	wc = int64(len(maxConsecutiveWordSlice))
	log.Debugf("message had %d consecutive dictionary words in it", wc)
	if wc > 0 {
		log.Debugf("longest dictionary word phrase found: %s", strings.Join(maxConsecutiveWordSlice, ","))
	}
	return wc
}
