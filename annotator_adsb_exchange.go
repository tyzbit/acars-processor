package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/marcinwyszynski/geopoint"
	log "github.com/sirupsen/logrus"
)

const adsbapiv2 = "https://adsbexchange-com1.p.rapidapi.com/v2/%s"
const adsbapikeyheader = "x-rapidapi-key"

func (a ADSBHandlerAnnotator) Name() string {
	return "ADS-B Exchange"
}

type ADSBHandlerAnnotator struct {
	SingleAircraftPosition SingleAircraftPosition
}

type SingleAircraftPosition struct {
	Aircraft []struct {
		HexCode                                       string        `json:"hex"`
		Type                                          string        `json:"type"`
		FlightNumber                                  string        `json:"flight"`
		AircraftTailCode                              string        `json:"r"`
		AircraftModel                                 string        `json:"t"`
		AltimeterBarometer                            interface{}   `json:"alt_baro"`
		AltimeterGeometricFeet                        int64         `json:"alt_geom"`
		GroundSpeedKnots                              float64       `json:"gs"`
		TrueGroundTrack                               float64       `json:"track"`
		AltimeterBarometerRateOfChangeFeet            int64         `json:"baro_rate"`
		Squawk                                        string        `json:"squawk"`
		Emergency                                     string        `json:"emergency"`
		EmitterCategory                               string        `json:"category"`
		NavAltimeterSettinghPa                        float64       `json:"nav_qnh"`
		NacAltitudeMCP                                float64       `json:"nav_altitude_mcp"`
		Latitude                                      float64       `json:"lat"`
		Longitude                                     float64       `json:"lon"`
		NavigationIntegrityCategory                   float64       `json:"nic"`
		RadiusOfContainment                           float64       `json:"rc"`
		SecondsSincePositionUpdated                   float64       `json:"seen_pos"`
		Version                                       float64       `json:"version"`
		NavigationIntegrityCategoryBarometricAltitude float64       `json:"nic_baro"`
		NavigationalPositionAccuracy                  float64       `json:"nac_p"`
		NavigationalVelocityAccuracy                  float64       `json:"nac_v"`
		SourceIntegrityLevel                          float64       `json:"sil"`
		SourceIntegrityLevelType                      string        `json:"sil_type"`
		GeometricVerticalAccuracy                     float64       `json:"gva"`
		SystemDesignAssurance                         float64       `json:"sda"`
		FlightStatusAlert                             float64       `json:"alert"`
		SpecialPositionIdentification                 float64       `json:"spi"`
		MLAT                                          []interface{} `json:"mlat"`
		TISB                                          []interface{} `json:"tisb"`
		AircraftTotalModeSMessages                    int64         `json:"messages"`
		SecondsSinceLastMessage                       float64       `json:"seen"`
		RSSISignalPowerdBm                            float64       `json:"rssi"`
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
	req.Header.Add(adsbapikeyheader, config.ADSBExchangeAPIKey)
	// Create a new HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &ac)
	if err != nil {
		return ac, err
	}

	if (&ac != &SingleAircraftPosition{}) {
		return ac, nil
	} else {
		return ac, errors.New("unable to parse returned aircraft position")
	}
}

// Helper function to calculate distance using registration and geopoint
func (a ADSBHandlerAnnotator) SingleAircraftDistanceFromPoint(reg string, g geopoint.GeoPoint) (geopoint.Kilometres, error) {
	ac, err := a.SingleAircraftPositionByRegistration(reg)
	if err != nil {
		return geopoint.Kilometres(0), err
	}

	lat := geopoint.Degrees(ac.Aircraft[0].Latitude)
	lon := geopoint.Degrees(ac.Aircraft[0].Longitude)
	km := g.DistanceTo(geopoint.NewGeoPoint(lat, lon), geopoint.Haversine)
	return km, nil
}

// Helper function to calculate distance using registration and geopoint, but in miles
func (a ADSBHandlerAnnotator) SingleAircraftDistanceFromPointMiles(reg string, g geopoint.GeoPoint) (geopoint.Miles, error) {
	km, err := a.SingleAircraftDistanceFromPoint(reg, g)
	if err != nil {
		return geopoint.Miles(0), err
	}
	return km.Miles(), nil
}

// Interface function to satisfy ACARSHandler
func (a ADSBHandlerAnnotator) AnnotateACARSMessage(m ACARSMessage) (annotation ACARSAnnotation) {
	if config.ADSBExchangeReferenceGeolocation == "" {
		log.Warn("ADSB enabled but geolocation not set")
		return
	}
	coords := strings.Split(config.ADSBExchangeReferenceGeolocation, ",")
	if len(coords) != 2 {
		log.Warn("geolocation coordinates are not in the format 'LAT,LON'")
		return annotation
	}
	flat, _ := strconv.ParseFloat(coords[0], 64)
	flon, _ := strconv.ParseFloat(coords[1], 64)
	lat := geopoint.Degrees(flat)
	lon := geopoint.Degrees(flon)
	g := geopoint.NewGeoPoint(lat, lon)

	km, err := a.SingleAircraftDistanceFromPoint(m.AircraftTailCode, *g)
	if err != nil {
		log.Errorf("error determining distance between aircraft and point: %v", err)
		return annotation
	}

	mi, err := a.SingleAircraftDistanceFromPointMiles(m.AircraftTailCode, *g)
	if err != nil {
		log.Errorf("error determining distance between aircraft and point: %v", err)
		return annotation
	}

	event := map[string]interface{}{
		"adsbFrequencyMHz":               m.FrequencyMHz,
		"adsbChannel":                    m.Channel,
		"adsbErrorCode":                  m.ErrorCode,
		"adsbSignaldBm":                  m.SignaldBm,
		"adsbAppName":                    m.App.Name,
		"adsbAppVersion":                 m.App.Version,
		"adsbRouterUUID":                 m.App.ACARSRouterUUID,
		"adsbStationID":                  m.StationID,
		"adsbAircraftTailCode":           m.AircraftTailCode,
		"adsbMessageText":                m.MessageText,
		"adsbMessageNumber":              m.MessageNumber,
		"adsbFlightNumber":               m.FlightNumber,
		"adsbOriginGeolocation":          config.ADSBExchangeReferenceGeolocation,
		"adsbOriginGeolocationLatitude":  flat,
		"adsbOriginGeolocationLongitude": flon,
		"adsbDistanceKm":                 int64(km),
		"adsbDistanceMi":                 int64(mi),
	}

	return ACARSAnnotation{
		Annotator:  "ADS-B Exchange",
		Annotation: event,
	}
}
