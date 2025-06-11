package main

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/words"
)

// Unoptomized asf
func LongestDictionaryWordPhraseLength(messageText string) (wc int64) {
	var consecutiveWordSlice, maxConsecutiveWordSlice []string
	wordSlice := strings.Split(messageText, " ")
	for idx, word := range wordSlice {
		var found bool
		for _, dictWord := range words.Words {
			found = false
			if strings.EqualFold(word, dictWord) {
				consecutiveWordSlice = append(consecutiveWordSlice, word)
				found = true
				// We don't need to search for further matches
				break
			}
		}
		if !found || idx == len(wordSlice)-1 {
			if len(consecutiveWordSlice) >= len(maxConsecutiveWordSlice) {
				maxConsecutiveWordSlice = consecutiveWordSlice
				consecutiveWordSlice = []string{}
			}
		}
	}

	wc = int64(len(maxConsecutiveWordSlice))
	log.Debug(yo.INFODUMP("message had %d consecutive dictionary words in it", wc))
	if wc > 0 {
		log.Debug(yo.INFODUMP("longest dictionary word phrase found: %s", strings.Join(maxConsecutiveWordSlice, " ")))
	}
	return wc
}
