package main

import (
	"fmt"
	"reflect"
)

var ACARSProcessorPrefix = "ACARSProcessor."

// FormatAsAPMessage converts ACARS/VDLM2 messages into a flat object with
// dot-separated keys. Blank prefix uses the struct name.
func FormatAsAPMessage(s any, prefix string) APMessage {
	if prefix == "" {
		prefix = reflect.TypeOf(s).Name()
	}
	return squashValueIntoMapStringAny(prefix, reflect.ValueOf(s))
}

// reflectValue recursively processes a value (struct, map, etc.) and populates
// the map with string keys and any value
func squashValueIntoMapStringAny(prefix string, v reflect.Value) APMessage {
	m := APMessage{}
	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return m
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
			m = MergeAPMessages(m, squashValueIntoMapStringAny(newPrefix, fieldVal))

			// If acars tag exists, also add an entry with that key
			if acarsTag, ok := field.Tag.Lookup("ap"); ok && acarsTag != "" {
				m = MergeAPMessages(m, squashValueIntoMapStringAny(ACARSProcessorPrefix+acarsTag, fieldVal))
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
			m = MergeAPMessages(m, squashValueIntoMapStringAny(newPrefix, v.MapIndex(key)))
		}

	case reflect.Slice, reflect.Array:
		// Handle slices/arrays
		for i := 0; i < v.Len(); i++ {
			newPrefix := prefix
			if newPrefix != "" {
				newPrefix += "."
			}
			// Append index numbers to the field names
			newPrefix += fmt.Sprintf("[%d]", i)
			m = MergeAPMessages(m, squashValueIntoMapStringAny(newPrefix, v.Index(i)))
		}

	case reflect.Int:
		m = MergeAPMessages(m, APMessage{prefix: v.Int()})

	case reflect.Int64:
		m = MergeAPMessages(m, APMessage{prefix: v.Int()})

	case reflect.Uint:
		m = MergeAPMessages(m, APMessage{prefix: v.Uint()})

	case reflect.String:
		m = MergeAPMessages(m, APMessage{prefix: v.String()})

	case reflect.Float64:
		m = MergeAPMessages(m, APMessage{prefix: v.Float()})

	case reflect.Bool:
		m = MergeAPMessages(m, APMessage{prefix: v.Bool()})

	default:
		m = MergeAPMessages(m, APMessage{prefix: fmt.Sprintf("%s", v)})
	}
	return m
}

func MergeAPMessages(m1, m2 APMessage) APMessage {
	for k, v := range m2 {
		m1[k] = v
	}
	return m1
}

// Callers should know if the type is not what's assumed, zero
// value for that type will be returned.
func GetAPMessageCommonFieldAsString(a APMessage, s string) string {
	if val, ok := a[ACARSProcessorPrefix+s]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// Callers should know if the type is not what's assumed, zero
// value for that type will be returned.
func GetAPMessageCommonFieldAsInt(a APMessage, s string) int {
	if val, ok := a[ACARSProcessorPrefix+s]; ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

// Callers should know if the type is not what's assumed, zero
// value for that type will be returned.
func GetAPMessageCommonFieldAsInt64(a APMessage, s string) int64 {
	if val, ok := a[ACARSProcessorPrefix+s]; ok {
		if i, ok := val.(int64); ok {
			return i
		}
	}
	return 0
}

// Callers should know if the type is not what's assumed, zero
// value for that type will be returned.
func GetAPMessageCommonFieldAsFloat64(a APMessage, s string) float64 {
	if val, ok := a[ACARSProcessorPrefix+s]; ok {
		if flt, ok := val.(float64); ok {
			return flt
		}
	}
	return 0.0
}

// Callers should know if the type is not what's assumed, zero
// value for that type will be returned.
func GetAPMessageCommonFieldAsBoolean(a APMessage, s string) bool {
	if val, ok := a[ACARSProcessorPrefix+s]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
