package main

import (
	"fmt"
	"reflect"
)

var ACARSProcessorPrefix = "ACARSProcessor."

// FormatAsAPMessage converts ACARS/VDLM2 messages into a flat object with
// dot-separated keys
func FormatAsAPMessage(s any) APMessage {
	result := make(map[string]any)
	reflectValue(reflect.TypeOf(s).Name(), reflect.ValueOf(s), result)
	return result
}

// reflectValue recursively processes a value (struct, map, etc.) and populates
// the map with string keys and any value
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
					m["ACARSProcessor."+acarsTag] = nil
				} else {
					m["ACARSProcessor."+acarsTag] = fieldVal.Interface()
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

func GetAPMessageCommonFieldAsString(a APMessage, s string) string {
	if val, ok := a["ACARSProcessor."+s]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func GetAPMessageCommonFieldAsInt(a APMessage, s string) int {
	if val, ok := a["ACARSProcessor."+s]; ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0.0
}

func GetAPMessageCommonFieldAsInt64(a APMessage, s string) int64 {
	if val, ok := a["ACARSProcessor."+s]; ok {
		if i, ok := val.(int64); ok {
			return i
		}
	}
	return 0.0
}

func GetAPMessageCommonFieldAsFloat64(a APMessage, s string) float64 {
	if val, ok := a["ACARSProcessor."+s]; ok {
		if flt, ok := val.(float64); ok {
			return flt
		}
	}
	return 0.0
}

func GetAPMessageCommonFieldAsBoolean(a APMessage, s string) bool {
	if val, ok := a["ACARSProcessor."+s]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
