package main

import (
	"fmt"
	"regexp"

	// geo "github.com/flopp/go-coordsparser"
	geopoint "github.com/marcinwyszynski/geopoint"
)

type ACARSGeoHandlerAnnotator struct {
}

func (a ACARSGeoHandlerAnnotator) Name() string {
	return "ACARS Parsed Geolocation"
}

// Formats I've seen geolocation in (in the raw messages, not parsed by ACARSHUB)
//  3715 -7641 // HDM
// /POS N3659.4W07558.7 // HDM
// /POS N37056W075493 // HDM
// N 371406W 755226 // HDM
// N3639.1W07620.0 // HDM
// N36933W 76312369 // HDM
// POSN 36.841W 76.354 // D
// POSN36369W076071 // HDM

// Interface function to satisfy ACARSHandler
func (a ACARSGeoHandlerAnnotator) AnnotateACARSMessage(m ACARSMessage) (annotation Annotation) {
	messageText := m.MessageText
	// get ready for a gnarly fuckin regex
	re := regexp.MustCompile(`(\/?POS?\s?)?(?<lat>((([Nn]|[Ss])\s)?[0-9\.]+|([NSns0-9\.\-]+)))\s?(?<lon>.*)`)

	match := re.FindStringSubmatch(messageText)
	if match == nil {
		fmt.Println("No match found")
		return
	}
	return nil
}

func (a ACARSGeoHandlerAnnotator) ExtractGeolocationCoordinates(s string) geopoint.GeoPoint {
	return geopoint.GeoPoint{}
}
