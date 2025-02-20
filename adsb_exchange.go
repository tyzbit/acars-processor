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

type ADSBHandler struct {
	SingleAircraftPosition SingleAircraftPosition
}

type SingleAircraftPosition struct {
	Aircraft []struct {
		HexCode                                       string        `json:"hex"`
		Type                                          string        `json:"type"`
		FlightNumber                                  string        `json:"flight"`
		AircraftTailCode                              string        `json:"r"`
		AircraftModel                                 string        `json:"t"`
		AltimeterBarometer                            int64         `json:"alt_baro"`
		AltimeterGeometricFeet                        int64         `json:"alt_geom"`
		GroundSpeedKnots                              int64         `json:"gs"`
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
func (a ADSBHandler) SingleAircraftPositionByRegistration(reg string) (ac SingleAircraftPosition, err error) {
	resp, err := http.Get(fmt.Sprintf(adsbapiv2, fmt.Sprintf("registration/%s/", reg)))
	if err != nil {
		return ac, err
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
func (a ADSBHandler) SingleAircraftDistanceFromPoint(reg string, g geopoint.GeoPoint) (geopoint.Kilometres, error) {
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
func (a ADSBHandler) SingleAircraftDistanceFromPointMiles(reg string, g geopoint.GeoPoint) (geopoint.Miles, error) {
	km, err := a.SingleAircraftDistanceFromPoint(reg, g)
	if err != nil {
		return geopoint.Miles(0), err
	}
	return km.Miles(), nil
}

// Interface function to satisfy ACARSHandler
func (a ADSBHandler) HandleACARSMessage(m ACARSMessage) string {
	if config.ADSBExchangeReferenceGeolocation == "" {
		log.Warn("ADSB enabled but geolocation not set")
		return ""
	}
	coords := strings.Split(config.ADSBExchangeReferenceGeolocation, ",")
	if len(coords) != 2 {
		log.Warn("geolocation coordinates are not in the format 'LAT,LON'")
		return ""
	}
	flat, _ := strconv.ParseFloat(coords[0], 64)
	flon, _ := strconv.ParseFloat(coords[1], 64)
	lat := geopoint.Degrees(flat)
	lon := geopoint.Degrees(flon)
	g := geopoint.NewGeoPoint(lat, lon)
	d, err := a.SingleAircraftDistanceFromPointMiles(m.AircraftTailCode, *g)
	if err != nil {
		log.Error("error determining distance between aircraft and point: %v", err)
		return ""
	}
	return fmt.Sprintf(`{"distance_miles": %v}`, float64(d))
}
