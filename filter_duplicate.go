package main

import (
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

func FilterDuplicateACARS(m ACARSMessage) bool {
	allowMessage := true
	if len(RecentACARSMessages) < RecentMessageMax {
		RecentACARSMessages = append(RecentACARSMessages, m)
	}
	for _, c := range RecentACARSMessages {
		similarity := strutil.Similarity(m.MessageText, c.MessageText, metrics.NewHamming())
		if similarity > config.FilterCriteriaACARSDuplicateMessageSimilarity {
			// Message is too similar, filter it out
			allowMessage = false
			log.Debugf("message is %f percent similar to a previous message, filtering",
				similarity*100)
			break
		}
	}
	if len(RecentACARSMessages) >= RecentMessageMax {
		// Remove the oldest message
		RecentACARSMessages = RecentACARSMessages[1:]
	}
	return allowMessage
}

func FilterDuplicateVDLM2(m VDLM2Message) bool {
	allowMessage := true
	if len(RecentVDLM2Messages) < RecentMessageMax {
		RecentVDLM2Messages = append(RecentVDLM2Messages, m)
	}
	for _, c := range RecentVDLM2Messages {
		similarity := strutil.Similarity(m.VDL2.AVLC.ACARS.MessageText, c.VDL2.AVLC.ACARS.MessageText, metrics.NewHamming())
		if similarity > config.FilterCriteriaVDLM2DuplicateMessageSimilarity {
			// Message is too similar, filter it out
			allowMessage = false
			log.Debugf("message is %f percent similar to a previous message, filtering",
				similarity*100)
			break
		}
	}
	if len(RecentVDLM2Messages) >= RecentMessageMax {
		// Remove the oldest message
		RecentVDLM2Messages = RecentVDLM2Messages[1:]
	}
	return allowMessage
}
