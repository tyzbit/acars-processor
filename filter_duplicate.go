package main

import (
	"regexp"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	log "github.com/sirupsen/logrus"
)

type StringMetric interface {
	Compare(a, b string) float64
}

var (
	RecentMessageMax    = 10000
	RecentACARSMessages []ACARSMessage
	RecentVDLM2Messages []VDLM2Message
)

// Compares message to a buffer of previous messages and compares
// using Hamming distance. If similarity above configured setting,
// filter out the message.
func FilterDuplicateACARS(m ACARSMessage) bool {
	// Don't filter if 0 similarity or unset
	if config.Filters.ACARS.DuplicateMessageSimilarity == 0.0 {
		return true
	}
	if regexp.MustCompile(`^\s*$`).MatchString(m.MessageText) {
		log.Debug("empty message, filtering as duplicate")
		return false
	}
	allowMessage := true
	for _, c := range RecentACARSMessages {
		similarity := strutil.Similarity(m.MessageText, c.MessageText, metrics.NewHamming())
		if similarity > config.Filters.ACARS.DuplicateMessageSimilarity {
			// Message is too similar, filter it out
			allowMessage = false
			log.Debugf("message is %d percent similar to a previous message, filtering",
				int(similarity*100))
			break
		}
	}
	if len(RecentACARSMessages) >= RecentMessageMax {
		// Remove the oldest message
		RecentACARSMessages = RecentACARSMessages[1:]
	}
	RecentACARSMessages = append(RecentACARSMessages, m)
	return allowMessage
}

// Compares message to a buffer of previous messages and compares
// using Hamming distance. If similarity above configured setting,
// filter out the message.
func FilterDuplicateVDLM2(m VDLM2Message) bool {
	// Don't filter if 0 similarity or unset
	if config.Filters.VDLM2.DuplicateMessageSimilarity == 0.0 {
		return true
	}
	if regexp.MustCompile(`^\s*$`).MatchString(m.VDL2.AVLC.ACARS.MessageText) {
		log.Debug("empty message, filtering as duplicate")
		return false
	}
	allowMessage := true
	for _, c := range RecentVDLM2Messages {
		similarity := strutil.Similarity(m.VDL2.AVLC.ACARS.MessageText, c.VDL2.AVLC.ACARS.MessageText, metrics.NewHamming())
		if similarity > config.Filters.VDLM2.DuplicateMessageSimilarity {
			// Message is too similar, filter it out
			allowMessage = false
			log.Debugf("message is %d percent similar to a previous message, filtering",
				int(similarity*100))
			break
		}
	}
	if len(RecentVDLM2Messages) >= RecentMessageMax {
		// Remove the oldest message
		RecentVDLM2Messages = RecentVDLM2Messages[1:]
	}
	RecentVDLM2Messages = append(RecentVDLM2Messages, m)
	return allowMessage
}
