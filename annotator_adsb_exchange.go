package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/jftuga/geodist"
	log "github.com/sirupsen/logrus"
)

const adsbapiv2 = "https://adsbexchange-com1.p.rapidapi.com/v2/%s"
const adsbapikeyheader = "x-rapidapi-key"

func (a ADSBHandlerAnnotator) Name() string {
	return "ads-b exchange"
}

func (a ADSBHandlerAnnotator) SelectFields(annotation Annotation) Annotation {
	// If no fields are being selected, return annotation unchanged
	if config.Annotators.ADSBExchange.SelectedFields == nil {
		return annotation
	}
	selectedFields := Annotation{}
	for field, value := range annotation {
		if slices.Contains(config.Annotators.ADSBExchange.SelectedFields, field) {
			selectedFields[field] = value
		}
	}
	return selectedFields
}

type ADSBHandlerAnnotator struct {
	SingleAircraftPosition SingleAircraftPosition
}

// https://www.adsbexchange.com/api/aircraft/v2/docs along with some guesswork
type SingleAircraftPosition struct {
	Aircraft []struct {
		HexCode                                       string  `json:"hex"`
		Type                                          string  `json:"type"`
		FlightNumber                                  string  `json:"flight"`
		AircraftTailCode                              string  `json:"r"`
		AircraftModel                                 string  `json:"t"`
		AltimeterBarometerFeet                        any     `json:"alt_baro"`
		AltimeterGeometricFeet                        int64   `json:"alt_geom"`
		GroundSpeedKnots                              float64 `json:"gs"`
		TrueGroundTrack                               float64 `json:"track"`
		AltimeterBarometerRateOfChangeFeet            int64   `json:"baro_rate"`
		Squawk                                        string  `json:"squawk"`
		Emergency                                     string  `json:"emergency"`
		EmitterCategory                               string  `json:"category"`
		NavAltimeterSettinghPa                        float64 `json:"nav_qnh"`
		NacAltitudeMCP                                float64 `json:"nav_altitude_mcp"`
		Latitude                                      float64 `json:"lat"`
		Longitude                                     float64 `json:"lon"`
		NavigationIntegrityCategory                   float64 `json:"nic"`
		RadiusOfContainment                           float64 `json:"rc"`
		SecondsSincePositionUpdated                   float64 `json:"seen_pos"`
		Version                                       float64 `json:"version"`
		NavigationIntegrityCategoryBarometricAltitude float64 `json:"nic_baro"`
		NavigationalPositionAccuracy                  float64 `json:"nac_p"`
		NavigationalVelocityAccuracy                  float64 `json:"nac_v"`
		SourceIntegrityLevel                          float64 `json:"sil"`
		SourceIntegrityLevelType                      string  `json:"sil_type"`
		GeometricVerticalAccuracy                     float64 `json:"gva"`
		SystemDesignAssurance                         float64 `json:"sda"`
		FlightStatusAlert                             float64 `json:"alert"`
		SpecialPositionIdentification                 float64 `json:"spi"`
		MLAT                                          []any   `json:"mlat"`
		TISB                                          []any   `json:"tisb"`
		AircraftTotalModeSMessages                    int64   `json:"messages"`
		SecondsSinceLastMessage                       float64 `json:"seen"`
		RSSISignalPowerdBm                            float64 `json:"rssi"`
	} `json:"ac"`
	Message              string `json:"msg"`
	APITimestamp         int64  `json:"now"`
	TotalAircraftResults int64  `json:"total"`
	CacheTime            int64  `json:"ctime"`
	ServerProcessingTime int64  `json:"ptime"`
}

// Wrapper around the SingleAircraftPositionByRegistration API
func (a ADSBHandlerAnnotator) SingleAircraftPositionByRegistration(reg string) (ac SingleAircraftPosition, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(adsbapiv2, fmt.Sprintf("registration/%s/", reg)), nil)
	if err != nil {
		return ac, err
	}
	req.Header.Add(adsbapikeyheader, config.Annotators.ADSBExchange.APIKey)
	client := &http.Client{}

	log.Debug(yo().INFODUMP("making call to ads-b exchange").FRFR())
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &ac)
	if err != nil {
		return ac, err
	}

	if (&ac != &SingleAircraftPosition{}) {
		log.Debug(yo().INFODUMP("returning data from ads-b").FRFR())
		return ac, nil
	} else {
		return ac, errors.New("unable to parse returned aircraft position")
	}
}

// Interface function to satisfy ACARSHandler
func (a ADSBHandlerAnnotator) AnnotateACARSMessage(m ACARSMessage) (annotation Annotation) {
	if config.Annotators.ADSBExchange.ReferenceGeolocation == "" {
		log.Info(yo().Hmm("adsb enabled but geolocation not set, using '0,0'").FRFR())
		config.Annotators.ADSBExchange.ReferenceGeolocation = "0,0"
	}
	coords := strings.Split(config.Annotators.ADSBExchange.ReferenceGeolocation, ",")
	if len(coords) != 2 {
		log.Warn(yo().Uhh("geolocation coordinates are not in the format 'LAT,LON'").FRFR())
		return annotation
	}
	olat, _ := strconv.ParseFloat(coords[0], 64)
	olon, _ := strconv.ParseFloat(coords[1], 64)
	origin := geodist.Coord{Lat: olat, Lon: olon}

	position, err := a.SingleAircraftPositionByRegistration(m.AircraftTailCode)
	if err != nil {
		log.Warn(yo().Uhh("error getting aircraft position: %v", err).FRFR())
	}
	if len(position.Aircraft) == 0 {
		log.Warn(yo().Uhh("no aircraft were returned from ADS-B API, response message was: %s", position.Message).FRFR())
		return annotation
	}

	alat, alon := position.Aircraft[0].Latitude, position.Aircraft[0].Longitude
	aircraft := geodist.Coord{Lat: alat, Lon: alon}
	mi, km, err := geodist.VincentyDistance(origin, aircraft)
	if err != nil {
		log.Warn(yo().Uhh("error calculating distance: %s", err).FRFR())
	}
	event := Annotation{
		"adsbOriginGeolocation":          config.Annotators.ADSBExchange.ReferenceGeolocation,
		"adsbOriginGeolocationLatitude":  olat,
		"adsbOriginGeolocationLongitude": olon,
		"adsbAircraftGeolocation":        fmt.Sprintf("%f,%f", alat, alon),
		"adsbAircraftLatitude":           alat,
		"adsbAircraftLongitude":          alon,
		"adsbAircraftDistanceKm":         km,
		"adsbAircraftDistanceMi":         mi,
	}

	return event
}

// Interface function to satisfy ACARSHandler
func (a ADSBHandlerAnnotator) AnnotateVDLM2Message(m VDLM2Message) (annotation Annotation) {
	if config.Annotators.ADSBExchange.ReferenceGeolocation == "" {
		log.Info(yo().Hmm("adsb exchange enabled but geolocation not set, using '0,0'").FRFR())
		config.Annotators.ADSBExchange.ReferenceGeolocation = "0,0"
	}
	coords := strings.Split(config.Annotators.ADSBExchange.ReferenceGeolocation, ",")
	if len(coords) != 2 {
		log.Warn(yo().Uhh("adsb exchange geolocation coordinates are not in the format 'LAT,LON'").FRFR())
		return annotation
	}
	olat, _ := strconv.ParseFloat(coords[0], 64)
	olon, _ := strconv.ParseFloat(coords[1], 64)
	origin := geodist.Coord{Lat: olat, Lon: olon}

	position, err := a.SingleAircraftPositionByRegistration(NormalizeAircraftRegistration(m.VDL2.AVLC.ACARS.Registration))
	if err != nil {
		log.Warn(yo().Uhh("error getting aircraft position from adsb exchange: %v", err).FRFR())
	}
	if len(position.Aircraft) == 0 {
		log.Warn(yo().Uhh("no aircraft were returned from adsb exchange, response message was: %s", position.Message).FRFR())
		return annotation
	}

	alat, alon := position.Aircraft[0].Latitude, position.Aircraft[0].Longitude
	aircraft := geodist.Coord{Lat: alat, Lon: alon}
	mi, km, err := geodist.VincentyDistance(origin, aircraft)
	if err != nil {
		log.Warn(yo().Uhh("error calculating distance: %s", err).FRFR())
	}
	// Please update config example values if changed
	event := Annotation{
		"adsbOriginGeolocation":          config.Annotators.ADSBExchange.ReferenceGeolocation,
		"adsbOriginGeolocationLatitude":  olat,
		"adsbOriginGeolocationLongitude": olon,
		"adsbAircraftGeolocation":        fmt.Sprintf("%f,%f", alat, alon),
		"adsbAircraftLatitude":           alat,
		"adsbAircraftLongitude":          alon,
		"adsbAircraftDistanceKm":         km,
		"adsbAircraftDistanceMi":         mi,
	}

	return event
}
