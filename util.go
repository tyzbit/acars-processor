package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type RetriableError struct {
	Err        error
	RetryAfter time.Duration
}

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
			log.Error(yo().Uhh("Error reading the file: %v", err).FRFR())
			os.Exit(1)
		}
	}
	return content
}

func WriteFile(filePath string, contents []byte) {
	err := os.WriteFile(filePath, contents, 0644)
	if err != nil {
		log.Error(yo().Uhh("Error writing file: %w", err))
	}
}

// Saves a file and returns true if the file changed
func UpdateFile(filePath string, contents []byte) (changed bool) {
	file := ReadFile(filePath)
	err := os.WriteFile(filePath, contents, 0644)
	if err != nil {
		log.Error(yo().Uhh("Error writing file: %w", err))
	}
	return string(file) != string(contents)
}

func MergeMaps(m1, m2 map[string]any) map[string]any {
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
	if len(s) <= 20 {
		return s
	}
	return s[len(s)-20:]
}
