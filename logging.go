package main

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	resetColor = color.New(color.Reset).Sprint("")
	fgGreen    = color.New(color.FgGreen)
	fgMagenta  = color.New(color.FgMagenta)
	fgCyan     = color.New(color.FgCyan)
	fgYellow   = color.New(color.FgYellow)
	darkGrey   = color.RGB(90, 90, 90)
	bio        = color.New(color.Bold, color.Italic)
)

// Render n strings using c color, returning a string with color escape sequences
func ColorSprintf(c *color.Color, n ...any) (rs string) {
	if config.ACARSProcessorSettings.ColorOutput {
		c.EnableColor()
	}
	if len(n) == 1 {
		// Simple message
		return c.Sprint(n[0].(string)) + resetColor
	} else {
		// Message with formatting
		return c.Sprint(fmt.Sprintf(n[0].(string), n[1:]...)) + resetColor
	}
}

// green
func Success(s ...any) string {
	return ColorSprintf(fgGreen, s...)
}

// magenta
func Content(s ...any) string {
	return ColorSprintf(fgMagenta, s...)
}

// cyan
func Note(s ...any) string {
	return ColorSprintf(fgCyan, s...)
}

// yellow
func Attention(s ...any) string {
	return ColorSprintf(fgYellow, s...)
}

// grey
func Aside(s ...any) string {
	return ColorSprintf(darkGrey, s...)
}

// bold+italic
func Emphasised(s ...any) string {
	return ColorSprintf(bio, s...)
}

// Custom
func Custom(c *color.Color, s ...any) string {
	return ColorSprintf(c, s...)
}
