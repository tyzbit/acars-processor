package main

import "regexp"

type ACARSCriteriaFilter struct {
}

func (a ACARSCriteriaFilter) Name() string {
	return "CriteriaFilter"
}

// All filters are defined here
var (
	filterFunctions = map[string]func(m ACARSMessage) bool{
		"HasText": func(m ACARSMessage) bool {
			re := regexp.MustCompile("[\\S]+")
			return re.MatchString(m.MessageText)
		},
		"MatchesTailCode": func(m ACARSMessage) bool {
			return config.FilterCriteriaMatchTailCode == m.AircraftTailCode
		},
	}
)

// Return true if a message passes a filter, false otherwise
func (f ACARSCriteriaFilter) Filter(m ACARSMessage, inclusive bool) bool {
	fullMatch := false
	for _, filter := range enabledFilters {
		match := filterFunctions[filter](m)
		// Any filter must pass
		if !inclusive && match {
			return true
		}
		// Every filter must pass
		if inclusive && !match {
			return false
		}
	}
	return fullMatch
}
