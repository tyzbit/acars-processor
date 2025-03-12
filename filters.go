package main

import (
	"strings"

	log "github.com/sirupsen/logrus"
	// English only
)

// Enables filters based on configuration; filters are identified with a
// string like "MatchesTailCode" and filter providers can implement them.
func ConfigureFilters() {
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
	wordSlice := SplitByAny(messageText, " ,;'.")
	for idx, word := range wordSlice {
		var found bool
		for _, dictWord := range ACARSDictionary() {
			found = false
			if strings.EqualFold(word, dictWord) {
				consecutiveWordSlice = append(consecutiveWordSlice, word)
				found = true
				// We don't need to search for further matches
				break
			}
		}
		if !found || idx == len(wordSlice)-1 {
			if len(consecutiveWordSlice) >= len(maxConsecutiveWordSlice) {
				maxConsecutiveWordSlice = consecutiveWordSlice
				consecutiveWordSlice = []string{}
			}
		}
	}

	wc = int64(len(maxConsecutiveWordSlice))
	if wc > 0 {
		log.Debugf("longest dictionary word phrase found (%d words): %s", wc, strings.Join(maxConsecutiveWordSlice, " "))
	}
	return wc
}
