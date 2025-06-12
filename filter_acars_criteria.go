package main

import (
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

type ACARSCriteriaFilter struct {
}

func (a ACARSCriteriaFilter) Name() string {
	return "acars criteria filter"
}

// All filters are defined here
var (
	ACARSFilterFunctions = map[string]func(m ACARSMessage) bool{
		"HasText": func(m ACARSMessage) bool {
			re := regexp.MustCompile(`[\S]+`)
			return re.MatchString(m.MessageText)
		},
		"MatchesTailCode": func(m ACARSMessage) bool {
			return config.Filters.Generic.TailCode == m.AircraftTailCode
		},
		"MatchesFlightNumber": func(m ACARSMessage) bool {
			return config.Filters.Generic.FlightNumber == m.FlightNumber
		},
		"MatchesFrequency": func(m ACARSMessage) bool {
			return config.Filters.Generic.Frequency == m.FrequencyMHz
		},
		"MatchesStationID": func(m ACARSMessage) bool {
			return config.Filters.Generic.StationID == m.StationID
		},
		"AboveMinimumSignal": func(m ACARSMessage) bool {
			return config.Filters.Generic.AboveSignaldBm <= m.SignaldBm
		},
		"BelowMaximumSignal": func(m ACARSMessage) bool {
			return config.Filters.Generic.BelowSignaldBm >= m.SignaldBm
		},
		"ASSStatus": func(m ACARSMessage) bool {
			return config.Filters.Generic.ASSStatus == m.ASSStatus
		},
		"FromTower": func(m ACARSMessage) bool {
			b, _ := regexp.Match("\\S+", []byte(m.FlightNumber))
			return !b
		},
		"FromAircraft": func(m ACARSMessage) bool {
			b, _ := regexp.Match("\\S+", []byte(m.FlightNumber))
			return b
		},
		"More": func(m ACARSMessage) bool {
			return true
		},
		"MessageSimilarity": func(m ACARSMessage) bool {
			return FilterDuplicateACARS(m)
		},
		"DictionaryPhraseLengthMinimum": func(m ACARSMessage) bool {
			return int64(config.Filters.Generic.DictionaryPhraseLengthMinimum) <= LongestDictionaryWordPhraseLength(m.MessageText)
		},
		"OpenAIPromptFilter": func(m ACARSMessage) bool {
			return OpenAIFilter(m.MessageText)
		},
		"OllamaPromptFilter": func(m ACARSMessage) bool {
			return OllamaFilter(m.MessageText)
		},
	}
)

// Return true if a message passes a filter, false otherwise
func (f ACARSCriteriaFilter) Filter(m ACARSMessage) (ok bool, failedFilters []string) {
	ok = true
	for _, filter := range enabledFilters {
		if !ACARSFilterFunctions[filter](m) {
			ok = false
			log.Debug(
				yo.FYI("message ending in ").
					Hmm(Last20Characters(m.MessageText)).
					FYI("took ").
					Hmm(time.Since(m.CreatedAt).String()).
					FYI("to filter with %s after ingest", filter).FRFR())
			db.Delete(&m)
			failedFilters = append(failedFilters, filter)
		}
	}
	return ok, failedFilters
}
