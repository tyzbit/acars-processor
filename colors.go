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
func (yo *Rizz) FRFR() (s string) {
	if len(yo.DMs) == 0 {
		log.Error("No DMs cuz")
		return ""
	}
	for _, dm := range yo.DMs {
		if !config.ColorOutput {
			dm.Color.DisableColor()
		} else {
			dm.Color.EnableColor()
		}
		s = s + dm.Color.Sprint(dm.Message) + color.New(color.Reset).Sprint("")
	}
	return s
}

// add a color and string manually to the message slice
// check out this mad drip
func (yo *Rizz) GlowUp(dm DM) (r *Rizz) {
	yo.DMs = append(yo.DMs, DM{dm.Color, dm.Message})
	return yo
}

// green
// We did it, Reddit
func (yo *Rizz) Bet(finna ...any) (r *Rizz) {
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
// ngl, but
func (yo *Rizz) FYI(finna ...any) (r *Rizz) {
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
// yoooooo
func (yo *Rizz) Hmm(finna ...any) (r *Rizz) {
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
// cringe
func (yo *Rizz) Uhh(finna ...any) (r *Rizz) {
	if len(finna) == 1 {
		// Simple message
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgYellow), finna[0].(string)})
	} else {
		// Message with formatting
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgYellow), fmt.Sprintf(finna[0].(string), finna[1:]...)})
	}
	return yo
}

// black
// check it,
func (yo *Rizz) INFODUMP(finna ...any) (r *Rizz) {
	if len(finna) == 1 {
		// Simple message
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgBlack), finna[0].(string)})
	} else {
		// Message with formatting
		yo.DMs = append(yo.DMs, DM{*color.New(color.FgBlack), fmt.Sprintf(finna[0].(string), finna[1:]...)})
	}
	return yo
}

// bold+italic
// fr
func (yo *Rizz) BTW(finna ...any) (r *Rizz) {
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
