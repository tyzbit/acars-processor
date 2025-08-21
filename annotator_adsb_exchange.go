package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/jftuga/geodist"
	log "github.com/sirupsen/logrus"
)

const adsbapiv2 = "https://adsbexchange-com1.p.rapidapi.com/v2/%s"
const adsbapikeyheader = "x-rapidapi-key"

type ADSBExchangeAnnotator struct {
	Annotator
	Module
	// APIKey provided by signing up at ADSB-Exchange.
	APIKey string `jsonschema:"required" default:"example_key"`
	// Geolocation to use for distance calculations (LAT,LON).
	ReferenceGeolocation string `default:"35.6244416,139.7753782"`
	// Only provide these fields to future steps.
	SelectedFields []string
}

func (a ADSBExchangeAnnotator) Name() string {
	return reflect.TypeOf(a).Name()
}

func (a ADSBExchangeAnnotator) GetDefaultFields() (s []string) {
	sap := FormatAsAPMessage(SingleAircraftPosition{})
	c := FormatAsAPMessage(ADSBExchangeCalculated{})
	for f := range MergeAPMessages(sap, c) {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
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
		NavigationIntegrityCategoryBarometricAltitude float64 `json:"nic_baro" ap:"AircraftAltitudeFeet"`
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

type ADSBExchangeCalculated struct {
	AircraftGeolocation          string  `ap:"AircraftGeolocation"`
	AircraftGeolocationLatitude  float64 `ap:"AircraftLatitude"`
	AircraftGeolocationLongitude float64 `ap:"AircraftLongitude"`
	AircraftDistanceKm           float64 `ap:"AircraftDistanceKm"`
	AircraftDistanceMi           float64 `ap:"AircraftDistanceMi"`
}

// Wrapper around the SingleAircraftPositionByRegistration API
func (a ADSBExchangeAnnotator) SingleAircraftPositionByRegistration(reg string) (ac SingleAircraftPosition, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(adsbapiv2, fmt.Sprintf("registration/%s/", reg)), nil)
	if err != nil {
		return ac, err
	}
	req.Header.Add(adsbapikeyheader, a.APIKey)
	client := &http.Client{}

	log.Debug(Aside("making call to ads-b exchange"))
	resp, err := client.Do(req)
	if err != nil {
		log.Error(Attention(fmt.Sprintf("%s", err)))
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

func (a ADSBExchangeAnnotator) Annotate(m APMessage) (APMessage, error) {
	if reflect.DeepEqual(a, ADSBExchangeAnnotator{}) {
		return m, nil
	}
	if a.ReferenceGeolocation == "" {
		log.Info(Note("%s enabled but geolocation not set, using '0,0'", a.Name()))
		a.ReferenceGeolocation = "0,0"
	}
	coords := strings.Split(a.ReferenceGeolocation, ",")
	if len(coords) != 2 {
		return m, fmt.Errorf("%s: geolocation coordinates are not in the format 'LAT,LON'", a.Name())
	}
	olat, _ := strconv.ParseFloat(coords[0], 64)
	olon, _ := strconv.ParseFloat(coords[1], 64)
	origin := geodist.Coord{Lat: olat, Lon: olon}
	tailcode := GetAPMessageCommonFieldAsString(m, "TailCode")
	if tailcode == "" {
		log.Debug(Note("%s: did not find a tail code in message, this is normal.", a.Name()))
		return m, nil
	}
	tailcode = normalizeComment(tailcode)
	pos, err := a.SingleAircraftPositionByRegistration(tailcode)
	if err != nil {
		return m, fmt.Errorf("error finding aircraft position: %v", err)
	}
	if len(pos.Aircraft) == 0 {
		log.Info(Note("%s: no aircraft were returned from ADS-B API, response message was: %s", a.Name(), pos.Message))
		return m, nil
	}
	apm := FormatAsAPMessage(pos)
	var alat, alon float64
	alat, alon = pos.Aircraft[0].Latitude, pos.Aircraft[0].Longitude
	aircraft := geodist.Coord{Lat: alat, Lon: alon}
	mi, km, err := geodist.VincentyDistance(origin, aircraft)
	if err != nil {
		log.Warn(Attention("%s: error calculating distance: %s", a.Name(), err))
	}
	calc := ADSBExchangeCalculated{
		AircraftGeolocation:          fmt.Sprintf("%f,%f", alat, alon),
		AircraftGeolocationLatitude:  alat,
		AircraftGeolocationLongitude: alon,
		AircraftDistanceKm:           km,
		AircraftDistanceMi:           mi,
	}
	calcapm := FormatAsAPMessage(calc)
	apm = MergeAPMessages(apm, calcapm)
	// Remove all but any selected fields
	if len(a.SelectedFields) > 0 {
		for field := range apm {
			if !slices.Contains(a.SelectedFields, field) {
				delete(apm, field)
			}
		}
	}
	return apm, nil
}

func (a ADSBExchangeAnnotator) Configured() bool {
	return !reflect.DeepEqual(a, ADSBExchangeAnnotator{})
}
