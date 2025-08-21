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
	"github.com/tidwall/words"
)

type StringMetric interface {
	Compare(a, b string) float64
}

const (
	fieldWasEmpty = "%s field was empty"
)

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
	// Only process messages that were from a ground-based transmitter - determined by the presence (From aircraft) or lack of (From ground) a flight number.
	FromTower *bool `json:",omitempty" default:"true"`
	// Only process messages that were from an aircraft - determined by the presence (From aircraft) or lack of (From ground) a flight number.
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

func (a BuiltinFilter) Name() string {
	return reflect.TypeOf(a).Name()
}

func (f BuiltinFilter) Configured() bool {
	return !reflect.DeepEqual(f, BuiltinFilter{})
}

// Return true if a message passes a filter, false otherwise
func (f BuiltinFilter) Filter(m APMessage) (filtered bool, reason string, errs error) {
	nameStep := fmt.Sprintf("%s in step %d", f.Name(), m["StepNumber"])
	configuredFields := NonZeroFields(f)
	var reasons []string
	var filter bool
	for _, field := range configuredFields {
		// This is not a function but a setting, so we skip it
		if field == "FilterOnFailure" {
			continue
		}
		if _, ok := BuiltinFilterFunctions[field]; !ok {
			errs = errors.Join(errs, fmt.Errorf("%s: tried to call %s but it is not a built-in filter function", nameStep, field))
			filtered = filtered || f.FilterOnFailure
		} else {
			var err error
			var filterReason string
			filter, filterReason, err = BuiltinFilterFunctions[field](f, m)
			if err != nil {
				errs = errors.Join(errs, err)
			}
			if filter {
				if filterReason == "" {
					reasons = append(reasons, field)
				} else {
					reasons = append(reasons, fmt.Sprintf("%s: %s", field, filterReason))
				}
			}
		}
		filtered = filtered || filter
	}
	return filtered, strings.Join(reasons, ","), errs
}

var (
	BuiltinFilterFunctions = map[string]func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error){
		"HasText": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			re := regexp.MustCompile(`^\s*$`)
			mt := GetAPMessageCommonFieldAsString(m, "MessageText")
			// dereferencing this pointer to nil is "impossible"
			// because it must be set to be here.........
			filter = *f.HasText == re.MatchString(mt)
			return filter, reason, nil
		},
		"TailCode": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			tc := GetAPMessageCommonFieldAsString(m, "TailCode")
			tailCodeMatches := f.TailCode == tc
			return !tailCodeMatches, reason, nil
		},
		"FlightNumber": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			fn := GetAPMessageCommonFieldAsString(m, "flight_number")
			flightNumberMatches := f.FlightNumber == fn
			return !flightNumberMatches, reason, nil
		},
		"Frequency": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			fmhz := GetAPMessageCommonFieldAsFloat64(m, "FrequencyMhz")
			fhz := GetAPMessageCommonFieldAsFloat64(m, "frequency_hz")
			if fmhz == 0.0 && fhz == 0.0 {
				return false, "FrequencyMhz and frequency_hz were empty", nil
			}
			if fmhz != 0.0 && fhz == 0.0 {
				fmhz = fmhz * 1000000
			} else {
				// In case both are present, use Hz
				fmhz = 0
			}
			freq := fmhz + fhz
			frequencyMatches := f.Frequency == freq
			return !frequencyMatches, reason, nil
		},
		"StationID": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			sid := GetAPMessageCommonFieldAsString(m, "station_id")
			stationIDMatches := f.StationID == sid
			return !stationIDMatches, reason, nil
		},
		"AboveMinimumSignal": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			signalStrength := GetAPMessageCommonFieldAsFloat64(m, "signal_dbm")
			aboveSignalStrength := f.AboveSignaldBm >= signalStrength
			return !aboveSignalStrength, reason, nil
		},
		"BelowMaximumSignal": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			signalStrength := GetAPMessageCommonFieldAsFloat64(m, "signal_dbm")
			belowSignalStrength := f.AboveSignaldBm <= signalStrength
			return !belowSignalStrength, reason, nil
		},
		"ASSStatus": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			ass := GetAPMessageCommonFieldAsString(m, "ass_status")
			assMatches := f.ASSStatus == ass
			return !assMatches, reason, nil
		},
		"FromTower": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			fnum := GetAPMessageCommonFieldAsString(m, "flight_number")
			flightNumberEmpty, _ := regexp.Match("^\\s*$", []byte(fnum))
			FromTower := *f.FromTower == flightNumberEmpty
			return !FromTower, reason, nil
		},
		"FromAircraft": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			fnum := GetAPMessageCommonFieldAsString(m, "flight_number")
			flightNumberNotEmpty, _ := regexp.Match("\\S", []byte(fnum))
			FromAircraft := *f.FromAircraft == flightNumberNotEmpty
			return !FromAircraft, reason, nil
		},
		"More": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			hasMore := GetAPMessageCommonFieldAsBoolean(m, "more")
			return !hasMore, reason, nil
		},
		"AboveDistanceNm": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			field := "AircraftDistanceNm"
			distance := GetAPMessageCommonFieldAsFloat64(m, field)
			if distance == 0.0 {
				return false, fmt.Sprintf(fieldWasEmpty, field), nil
			}
			distanceAbove := distance >= f.AboveDistanceNm
			return !distanceAbove, reason, nil
		},
		"BelowDistanceNm": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			field := "AircraftDistanceNm"
			distance := GetAPMessageCommonFieldAsFloat64(m, field)
			if distance == 0.0 {
				return false, fmt.Sprintf(fieldWasEmpty, field), nil
			}
			distanceBelow := distance <= f.BelowDistanceNm
			return !distanceBelow, reason, nil
		},
		"AboveDistanceMi": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			field := "AircraftDistanceMi"
			distance := GetAPMessageCommonFieldAsFloat64(m, field)
			if distance == 0.0 {
				return false, fmt.Sprintf(fieldWasEmpty, field), nil
			}
			distanceAbove := distance >= f.AboveDistanceMi
			return !distanceAbove, reason, nil
		},
		"BelowDistanceMi": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			field := "AircraftDistanceMi"
			distance := GetAPMessageCommonFieldAsFloat64(m, field)
			if distance == 0.0 {
				return false, fmt.Sprintf(fieldWasEmpty, field), nil
			}
			distanceBelow := distance <= f.BelowDistanceMi
			return !distanceBelow, reason, nil
		},
		"PreviousMessageSimilarity": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			filter, reason, err = f.FilterSimilarAPMessage(m)
			return filter, reason, nil
		},
		"DictionaryPhraseLengthMinimum": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			field := "MessageText"
			mt := GetAPMessageCommonFieldAsString(m, field)
			if mt == "" {
				return false, fmt.Sprintf(fieldWasEmpty, field), nil
			}
			length, dictionaryReason := LongestDictionaryWordPhraseLength(mt)
			lengthSufficient := length >= int64(f.DictionaryPhraseLengthMinimum)
			reason = dictionaryReason
			return !lengthSufficient, reason, nil
		},
		"FreetextTermPresent": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			field := "MessageText"
			mt := GetAPMessageCommonFieldAsString(m, field)
			if mt == "" {
				return false, fmt.Sprintf(fieldWasEmpty, field), nil
			}
			freetextTermPresent := *f.FreetextTermPresent == FreetextTermPresent(mt)
			return !freetextTermPresent, reason, nil
		},
	}
)

// Compares the message to a list of "Freetext" terms that are more likely to
// be present in human-initiated communication.
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

// Compares current message to last MaximumLookBehind messages
// using Hamming distance. If similarity is greater than Similarity,
// filter out the message.
func (d BuiltinFilter) FilterSimilarAPMessage(m APMessage) (filter bool, reason string, err error) {
	nameStep := fmt.Sprintf("%s in step %d", d.Name(), m["StepNumber"])
	// Don't filter if 0 similarity or unset
	if d.PreviousMessageSimilarity.Similarity == 0.0 {
		return false, "similarity was 0.0", nil
	}
	field := "MessageText"
	mt := GetAPMessageCommonFieldAsString(m, field)
	if regexp.MustCompile(`^\s*$`).MatchString(mt) {
		return true, fmt.Sprintf(fieldWasEmpty, field), nil
	}
	if d.PreviousMessageSimilarity.MaximumLookBehind != 0 {
		defaultMaxLookbehind = d.PreviousMessageSimilarity.MaximumLookBehind
	}
	am, vm, msgs := []ACARSMessage{}, []VDLM2Message{}, []string{}
	db.Where(ACARSMessage{Processed: true}).
		Limit(defaultMaxLookbehind).
		Find(&am)
	db.Where(VDLM2Message{Processed: true}).
		Limit(defaultMaxLookbehind).
		Find(&am)
	for _, acm := range am {
		acmapm := FormatAsAPMessage(acm)
		acmts := GetAPMessageCommonFieldAsString(acmapm, field)
		if err != nil {
			return d.FilterOnFailure, "", err
		}
		if acmts != "" {
			msgs = append(msgs, acmts)
		}
	}
	for _, acm := range vm {
		acmapm := FormatAsAPMessage(acm)
		acmts := GetAPMessageCommonFieldAsString(acmapm, field)
		if err != nil {
			return d.FilterOnFailure, "", err
		}
		if acmts != "" {
			msgs = append(msgs, acmts)
		}
	}
	for _, mcmp := range msgs {
		similarity := strutil.Similarity(mt, mcmp, metrics.NewHamming())
		if similarity >= d.PreviousMessageSimilarity.Similarity {
			pctSimilar := int(similarity * 100)
			if len(mt) > len(mcmp) && d.PreviousMessageSimilarity.DontFilterIfLonger {
				log.Debug(Aside("%s: message is greater than %d%% similarity(%d%%) but not filtering due to DontFilterIfLonger",
					nameStep, pctSimilar, d.PreviousMessageSimilarity.Similarity*100))
				continue
			} else {
				// Message is too similar, filter it out
				filter = true

				reason = fmt.Sprintf("%s: message is %d%% similar to a previous message", nameStep, pctSimilar)
				break
			}
		}
	}
	return filter, reason, err
}

// Unoptomized asf
func LongestDictionaryWordPhraseLength(messageText string) (wc int64, reason string) {
	var consecutiveWordSlice, maxConsecutiveWordSlice []string
	wordSlice := strings.FieldsFunc(messageText, SplitOnCommonMessageDelimiters)
	for idx, word := range wordSlice {
		var found bool
		for _, dictWord := range words.Words {
			found = false
			if strings.EqualFold(word, dictWord) {
				consecutiveWordSlice = append(consecutiveWordSlice, word)
				found = true
				// We don't need to search for further matches
				break
			}
		}
		if !found || idx == len(wordSlice)-1 {
			if len(consecutiveWordSlice) > len(maxConsecutiveWordSlice) {
				maxConsecutiveWordSlice = consecutiveWordSlice
				consecutiveWordSlice = []string{}
			}
		}
	}

	wc = int64(len(maxConsecutiveWordSlice))
	reason = fmt.Sprintf("message had %d consecutive dictionary words in it", wc)
	if wc > 0 {
		reason = reason + fmt.Sprintf(". longest dictionary word phrase found: %s", strings.Join(maxConsecutiveWordSlice, " "))
	}
	return wc, reason
}

func SplitOnCommonMessageDelimiters(r rune) bool {
	return r == ' ' || r == ','
}

// NonZeroFields returns all struct field names that are non-zero.
func NonZeroFields(v interface{}) []string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	var fields []string
	for i := 0; i < val.NumField(); i++ {
		if !val.Field(i).IsZero() {
			fields = append(fields, typ.Field(i).Name)
		}
	}
	return fields
}
