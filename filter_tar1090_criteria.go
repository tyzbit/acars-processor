package main

type TAR1090CriteriaFilter struct {
}

func (a TAR1090CriteriaFilter) Name() string {
	return "tar1090 criteria filter"
}

// All filters are defined here
var (
	TAR1090FilterFunctions = map[string]func(tar Tar1090AircraftJSON) bool{
		"MatchesTailCode": func(tar Tar1090AircraftJSON) (result bool) {
			for _, aircraft := range tar.Aircraft {
				if config.FilterCriteriaMatchTailCode == NormalizeAircraftRegistration(aircraft.Registration) {
					result = true
					break
				}
			}
			return result
		},
		"MatchesFlightNumber": func(tar Tar1090AircraftJSON) (result bool) {
			for _, aircraft := range tar.Aircraft {
				if config.FilterCriteriaMatchFlightNumber == aircraft.AircraftTailCode {
					result = true
					break
				}
			}
			return result
		},
		"AboveMinimumSignal": func(tar Tar1090AircraftJSON) (result bool) {
			for _, aircraft := range tar.Aircraft {
				if config.FilterCriteriaAboveSignaldBm <= aircraft.RSSISignalPowerdBm {
					result = true
					break
				}
			}
			return result
		},
		"BelowMaximumSignal": func(tar Tar1090AircraftJSON) (result bool) {
			for _, aircraft := range tar.Aircraft {
				if config.FilterCriteriaBelowSignaldBm >= aircraft.RSSISignalPowerdBm {
					result = true
					break
				}
			}
			return result
		},
		"AboveMinimumDistance": func(tar Tar1090AircraftJSON) (result bool) {
			for _, aircraft := range tar.Aircraft {
				if config.FilterCriteriaAboveDistanceNm <= aircraft.DistanceFromReceiverNm {
					result = true
					break
				}
			}
			return result
		},
		"BelowMaximumDistance": func(tar Tar1090AircraftJSON) (result bool) {
			for _, aircraft := range tar.Aircraft {
				if config.FilterCriteriaBelowDistanceNm >= aircraft.DistanceFromReceiverNm {
					result = true
					break
				}
			}
			return result
		},
		"Emergency": func(tar Tar1090AircraftJSON) (result bool) {
			for _, aircraft := range tar.Aircraft {
				if aircraft.Emergency != "" {
					result = true
					break
				}
			}
			return result
		},
	}
)

// Return true if a message passes a filter, false otherwise
func (t Tar1090AircraftJSON) Filter(tar Tar1090AircraftJSON) (ok bool, failedFilters []string) {
	ok = true
	for _, filter := range enabledFilters {
		if !TAR1090FilterFunctions[filter](tar) {
			ok = false
			failedFilters = append(failedFilters, filter)
		}
	}
	return ok, failedFilters
}
