package main

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/words"
)

type RetriableError struct {
	Err        error
	RetryAfter time.Duration
}

var (
	formatFilterAction = map[bool]string{
		true:  Custom(*color.New(color.FgYellow), "filtered"),
		false: Custom(*color.New(color.FgCyan), "allowed"),
	}
)

// Error returns error message and a Retry-After duration.
func (e *RetriableError) Error() string {
	return fmt.Sprintf("%s (retry after %v)", e.Err.Error(), e.RetryAfter)
}

func NormalizeAircraftRegistration(reg string) string {
	s := []string{
		".",
		" ",
		"-",
	}
	for _, r := range s {
		reg = strings.ReplaceAll(reg, r, "")
	}
	return strings.ToLower(reg)
}

func ReadFile(filePath string) []byte {
	// Read the content of the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Error(Attention("Error reading the file: %v", err))
			os.Exit(1)
		}
	}
	return content
}

func WriteFile(filePath string, contents []byte) {
	err := os.WriteFile(filePath, contents, 0644)
	if err != nil {
		log.Error(Attention("Error writing file: %s", err))
	}
}

// Saves a file and returns true if the file changed
func UpdateFile(filePath string, contents []byte) (changed bool) {
	file := ReadFile(filePath)
	err := os.WriteFile(filePath, contents, 0644)
	if err != nil {
		log.Error(Attention("Error writing file: %s", err))
	}
	return string(file) != string(contents)
}

func MergeAPMessages(m1, m2 map[string]any) APMessage {
	// Create a new map to avoid modifying the original maps.
	merged := make(map[string]any)

	// Copy m1 into merged.
	for k, v := range m1 {
		merged[k] = v
	}

	// Copy m2 into merged. If keys overlap, m2's values will overwrite m1's.
	for k, v := range m2 {
		merged[k] = v
	}

	return merged
}

// Returns if the string is empty or if it only contains nonprintable characters
func AircraftOrTower(s string) (r string) {
	b, _ := regexp.Match("\\S+", []byte(s))
	if b {
		return "Aircraft"
	}
	return "Tower"
}

// Fixes AI output bullshit
func SanitizeJSONString(s string) string {
	replacer := strings.NewReplacer(
		"“", "\"",
		"”", "\"",
		"‘", "'",
		"’", "'",
	)
	return replacer.Replace(s)
}

func Last20Characters(s string) string {
	// Remove newlines and trim leading spaces
	s = strings.ReplaceAll(s, "\n", " ")
	whitespaceCharacters := " \t\n\r\f\v"
	s = strings.TrimLeft(s, whitespaceCharacters)
	s = strings.TrimRight(s, whitespaceCharacters)
	if len(s) <= 20 {
		return s
	}
	return s[len(s)-20:]
}

// Unoptomized asf
func LongestDictionaryWordPhraseLength(messageText string) (wc int64) {
	var consecutiveWordSlice, maxConsecutiveWordSlice []string
	wordSlice := strings.Split(messageText, " ")
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
			if len(consecutiveWordSlice) >= len(maxConsecutiveWordSlice) {
				maxConsecutiveWordSlice = consecutiveWordSlice
				consecutiveWordSlice = []string{}
			}
		}
	}

	wc = int64(len(maxConsecutiveWordSlice))
	log.Debug(Aside("message had %d consecutive dictionary words in it", wc))
	if wc > 0 {
		log.Debug(Aside("longest dictionary word phrase found: %s", strings.Join(maxConsecutiveWordSlice, " ")))
	}
	return wc
}

// FormatAsAPMessage converts ACARS/VDLM2 messages into a flat object with dot-separated keys
func FormatAsAPMessage(s any) APMessage {
	result := make(map[string]any)
	reflectValue(reflect.TypeOf(s).Name(), reflect.ValueOf(s), result)
	return result
}

// reflectValue recursively processes a value (struct, map, etc.) and populates the map
func reflectValue(prefix string, v reflect.Value, m map[string]any) {
	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		// Process struct fields
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			tagName := field.Name
			newPrefix := prefix
			if newPrefix != "" {
				newPrefix += "."
			}
			newPrefix += tagName

			fieldVal := v.Field(i)
			reflectValue(newPrefix, fieldVal, m)

			// If acars tag exists, also add an entry with that key
			if acarsTag, ok := field.Tag.Lookup("ap"); ok && acarsTag != "" {
				// Use the raw field value
				if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
					m[acarsTag] = nil
				} else {
					m[acarsTag] = fieldVal.Interface()
				}
			}
		}

	case reflect.Map:
		// Process map entries
		for _, key := range v.MapKeys() {
			keyStr := key.String()
			newPrefix := prefix
			if newPrefix != "" {
				newPrefix += "."
			}
			newPrefix += keyStr
			reflectValue(newPrefix, v.MapIndex(key), m)
		}

	case reflect.Slice, reflect.Array:
		// Handle slices/arrays
		for i := 0; i < v.Len(); i++ {
			newPrefix := prefix
			if newPrefix != "" {
				newPrefix += "."
			}
			newPrefix += fmt.Sprintf("[%d]", i)
			reflectValue(newPrefix, v.Index(i), m)
		}

	default:
		// Add primitive types directly
		m[prefix] = v.Interface()
	}
}

func GetAPMessageCommonFieldAsString(a APMessage, s string) string {
	if val, ok := a[s]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func GetAPMessageCommonFieldAsInt(a APMessage, s string) int {
	if val, ok := a[s]; ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0.0
}

func GetAPMessageCommonFieldAsInt64(a APMessage, s string) int64 {
	if val, ok := a[s]; ok {
		if i, ok := val.(int64); ok {
			return i
		}
	}
	return 0.0
}

func GetAPMessageCommonFieldAsFloat64(a APMessage, s string) float64 {
	if val, ok := a[s]; ok {
		if flt, ok := val.(float64); ok {
			return flt
		}
	}
	return 0.0
}

func GetAPMessageCommonFieldAsBoolean(a APMessage, s string) bool {
	if val, ok := a[s]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
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
