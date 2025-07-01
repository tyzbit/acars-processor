package filter

import (
	"regexp"

	"github.com/tyzbit/acars-processor/acarshub"
	. "github.com/tyzbit/acars-processor/config"
)

type VDLM2CriteriaFilter struct {
}

func (a VDLM2CriteriaFilter) Name() string {
	return "vdlm2 criteria filter"
}

// All filters are defined here
var (
	VDLM2FilterFunctions = map[string]func(m acarshub.VDLM2Message) bool{
		"HasText": func(m acarshub.VDLM2Message) bool {
			re := regexp.MustCompile(`[\S]+`)
			return re.MatchString(m.VDL2.AVLC.ACARS.MessageText)
		},
		"MatchesTailCode": func(m acarshub.VDLM2Message) bool {
			return Config.Filters.Generic.TailCode == m.VDL2.AVLC.ACARS.Registration
		},
		"MatchesFlightNumber": func(m acarshub.VDLM2Message) bool {
			return Config.Filters.Generic.FlightNumber == m.VDL2.AVLC.ACARS.FlightNumber
		},
		"MatchesFrequency": func(m acarshub.VDLM2Message) bool {
			return Config.Filters.Generic.Frequency == float64(m.VDL2.FrequencyHz)
		},
		"MatchesStationID": func(m acarshub.VDLM2Message) bool {
			return Config.Filters.Generic.StationID == m.VDL2.AVLC.Source.Address ||
				Config.Filters.Generic.StationID == m.VDL2.AVLC.Destination.Address
		},
		"AboveMinimumSignal": func(m acarshub.VDLM2Message) bool {
			return Config.Filters.Generic.AboveSignaldBm <= m.VDL2.SignalLevel
		},
		"BelowMaximumSignal": func(m acarshub.VDLM2Message) bool {
			return Config.Filters.Generic.BelowSignaldBm >= m.VDL2.SignalLevel
		},
		"FromTower": func(m acarshub.VDLM2Message) bool {
			b, _ := regexp.Match("\\S+", []byte(m.VDL2.AVLC.ACARS.FlightNumber))
			return !b
		},
		"FromAircraft": func(m acarshub.VDLM2Message) bool {
			b, _ := regexp.Match("\\S+", []byte(m.VDL2.AVLC.ACARS.FlightNumber))
			return b
		},
		"More": func(m acarshub.VDLM2Message) bool {
			return !m.VDL2.AVLC.ACARS.More
		},
		"MessageSimilarity": func(m acarshub.VDLM2Message) bool {
			return FilterDuplicateVDLM2(m)
		},
		"DictionaryPhraseLengthMinimum": func(m acarshub.VDLM2Message) bool {
			return int64(Config.Filters.Generic.DictionaryPhraseLengthMinimum) <= LongestDictionaryWordPhraseLength(m.VDL2.AVLC.ACARS.MessageText)
		},
		"OpenAIPromptFilter": func(m acarshub.VDLM2Message) bool {
			return OpenAIFilter(m.VDL2.AVLC.ACARS.MessageText)
		},
		"OllamaPromptFilter": func(m acarshub.VDLM2Message) bool {
			return OllamaFilter(m.VDL2.AVLC.ACARS.MessageText)
		},
	}
)

// Return true if a message passes a filter, false otherwise
func (f VDLM2CriteriaFilter) Filter(m acarshub.VDLM2Message) (ok bool, failedFilters []string) {
	ok = true
	for _, filter := range EnabledFilters {
		if !VDLM2FilterFunctions[filter](m) {
			ok = false
			failedFilters = append(failedFilters, filter)
		}
	}
	return ok, failedFilters
}
