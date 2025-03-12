package main

import (
	"fmt"

	"github.com/tidwall/words"
)

// Returns a dictionary to do word matching with.
func ACARSDictionary() []string {
	// Flight levels
	for i := 1; i < 50; i++ {
		ACARSLingo = append(ACARSLingo, fmt.Sprintf("FL%d0", i))
	}
	// Seats
	for i := 1; i < 30; i++ {
		for _, s := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"} {
			ACARSLingo = append(ACARSLingo, fmt.Sprintf("%d%s", i, s))
		}
	}
	ACARSLingo = append(ACARSLingo, words.Words...)
	return ACARSLingo
}

var ACARSLingo = []string{
	"A/T",
	"ACFT",
	"ADVZD",
	"ALT",
	"APCH",
	"APCHS",
	"ASR",
	"ASSIT",
	"ASSITS",
	"AUTOBRAKE",
	"AVAIL",
	"BLK",
	"CAPT",
	"CONNEX",
	"CONT",
	"CST",
	"DEADHEAD",
	"DEST",
	"DR",
	"DST",
	"EMS",
	"EST",
	"ETA",
	"ETOPS",
	"FAS",
	"FLOW",
	"FLT",
	"FMS",
	"FO",
	"FOB",
	"FWD",
	"FWD",
	"GPS",
	"GRND",
	"ICAO",
	"ILLUM",
	"INFLT",
	"INOP",
	"KNOTS",
	"KTS",
	"LAV",
	"LGHT",
	"LGT",
	"LT",
	"MST",
	"MTX",
	"MX",
	"OCCNL",
	"PAX",
	"PLS",
	"RE",
	"REQ",
	"RESYCD",
	"RLS",
	"RNAV",
	"RTN",
	"RTO",
	"RWY",
	"RWYS",
	"SIMUL",
	"T/O",
	"TCAS",
	"THX",
	"TURB",
	"TY",
	"UNSYCD",
	"UPL",
	"VIASAT",
	"VNAV",
	"WX",
}
