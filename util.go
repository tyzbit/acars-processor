package main

import (
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

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
	filePath = os.Getenv("HOME") + "/" + filePath
	// Read the content of the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Errorf("Error reading the file: %v", err)
			os.Exit(1)
		}
	}
	return content
}

func WriteFile(filePath string, contents []byte) {
	filePath = os.Getenv("HOME") + "/" + filePath
	err := os.WriteFile(filePath, contents, 0644)
	if err != nil {
		log.Error("Error writing file: %w", err)
	}
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
		return "Tower"
	}
	return "Aircraft"
}
