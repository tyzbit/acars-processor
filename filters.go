package main

import (
	"errors"
	"fmt"
	// English only
)

type BuiltinFilter struct {
	Filterer
	// Whether or not to filter the message if the filter has an error
	FilterOnFailure bool `json:",omitempty" default:"false"`
	// Generic Filters
	//
	// Only process messages with text included.
	HasText *bool `json:",omitempty" default:"false"`
	// Only process messages that have this tail code.
	TailCode string `json:",omitempty" default:"N999AP"`
	// Only process messages that have this flight number.
	FlightNumber string `json:",omitempty" default:"N999AP"`
	// Only process messages that have ASS Status.
	ASSStatus string `json:",omitempty" default:"anything"`
	// Only process messages that were received above this signal strength (in dBm).
	AboveSignaldBm float64 `json:",omitempty" default:"-9.9"`
	// Only process messages that were received below this signal strength (in dBm).
	BelowSignaldBm float64 `json:",omitempty" default:"-9.9"`
	// Only process messages received on this frequency.
	Frequency float64 `json:",omitempty" default:"136.950"`
	// Only process messages with this station ID.
	StationID string `json:",omitempty" default:"N12346"`
	// Only process messages that were from a ground-based transmitter - determined by the presence (from aircraft) or lack of (from ground) a flight number.
	FromTower *bool `json:",omitempty" default:"true"`
	// Only process messages that were from an aircraft - determined by the presence (from aircraft) or lack of (from ground) a flight number.
	FromAircraft *bool `json:",omitempty" default:"true"`
	// Only process messages that have the "More" flag set.
	More *bool `json:",omitempty" default:"true"`
	// Only process messages that came from aircraft further than this many nautical miles away (requires ADS-B or tar1090).
	AboveDistanceNm float64 `json:",omitempty" default:"15.5"`
	// Only process messages that came from aircraft closer than this many nautical miles away (requires ADS-B or tar1090).
	BelowDistanceNm float64 `json:",omitempty" default:"15.5"`
	// Only process messages that came from aircraft further than this many miles away (requires ADS-B or tar1090).
	AboveDistanceMi float64 `json:",omitempty" default:"15.5"`
	// Only process messages that came from aircraft closer than this many miles away (requires ADS-B or tar1090).
	BelowDistanceMi float64 `json:",omitempty" default:"15.5"`
	// Only process messages that have the "Emergency" flag set.
	Emergency *bool `json:",omitempty" default:"true"`
	// Only process messages that have at least this many valid dictionary words in a row.
	DictionaryPhraseLengthMinimum int `json:",omitempty" default:"5"`
	// Only process messages that have common freetext terms in them
	FreetextTermPresent *bool `json:",omitempty" default:"false"`
	// Only process ACARS messages that are at least this percent (ex: 0.8 for 80 percent) different than any other message received.
	PreviousMessageSimilarity struct {
		Similarity         float64 `default:"0.9"`
		MaximumLookBehind  int     `default:"100"`
		DontFilterIfLonger bool    `default:"true"`
	}
}

// Filter
func (f FilterStep) Filter(m APMessage) (s string, filtered bool, errs error) {
	filters := []Filterer{
		f.Builtin,
		f.Ollama,
		f.OpenAI,
	}
	for _, filter := range filters {
		if !filter.Configured() {
			continue
		}
		filterResult, reason, err := filter.Filter(m)
		if err != nil {
			errs = errors.Join(errs, err)
		}
		filtered = filterResult || filtered
		s = fmt.Sprintf("%s(%s)", filter.Name(), reason)
		if filtered {
			break
		}
	}
	return s, filtered, errs
}
