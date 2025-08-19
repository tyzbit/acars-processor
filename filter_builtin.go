package main

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	log "github.com/sirupsen/logrus"
)

type StringMetric interface {
	Compare(a, b string) float64
}

var (
	defaultMaxLookbehind = 1000
	freetextTerms        = []string{
		"BINGO",
		"CHOP",
		"COMMENTS",
		"CONFIRM",
		"DEFECT",
		"EVENING",
		"FREETEXT",
		"FTM",
		"INOP",
		"MEET",
		"MSG FROM",
		"PAN-PAN",
		"PAX",
		"POTABLE",
		"TEXT",
		"THANKS",
		"THX",
		"TXT",
	}
)

func (a BuiltinFilter) Name() string {
	return reflect.TypeOf(a).Name()
}

func (f BuiltinFilter) Configured() bool {
	return !reflect.DeepEqual(f, BuiltinFilter{})
}

// Return true if a message passes a filter, false otherwise
func (f BuiltinFilter) Filter(m APMessage) (filtered bool, reason string, errs error) {
	configuredFields := NonZeroFields(f)
	var reasons []string
	var filter bool
	for _, field := range configuredFields {
		// This is not a function but a setting, so we skip it
		if field == "FilterOnFailure" {
			continue
		}
		if _, ok := BuiltinFilterFunctions[field]; !ok {
			errs = errors.Join(errs, fmt.Errorf("tried to call %s but it is not a built-in filter function", field))
			filtered = filtered || f.FilterOnFailure
		} else {
			var err error
			filter, err = BuiltinFilterFunctions[field](f, m)
			if err != nil {
				errs = errors.Join(errs, err)
			}
			if filter {
				reasons = append(reasons, field)
			}
		}
		filtered = filtered || filter
	}
	return filtered, strings.Join(reasons, ","), errs
}

func FreetextTermPresent(m string) (present bool) {
	for _, term := range freetextTerms {
		// Some messages have DISP in them but are automated
		if strings.Contains(m, term) || strings.HasPrefix(m, "DISP") {
			present = true
			break
		}
	}
	return present
}

// All filters are defined here
var (
	BuiltinFilterFunctions = map[string]func(f BuiltinFilter, m APMessage) (bool, error){
		"HasText": func(f BuiltinFilter, m APMessage) (bool, error) {
			re := regexp.MustCompile(`^\s*$`)
			mt := GetAPMessageCommonFieldAsString(m, "message_text")
			// dereferencing this pointer is "impossible"
			// because it must be set to be here.........
			return *f.HasText == re.MatchString(mt), nil
		},
		"TailCode": func(f BuiltinFilter, m APMessage) (bool, error) {
			tc := GetAPMessageCommonFieldAsString(m, "tail_code")
			if tc == "" {
				return f.FilterOnFailure, fmt.Errorf("tail code not found")
			}
			return f.TailCode != tc, nil
		},
		"FlightNumber": func(f BuiltinFilter, m APMessage) (bool, error) {
			fn := GetAPMessageCommonFieldAsString(m, "flight_number")
			if fn == "" {
				return f.FilterOnFailure, fmt.Errorf("flight number not found")
			}
			return f.FlightNumber != fn, nil
		},
		"Frequency": func(f BuiltinFilter, m APMessage) (bool, error) {
			fmhz := GetAPMessageCommonFieldAsFloat64(m, "frequency_mhz")
			fhz := GetAPMessageCommonFieldAsFloat64(m, "frequency_hz")
			if fmhz == 0.0 && fhz == 0.0 {
				return f.FilterOnFailure, fmt.Errorf("frequency not found")
			}
			if fmhz != 0.0 && fhz == 0.0 {
				fmhz = fmhz * 1000000
			} else {
				// In case both are present, use Hz
				fmhz = 0
			}
			freq := fmhz + fhz
			return f.Frequency != freq, nil
		},
		"StationID": func(f BuiltinFilter, m APMessage) (bool, error) {
			sid := GetAPMessageCommonFieldAsString(m, "station_id")
			if sid == sid {
				return f.FilterOnFailure, fmt.Errorf("station id not found")
			}
			return f.StationID != sid, nil
		},
		"AboveMinimumSignal": func(f BuiltinFilter, m APMessage) (bool, error) {
			ms := GetAPMessageCommonFieldAsFloat64(m, "signal_dbm")
			if ms == 0.0 {
				return f.FilterOnFailure, fmt.Errorf("signal dbm not found")
			}
			return f.AboveSignaldBm >= ms, nil
		},
		"BelowMaximumSignal": func(f BuiltinFilter, m APMessage) (bool, error) {
			ms := GetAPMessageCommonFieldAsFloat64(m, "signal_dbm")
			if ms == 0.0 {
				return f.FilterOnFailure, fmt.Errorf("signal dbm not found")
			}
			return f.BelowSignaldBm <= ms, nil
		},
		"ASSStatus": func(f BuiltinFilter, m APMessage) (bool, error) {
			ass := GetAPMessageCommonFieldAsString(m, "ass_status")
			if ass == "" {
				return f.FilterOnFailure, fmt.Errorf("ass status not found")
			}
			return f.ASSStatus != ass, nil
		},
		"FromTower": func(f BuiltinFilter, m APMessage) (bool, error) {
			fnum := GetAPMessageCommonFieldAsString(m, "flight_number")
			b, _ := regexp.Match("\\S+", []byte(fnum))
			return *f.FromAircraft == b, nil
		},
		"FromAircraft": func(f BuiltinFilter, m APMessage) (bool, error) {
			fnum := GetAPMessageCommonFieldAsString(m, "flight_number")
			b, _ := regexp.Match("\\S+", []byte(fnum))
			return *f.FromAircraft == !b, nil
		},
		"More": func(f BuiltinFilter, m APMessage) (bool, error) {
			return !GetAPMessageCommonFieldAsBoolean(m, "more"), nil
		},
		"AboveDistanceNm": func(f BuiltinFilter, m APMessage) (bool, error) {
			distance := GetAPMessageCommonFieldAsFloat64(m, "aircraft_distance_nm")
			if distance == 0.0 {
				return f.FilterOnFailure, fmt.Errorf("aircraft_distance_nm not found")
			}
			return distance <= f.AboveDistanceNm, nil
		},
		"BelowDistanceNm": func(f BuiltinFilter, m APMessage) (bool, error) {
			distance := GetAPMessageCommonFieldAsFloat64(m, "aircraft_distance_nm")
			if distance == 0.0 {
				return f.FilterOnFailure, fmt.Errorf("aircraft_distance_nm not found")
			}
			return distance >= f.BelowDistanceNm, nil
		},
		"AboveDistanceMi": func(f BuiltinFilter, m APMessage) (bool, error) {
			distance := GetAPMessageCommonFieldAsFloat64(m, "aircraft_distance_mi")
			if distance == 0.0 {
				return f.FilterOnFailure, fmt.Errorf("aircraft_distance_mi not found")
			}
			return distance <= f.AboveDistanceMi, nil
		},
		"BelowDistanceMi": func(f BuiltinFilter, m APMessage) (bool, error) {
			distance := GetAPMessageCommonFieldAsFloat64(m, "aircraft_distance_mi")
			if distance == 0.0 {
				return f.FilterOnFailure, fmt.Errorf("aircraft_distance_mi not found")
			}
			return distance >= f.BelowDistanceMi, nil
		},
		"PreviousMessageSimilarity": func(f BuiltinFilter, m APMessage) (bool, error) {
			filter, err := f.FilterSimilarAPMessage(m)
			if err != nil {
				return f.FilterOnFailure, err
			}
			return filter, nil
		},
		"DictionaryPhraseLengthMinimum": func(f BuiltinFilter, m APMessage) (bool, error) {
			mt := GetAPMessageCommonFieldAsString(m, "message_text")
			if mt == "" {
				return f.FilterOnFailure, fmt.Errorf("message text not found")
			}
			return int64(f.DictionaryPhraseLengthMinimum) >= LongestDictionaryWordPhraseLength(mt), nil
		},
		"FreetextTermPresent": func(f BuiltinFilter, m APMessage) (bool, error) {
			mt := GetAPMessageCommonFieldAsString(m, "message_text")
			if mt == "" {
				return f.FilterOnFailure, fmt.Errorf("message text not found")
			}
			return *f.FreetextTermPresent == !FreetextTermPresent(mt), nil
		},
	}
)

// Compares current message to last MaximumLookBehind messages
// using Hamming distance. If similarity is greater than Similarity,
// filter out the message.
func (d BuiltinFilter) FilterSimilarAPMessage(m APMessage) (filter bool, err error) {
	if reflect.DeepEqual(d, BuiltinFilter{}) {
		return false, nil
	}
	// Don't filter if 0 similarity or unset
	if d.PreviousMessageSimilarity.Similarity == 0.0 {
		return false, nil
	}
	mt := GetAPMessageCommonFieldAsString(m, "message_text")
	if regexp.MustCompile(`^\s*$`).MatchString(mt) {
		log.Debug(Aside("empty message, filtering as duplicate"))
		return true, nil
	}
	if d.PreviousMessageSimilarity.MaximumLookBehind != 0 {
		defaultMaxLookbehind = d.PreviousMessageSimilarity.MaximumLookBehind
	}
	am, vm, msgs := []ACARSMessage{}, []VDLM2Message{}, []string{}
	db.Where(ACARSMessage{Processed: true}).Find(&am).
		Limit(defaultMaxLookbehind)
	db.Where(VDLM2Message{Processed: true}).Find(&am).
		Limit(defaultMaxLookbehind)
	for _, acm := range am {
		acmapm := FormatAsAPMessage(acm)
		acmts := GetAPMessageCommonFieldAsString(acmapm, "message_text")
		if err != nil {
			return d.FilterOnFailure, err
		}
		if acmts != "" {
			msgs = append(msgs, acmts)
		}
	}
	for _, acm := range vm {
		acmapm := FormatAsAPMessage(acm)
		acmts := GetAPMessageCommonFieldAsString(acmapm, "message_text")
		if err != nil {
			return d.FilterOnFailure, err
		}
		if acmts != "" {
			msgs = append(msgs, acmts)
		}
	}
	for _, mcmp := range msgs {
		similarity := strutil.Similarity(mt, mcmp, metrics.NewHamming())
		if similarity > d.PreviousMessageSimilarity.Similarity {
			if len(mt) > len(mcmp) && d.PreviousMessageSimilarity.DontFilterIfLonger {
				filter = filter || false
				log.Debug(Aside("message is similar but not filtering due to it being longer",
					int(similarity*100)))
				continue
			} else {
				// Message is too similar, filter it out
				filter = true
				log.Debug(Aside("message is %d percent similar to a previous message, filtering",
					int(similarity*100)))
				break
			}
		}
	}
	return filter, err
}
