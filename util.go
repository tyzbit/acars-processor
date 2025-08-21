package main

import (
	"fmt"
	"strings"
	"time"
)

type RetriableError struct {
	Err        error
	RetryAfter time.Duration
}

// Error returns error message and a Retry-After duration.
func (e *RetriableError) Error() string {
	return fmt.Sprintf("%s (retry after %v)", e.Err.Error(), e.RetryAfter)
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

// Returns just the last 20 characters with whitespace characters removed
// from both ends.
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
