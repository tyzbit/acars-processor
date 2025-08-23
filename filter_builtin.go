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
	// Only process messages that have common freetext terms in them. This also looks for messages that start with DISP since just containing DISP is not effective for fiding non-automated messages.
	FreetextTermPresent *bool `json:",omitempty" default:"false"`
	// Only process ACARS messages that are at least this percent (ex: 0.8 for 80 percent) different than any other message received.
	PreviousMessageSimilarity struct {
		Similarity         float64 `default:"0.9"`
		MaximumLookBehind  int     `default:"100"`
		DontFilterIfLonger bool    `default:"true"`
	}
	// Compare one or more fields and filter if they do not match specification.
	Compare struct {
		// The message field that will be evaluated.
		MessageField string
		// Compare with a decimal. MessageField must be float64.
		WithDecimal float64
		// Compare with multiple decimals. MessageField must be float64.
		WithDecimals []float64
		// Compare with an integer. MessageField must be int.
		WithNumber int
		// Compare with an integer. MessageField must be int.
		WithNumbers []int
		// Compare with a string. MessageField must be string.
		WithString string
		// Compare with multiple strings. MessageField must be string.
		WithStrings []string
		// Compare with another message field. MessageField must be same type.
		WithOtherMessageField string
		// Compare with multiple other message fields decimals. MessageFields must be the same type.
		WithOtherMessageFields []string
		// How should MessageField be compared? Only one operator may be enabled at a time. If you need more, create another step.
		Operation struct {
			// MessageField must be greater than any OtherMessageFields
			GreaterThan bool
			// MessageField must be less than any OtherMessageFields
			LessThan bool
			// MessageField must be equal to any OtherMessageFields
			EqualTo bool
			// MessageField must be longer than any OtherMessageFields
			LongerThan bool
			// MessageField must be shorter than any OtherMessageFields
			ShorterThan bool
			// MessageField must contain (with regex) the value of any OtherMessageFields
			MustContainRegex bool
			// MessageField must not contain the value of any OtherMessageFields
			MustNotContain bool
			// MessageField must not contain (with regex) the value of any OtherMessageFields
			MustNotContainRegex bool
		}
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
	configuredFields := NonZeroFields(f)
	var reasons []string
	var filter bool
	for _, field := range configuredFields {
		// This is not a function but a setting, so we skip it
		if field == "FilterOnFailure" {
			continue
		}
		if _, ok := BuiltinFilterFunctions[field]; !ok {
			errs = errors.Join(errs, fmt.Errorf("%s: tried to call %s but it is not a built-in filter function", f.Name(), field))
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
			mt := GetAPMessageCommonFieldAsString(m, "MessageText")
			// dereferencing this pointer to nil is "impossible"
			// because it must be set to be here.........
			empty, _ := regexp.MatchString(emptyStringRegex, mt)
			return !empty, reason, nil
		},
		"TailCode": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			tc := GetAPMessageCommonFieldAsString(m, "TailCode")
			tailCodeMatches := f.TailCode == tc
			return !tailCodeMatches, reason, nil
		},
		"FlightNumber": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			fn := GetAPMessageCommonFieldAsString(m, "FlightNumber")
			flightNumberMatches := f.FlightNumber == fn
			return !flightNumberMatches, reason, nil
		},
		"Frequency": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			fmhz := GetAPMessageCommonFieldAsFloat64(m, "FrequencyMhz")
			fhz := GetAPMessageCommonFieldAsFloat64(m, "FrequencyHz")
			if fmhz == 0.0 && fhz == 0.0 {
				return false, "FrequencyMhz and FrequencyHz were empty", nil
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
			stationIDMatches := f.StationID == GetAPMessageCommonFieldAsString(m, "StationId")
			return !stationIDMatches, reason, nil
		},
		"AboveMinimumSignal": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			aboveSignalStrength := f.AboveSignaldBm >= GetAPMessageCommonFieldAsFloat64(m, "SignaldBm")
			return !aboveSignalStrength, reason, nil
		},
		"BelowMaximumSignal": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			belowSignalStrength := f.AboveSignaldBm <= GetAPMessageCommonFieldAsFloat64(m, "SignaldBm")
			return !belowSignalStrength, reason, nil
		},
		"ASSStatus": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			ass := GetAPMessageCommonFieldAsString(m, "ASSStatus")
			assMatches := f.ASSStatus == ass
			return !assMatches, reason, nil
		},
		"FromTower": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			fnum := GetAPMessageCommonFieldAsString(m, "FlightNumber")
			flightNumberEmpty, _ := regexp.MatchString(emptyStringRegex, fnum)
			FromTower := *f.FromTower == flightNumberEmpty
			return !FromTower, reason, nil
		},
		"FromAircraft": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			fnum := GetAPMessageCommonFieldAsString(m, "FlightNumber")
			flightNumberNotEmpty, _ := regexp.MatchString(nonEmptyStringRegex, fnum)
			FromAircraft := *f.FromAircraft == flightNumberNotEmpty
			return !FromAircraft, reason, nil
		},
		"More": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			hasMore := GetAPMessageCommonFieldAsBoolean(m, "More")
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
		"Compare": func(f BuiltinFilter, m APMessage) (filter bool, reason string, err error) {
			if f.Compare.MessageField == "" {
				return f.FilterOnFailure, "MessageField was empty", nil
			}
			var multi bool
			if len(f.Compare.WithOtherMessageFields) > 0 {
				multi = true
			}

			was := map[bool]string {
				true: "was",
				false: "was not",
			}
			c := f.Compare
			field := c.MessageField
			fieldToCompare := c.WithOtherMessageField
			fieldsToCompare := c.WithOtherMessageFields
			op := c.Operation
			var fieldIsInt, fieldIsFloat, fieldIsString = bool, bool, bool
			_, fieldIsInt := field.(int)
			_, fieldIsFloat := field.(float64)
			_, fieldIsString := field.(string)
			var otherFieldIsInt, otherFieldIsFloat, otherFieldIsString = bool, bool, bool
			_, otherFieldIsInt := fieldToCompare.(int)
			_, otherFieldIsFloat := fieldToCompare.(float64)
			_, otherFieldIsString := fieldToCompare.(string)
			var compareOperations []string
			if fieldsToCompare {
				compareOperations = fieldsToCompare
			} else {
				compareOperations = []string{fieldToCompare}
			}
			for i, cmp := range compareOperations {
				switch {
				case op.GreaterThan:
					if fieldIsInt && otherFieldIsInt {
						return !GetAPMessageCommonFieldAsInt(field)>GetAPMessageCommonFieldAsInt(cmp), 
							fmt.Sprintf("%s %s matched as greater int than %s", field, was[a>b], cmp), nil
					}
					if fieldIsFloat && otherFieldIsFloat {
						return !GetAPMessageCommonFieldAsFloat64(field)>GetAPMessageCommonFieldAsFloat64(cmp), 
							fmt.Sprintf("%s %s matched as greater decimal than %s", field, was[a>b], cmp), nil
					}
					if fieldIsString && otherFieldIsString {
						return !GetAPMessageCommonFieldAsString(field)>GetAPMessageCommonFieldAsString(cmp), 
							fmt.Sprintf("%s %s matched as greater string than %s", field, was[a>b], cmp), nil
					}
				case op.LessThan:
					return "LessThan"
				case op.EqualTo:
					return "EqualTo"
				case op.LongerThan:
					return "LongerThan"
				case op.ShorterThan:
					return "ShorterThan"
				case op.MustContainRegex:
					return "MustContainRegex"
				case op.MustNotContain:
					return "MustNotContain"
				case op.MustNotContainRegex:
					return "MustNotContainRegex"
				default:
					return "" // none set
				}
				return filter, reason, err
			}
			return adsasdasda sds
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
	// Don't filter if 0 similarity or unset
	if d.PreviousMessageSimilarity.Similarity == 0.0 {
		return false, "similarity was 0.0", nil
	}
	field := "MessageText"
	mt := GetAPMessageCommonFieldAsString(m, field)
	if match, _ := regexp.MatchString(emptyStringRegex, mt); match {
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
		Find(&vm)
	for _, acm := range am {
		acmapm := FormatAsAPMessage(acm, "")
		acmts := GetAPMessageCommonFieldAsString(acmapm, field)
		if acmts != "" {
			msgs = append(msgs, acmts)
		}
	}
	for _, acm := range vm {
		acmapm := FormatAsAPMessage(acm, "")
		acmts := GetAPMessageCommonFieldAsString(acmapm, field)
		if acmts != "" {
			msgs = append(msgs, acmts)
		}
	}
	for _, mcmp := range msgs {
		similarity := strutil.Similarity(mt, mcmp, metrics.NewHamming())
		if similarity >= d.PreviousMessageSimilarity.Similarity {
			pctSimilar := int(similarity * 100)
			if len(mt) > len(mcmp) && d.PreviousMessageSimilarity.DontFilterIfLonger {
				log.Debug(Aside("%s: message is greater than %d%% similarity(%.2f%%) but not filtering due to DontFilterIfLonger",
					d.Name(), pctSimilar, d.PreviousMessageSimilarity.Similarity*100))
			} else {
				// Message is too similar, filter it out
				filter = true
				reason = fmt.Sprintf("%s: message is %d%% similar to a previous message", d.Name(), pctSimilar)
			}
			break
		}
	}
	return filter, reason, err
}

// Reads the string and finds the longest unbroken chain of dictionary words.
func LongestDictionaryWordPhraseLength(messageText string) (wc int64, reason string) {
	var consecutiveWordSlice, maxConsecutiveWordSlice []string
	// Split on space and comma as those are both used in real messages.
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
		// This follows the string above so the period is intentional.
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
