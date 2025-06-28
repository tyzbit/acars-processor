package main

import (
	"fmt"

	"github.com/fatih/color"
)

// Render n strings using c color, returning a string with color escape sequences
func ColorSprintf(c color.Color, n ...any) (rs string) {
	if len(n) == 1 {
		// Simple message
		return c.Sprint(n[0].(string)) + color.New(color.Reset).Sprint("")
	} else {
		// Message with formatting
		return c.Sprint(fmt.Sprintf(n[0].(string), n[1:]...)) +
			color.New(color.Reset).Sprint("")
	}
}

// green
func Success(s ...any) string {
	return ColorSprintf(*color.New(color.FgGreen), s...)
}

// magenta
func Content(s ...any) string {
	return ColorSprintf(*color.New(color.FgMagenta), s...)
}

// cyan
func Note(s ...any) string {
	return ColorSprintf(*color.New(color.FgCyan), s...)
}

// yellow
func Attention(s ...any) string {
	return ColorSprintf(*color.New(color.FgYellow), s...)
}

// grey
func Aside(s ...any) string {
	darkGrey := *color.RGB(90, 90, 90)
	return ColorSprintf(darkGrey, s...)
}

// bold+italic
func Emphasised(s ...any) string {
	bio := color.New(color.Bold, color.Italic)
	return ColorSprintf(*bio, s...)
}

// Custom
func Custom(c color.Color, s ...any) string {
	return ColorSprintf(c, s...)
}
