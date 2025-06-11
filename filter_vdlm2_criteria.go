package main

import (
	"regexp"
)

type VDLM2CriteriaFilter struct {
}

func (a VDLM2CriteriaFilter) Name() string {
	return "vdlm2 criteria filter"
}

// All filters are defined here
var (
	VDLM2FilterFunctions = map[string]func(m VDLM2Message) bool{
		"HasText": func(m VDLM2Message) bool {
			re := regexp.MustCompile(`[\S]+`)
			return re.MatchString(m.VDL2.AVLC.ACARS.MessageText)
		},
		"MatchesTailCode": func(m VDLM2Message) bool {
			return config.Filters.Generic.TailCode == m.VDL2.AVLC.ACARS.Registration
		},
		"MatchesFlightNumber": func(m VDLM2Message) bool {
			return config.Filters.Generic.FlightNumber == m.VDL2.AVLC.ACARS.FlightNumber
		},
		"MatchesFrequency": func(m VDLM2Message) bool {
			return config.Filters.Generic.Frequency == float64(m.VDL2.FrequencyHz)
		},
		"MatchesStationID": func(m VDLM2Message) bool {
			return config.Filters.Generic.StationID == m.VDL2.AVLC.Source.Address ||
				config.Filters.Generic.StationID == m.VDL2.AVLC.Destination.Address
		},
		"AboveMinimumSignal": func(m VDLM2Message) bool {
			return config.Filters.Generic.AboveSignaldBm <= m.VDL2.SignalLevel
		},
		"BelowMaximumSignal": func(m VDLM2Message) bool {
			return config.Filters.Generic.BelowSignaldBm >= m.VDL2.SignalLevel
		},
		"FromTower": func(m VDLM2Message) bool {
			b, _ := regexp.Match("\\S+", []byte(m.VDL2.AVLC.ACARS.FlightNumber))
			return !b
		},
		"FromAircraft": func(m VDLM2Message) bool {
			b, _ := regexp.Match("\\S+", []byte(m.VDL2.AVLC.ACARS.FlightNumber))
			return b
		},
		"More": func(m VDLM2Message) bool {
			return !m.VDL2.AVLC.ACARS.More
		},
		"MessageSimilarity": func(m VDLM2Message) bool {
			return FilterDuplicateVDLM2(m)
		},
		"DictionaryPhraseLengthMinimum": func(m VDLM2Message) bool {
			return int64(config.Filters.Generic.DictionaryPhraseLengthMinimum) <= LongestDictionaryWordPhraseLength(m.VDL2.AVLC.ACARS.MessageText)
		},
		"OpenAIPromptFilter": func(m VDLM2Message) bool {
			return OpenAIFilter(m.VDL2.AVLC.ACARS.MessageText)
		},
		"OllamaPromptFilter": func(m VDLM2Message) bool {
			return OllamaFilter(m.VDL2.AVLC.ACARS.MessageText)
		},
	}
)

// Return true if a message passes a filter, false otherwise
func (f VDLM2CriteriaFilter) Filter(m VDLM2Message) (ok bool, failedFilters []string) {
	ok = true
	for _, filter := range enabledFilters {
		if !VDLM2FilterFunctions[filter](m) {
			ok = false
			db.Delete(&m)
			failedFilters = append(failedFilters, filter)
		}
	}
	return ok, failedFilters
}
