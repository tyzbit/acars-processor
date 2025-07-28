package main

import "strings"

var freetextTerms = []string{
	"BINGO",
	"CHOP",
	"COMMENTS",
	"CONFIRM",
	"DEFECT",
	"DISP",
	"EVENING",
	"FREETEXT",
	"FTM",
	"INOP",
	"MEET",
	"MSG FROM",
	"PAN-PAN",
	"PAX",
	"POTABLE",
	"TEXT",
	"THANKS",
	"THX",
	"TXT",
}

func FreetextTermPresent(m string) (present bool) {
	for _, term := range freetextTerms {
		if strings.Contains(m, term) {
			present = true
			break
		}
	}
	return present
}
