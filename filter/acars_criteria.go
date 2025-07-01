package filter

import (
	"regexp"

	"github.com/tyzbit/acars-processor/acarshub"
	. "github.com/tyzbit/acars-processor/config"
)

type ACARSCriteriaFilter struct {
}

func (a ACARSCriteriaFilter) Name() string {
	return "acars criteria filter"
}

// All filters are defined here
var (
	ACARSFilterFunctions = map[string]func(m acarshub.ACARSMessage) bool{
		"HasText": func(m acarshub.ACARSMessage) bool {
			re := regexp.MustCompile(`[\S]+`)
			return re.MatchString(m.MessageText)
		},
		"MatchesTailCode": func(m acarshub.ACARSMessage) bool {
			return Config.Filters.Generic.TailCode == m.AircraftTailCode
		},
		"MatchesFlightNumber": func(m acarshub.ACARSMessage) bool {
			return Config.Filters.Generic.FlightNumber == m.FlightNumber
		},
		"MatchesFrequency": func(m acarshub.ACARSMessage) bool {
			return Config.Filters.Generic.Frequency == m.FrequencyMHz
		},
		"MatchesStationID": func(m acarshub.ACARSMessage) bool {
			return Config.Filters.Generic.StationID == m.StationID
		},
		"AboveMinimumSignal": func(m acarshub.ACARSMessage) bool {
			return Config.Filters.Generic.AboveSignaldBm <= m.SignaldBm
		},
		"BelowMaximumSignal": func(m acarshub.ACARSMessage) bool {
			return Config.Filters.Generic.BelowSignaldBm >= m.SignaldBm
		},
		"ASSStatus": func(m acarshub.ACARSMessage) bool {
			return Config.Filters.Generic.ASSStatus == m.ASSStatus
		},
		"FromTower": func(m acarshub.ACARSMessage) bool {
			b, _ := regexp.Match("\\S+", []byte(m.FlightNumber))
			return !b
		},
		"FromAircraft": func(m acarshub.ACARSMessage) bool {
			b, _ := regexp.Match("\\S+", []byte(m.FlightNumber))
			return b
		},
		"More": func(m acarshub.ACARSMessage) bool {
			return true
		},
		"MessageSimilarity": func(m acarshub.ACARSMessage) bool {
			return FilterDuplicateACARS(m)
		},
		"DictionaryPhraseLengthMinimum": func(m acarshub.ACARSMessage) bool {
			return int64(Config.Filters.Generic.DictionaryPhraseLengthMinimum) <= LongestDictionaryWordPhraseLength(m.MessageText)
		},
		"OpenAIPromptFilter": func(m acarshub.ACARSMessage) bool {
			return OpenAIFilter(m.MessageText)
		},
		"OllamaPromptFilter": func(m acarshub.ACARSMessage) bool {
			return OllamaFilter(m.MessageText)
		},
	}
)

// Return true if a message passes a filter, false otherwise
func (f ACARSCriteriaFilter) Filter(m acarshub.ACARSMessage) (ok bool, failedFilters []string) {
	ok = true
	for _, filter := range EnabledFilters {
		if !ACARSFilterFunctions[filter](m) {
			ok = false
			failedFilters = append(failedFilters, filter)
		}
	}
	return ok, failedFilters
}
