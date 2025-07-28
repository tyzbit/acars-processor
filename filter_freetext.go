package main

import "strings"

var freetextTerms = []string{
	"BINGO",
	"CHOP",
	"COMMENTS",
	"CONFIRM",
	"DEFECT",
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
		// Some messages have DISP in them but are automated
		if strings.Contains(m, term) || strings.HasPrefix(m, "DISP") {
			present = true
			break
		}
	}
	return present
}
