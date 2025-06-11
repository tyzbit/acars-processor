package main

import (
	"fmt"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

// Shorthand for colorizing output
// Increases overall program rizz
type Rizz struct {
	DMs []DM
}

type DM struct {
	Color   color.Color
	Message string
}

// execute the chain
// no cap
func (yo Rizz) FRFR() (s string) {
	if len(yo.DMs) == 0 {
		log.Error("no messages to print cuh")
		return ""
	}
	for _, dm := range yo.DMs {
		if !config.ACARSProcessorSettings.ColorOutput {
			dm.Color.DisableColor()
		} else {
			dm.Color.EnableColor()
		}
		s = s + dm.Color.Sprint(dm.Message) + color.New(color.Reset).Sprint("")
	}
	return s
}

// add a color and string manually to the message slice
// always gotta end it with .FRFR()
// check out this mad drip
func (yo Rizz) GlowUp(dm DM) (r Rizz) {
	yo.DMs = append(yo.DMs, DM{dm.Color, dm.Message})
	return yo
}

// green
// always gotta end it with .FRFR()
// We did it, Reddit
func (yo Rizz) Bet(finna ...any) (r Rizz) {
	if len(finna) == 1 {
		// Simple message
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgGreen), finna[0].(string)})
	} else {
		// Message with formatting
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgGreen), fmt.Sprintf(finna[0].(string), finna[1:]...)})
	}
	return yo
}

// magenta
// always gotta end it with .FRFR()
// ngl, but
func (yo Rizz) FYI(finna ...any) (r Rizz) {
	if len(finna) == 1 {
		// Simple message
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgMagenta), finna[0].(string)})
	} else {
		// Message with formatting
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgMagenta), fmt.Sprintf(finna[0].(string), finna[1:]...)})
	}
	return yo
}

// cyan
// always gotta end it with .FRFR()
// yoooooo
func (yo Rizz) Hmm(finna ...any) (r Rizz) {
	if len(finna) == 1 {
		// Simple message
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgCyan), finna[0].(string)})
	} else {
		// Message with formatting
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgCyan), fmt.Sprintf(finna[0].(string), finna[1:]...)})
	}
	return yo
}

// yellow
// always gotta end it with .FRFR()
// cringe
func (yo Rizz) Uhh(finna ...any) (r Rizz) {
	if len(finna) == 1 {
		// Simple message
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgYellow), finna[0].(string)})
	} else {
		// Message with formatting
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgYellow), fmt.Sprintf(finna[0].(string), finna[1:]...)})
	}
	return yo
}

// grey
// always gotta end it with .FRFR()
// check it,
func (yo Rizz) INFODUMP(finna ...any) (r Rizz) {
	darkGrey := *color.RGB(90, 90, 90)
	if len(finna) == 1 {
		// Simple message
		yo.DMs = append(yo.DMs, DM{darkGrey, finna[0].(string)})
	} else {
		// Message with formatting
		yo.DMs = append(yo.DMs, DM{darkGrey, fmt.Sprintf(finna[0].(string), finna[1:]...)})
	}
	return yo
}

// bold+italic
// always gotta end it with .FRFR()
// check out the rizz fam
func (yo Rizz) BTW(finna ...any) (r Rizz) {
	bio := color.New(color.Bold, color.Italic)
	if len(finna) == 1 {
		// Simple message
		yo.DMs = append(yo.DMs, DM{*bio, finna[0].(string)})
	} else {
		// Message with formatting
		yo.DMs = append(yo.DMs, DM{*bio, fmt.Sprintf(finna[0].(string), finna[1:]...)})
	}
	return yo
}
