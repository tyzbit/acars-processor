package main

import (
	"regexp"
)

type VDLM2CriteriaFilter struct {
}

func (a VDLM2CriteriaFilter) Name() string {
	return "CriteriaFilter"
}

// All filters are defined here
var (
	VDLM2FilterFunctions = map[string]func(m VDLM2Message) bool{
		"HasText": func(m VDLM2Message) bool {
			re := regexp.MustCompile(`[\S]+`)
			return re.MatchString(m.VDL2.AVLC.ACARS.MessageText)
		},
		"MatchesTailCode": func(m VDLM2Message) bool {
			return config.FilterCriteriaMatchTailCode == m.VDL2.AVLC.ACARS.Registration
		},
		"MatchesFlightNumber": func(m VDLM2Message) bool {
			return config.FilterCriteriaMatchFlightNumber == m.VDL2.AVLC.ACARS.FlightNumber
		},
		"MatchesFrequency": func(m VDLM2Message) bool {
			return config.FilterCriteriaMatchFrequency == float64(m.VDL2.FrequencyHz)
		},
		"MatchesStationID": func(m VDLM2Message) bool {
			return config.FilterCriteriaMatchStationID == m.VDL2.AVLC.Source.Address ||
				config.FilterCriteriaMatchStationID == m.VDL2.AVLC.Destination.Address
		},
		"AboveMinimumSignal": func(m VDLM2Message) bool {
			return config.FilterCriteriaAboveSignaldBm < m.VDL2.SignalLevel
		},
		"BelowMaximumSignal": func(m VDLM2Message) bool {
			return config.FilterCriteriaAboveSignaldBm > m.VDL2.SignalLevel
		},
	}
)

// Return true if a message passes a filter, false otherwise
func (f VDLM2CriteriaFilter) Filter(m VDLM2Message) (ok bool, failedFilters []string) {
	ok = true
	for _, filter := range enabledFilters {
		if !VDLM2FilterFunctions[filter](m) {
			ok = false
			failedFilters = append(failedFilters, filter)
		}
	}
	return ok, failedFilters
}
