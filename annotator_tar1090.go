package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/marcinwyszynski/geopoint"
	log "github.com/sirupsen/logrus"
)

func (a Tar1090Handler) Name() string {
	return "tar1090"
}

func (a Tar1090Handler) SelectFields(annotation Annotation) Annotation {
	// If no fields are being selected, return annotation unchanged
	if config.TAR1090AnnotatorSelectedFields == "" {
		return annotation
	}
	selectedFields := Annotation{}
	if config.TAR1090AnnotatorSelectedFields != "" {
		for field, value := range annotation {
			if strings.Contains(config.TAR1090AnnotatorSelectedFields, field) {
				selectedFields[field] = value
			}
		}
	}
	return selectedFields
}

type Tar1090Handler struct {
	Tar1090AircraftJSON
}

type Tar1090AircraftJSON struct {
	Now      float64         `json:"now,omitempty"`
	Messages int64           `json:"messages,omitempty"`
	Aircraft []TJSONAircraft `json:"aircraft,omitempty"`
}

// The FIXME values are because I don't know what they are
type TJSONAircraft struct {
	Hex                        string   `json:"hex,omitempty"`
	Type                       string   `json:"type,omitempty"`
	AircraftTailCode           string   `json:"flight,omitempty"`
	Registration               string   `json:"r,omitempty"`
	AircraftType               string   `json:"t,omitempty"`
	AircraftDescription        string   `json:"desc,omitempty"`
	AircraftOwnerOperator      string   `json:"ownOp,omitempty"`
	AircraftManufactureYear    string   `json:"year,omitempty"`
	AltimeterBarometerFeet     int64    `json:"alt_baro,omitempty"`
	AltimeterBarometerRateFeet float64  `json:"baro_rate,omitempty"`
	Squawk                     string   `json:"squawk,omitempty"`
	Emergency                  string   `json:"emergency,omitempty"`
	NavQNH                     float64  `json:"nav_qnh,omitempty"`
	NavAltitudeMCP             int64    `json:"nav_altitude_mcp,omitempty"`
	NavModes                   []string `json:"nav_modes,omitempty"`

	AltimeterGeometricFeet       float64 `json:"alt_geom,omitempty"`
	GsFIXME                      float64 `json:"gs,omitempty"`
	Track                        float64 `json:"track,omitempty"`
	Category                     string  `json:"category,omitempty"`
	Latitude                     float64 `json:"lat,omitempty"`
	Longitude                    float64 `json:"lon,omitempty"`
	NICFIXME                     int64   `json:"nic,omitempty"`
	RCFIXME                      int64   `json:"rc,omitempty"`
	SeenPosition                 float64 `json:"seen_pos,omitempty"`
	DistanceFromReceiverNm       float64 `json:"r_dst,omitempty"`
	DirectionFromReceiverDegrees float64 `json:"r_dir,omitempty"`
	Version                      int64   `json:"version,omitempty"`
	NICBarometric                int64   `json:"nic_baro,omitempty"`
	NACP                         int64   `json:"nac_p,omitempty"`
	NACV                         int64   `json:"nac_v,omitempty"`
	SIL                          int64   `json:"sil,omitempty"`
	SILType                      string  `json:"sil_type,omitempty"`
	Alert                        int64   `json:"alert,omitempty"`
	SPI                          int64   `json:"spi,omitempty"`
	GVA                          int64   `json:"gva,omitempty"`
	SDA                          int64   `json:"sda,omitempty"`
	// TODO
	// MLAT                         []struct {
	// } `json:"mlat,omitempty"`
	// TISB []struct {
	// } `json:"tisb,omitempty"`
	MessageCount       int64   `json:"messages,omitempty"`
	Seen               float64 `json:"seen,omitempty"`
	RSSISignalPowerdBm float64 `json:"rssi,omitempty"`
}

type MLAT struct {
}

type TISB struct {
}

// Wrapper around the SingleAircraftPositionByRegistration API
func (a Tar1090Handler) SingleAircraftPositionByRegistration(reg string) (aircraft TJSONAircraft, err error) {
	reg = NormalizeAircraftRegistration(reg)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/data/aircraft.json?_=%d/", config.TAR1090URL, time.Now().Unix()), nil)
	if err != nil {
		return aircraft, err
	}
	client := &http.Client{}

	log.Debug("making call to tar1090")
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	tjson := Tar1090AircraftJSON{}
	err = json.Unmarshal(body, &tjson)
	if err != nil {
		return aircraft, err
	}

	if (&tjson != &Tar1090AircraftJSON{}) {
		for _, aircraft := range tjson.Aircraft {
			// Strip spaces and periods
			if NormalizeAircraftRegistration(aircraft.Registration) == reg {
				log.Debug("returning data from tar1090")
				return aircraft, nil
			}
		}
		log.Debug("aircraft not found in tar1090 response")
		return aircraft, errors.New("aircraft not found in tar1090 response")
	} else {
		return aircraft, errors.New("unable to parse returned aircraft position")
	}
}

// Interface function to satisfy ACARSHandler
func (a Tar1090Handler) AnnotateACARSMessage(m ACARSMessage) (annotation Annotation) {
	if config.TAR1090ReferenceGeolocation == "" {
		log.Info("tar1090 enabled but geolocation not set, using '0,0'")
		config.TAR1090ReferenceGeolocation = "0,0"
	}
	coords := strings.Split(config.TAR1090ReferenceGeolocation, ",")
	if len(coords) != 2 {
		log.Warn("tar1090 geolocation coordinates are not in the format 'LAT,LON'")
		return annotation
	}
	flat, _ := strconv.ParseFloat(coords[0], 64)
	flon, _ := strconv.ParseFloat(coords[1], 64)
	lat := geopoint.Degrees(flat)
	lon := geopoint.Degrees(flon)
	o := geopoint.NewGeoPoint(lat, lon)

	aircraft, err := a.SingleAircraftPositionByRegistration(m.AircraftTailCode)
	if err != nil {
		log.Warnf("error getting aircraft position from tar1090: %v", err)
		return annotation
	}

	airlat := geopoint.Degrees(aircraft.Latitude)
	airlon := geopoint.Degrees(aircraft.Longitude)
	airgeo := fmt.Sprintf("%f,%f", aircraft.Latitude, aircraft.Longitude)
	air := geopoint.NewGeoPoint(airlat, airlon)

	var navmodes string
	for i, mode := range aircraft.NavModes {
		if i != 0 {
			navmodes = mode + ","
		}
		navmodes = navmodes + mode
	}
	event := Annotation{
		"tar1090OriginGeolocation":                           config.TAR1090ReferenceGeolocation,
		"tar1090OriginGeolocationLatitude":                   flat,
		"tar1090OriginGeolocationLongitude":                  flon,
		"tar1090AircraftEmergency":                           aircraft.Emergency,
		"tar1090AircraftGeolocation":                         airgeo,
		"tar1090AircraftLatitude":                            aircraft.Latitude,
		"tar1090AircraftLongitude":                           aircraft.Longitude,
		"tar1090AircraftHaversineDistanceKm":                 float64(air.DistanceTo(o, geopoint.Haversine)),
		"tar1090AircraftHaversineDistanceMi":                 float64(air.DistanceTo(o, geopoint.Haversine).Miles()),
		"tar1090AircraftDistanceNm":                          aircraft.DistanceFromReceiverNm,
		"tar1090AircraftDirectionDegrees":                    aircraft.DirectionFromReceiverDegrees,
		"tar1090AircraftAltimeterBarometerFeet":              aircraft.AltimeterBarometerFeet,
		"tar1090AircraftAltimeterGeometricFeet":              aircraft.AltimeterGeometricFeet,
		"tar1090AircraftAltimeterBarometerRateFeetPerSecond": aircraft.AltimeterBarometerRateFeet,
		"tar1090AircraftOwnerOperator":                       aircraft.AircraftOwnerOperator,
		"tar1090AircraftFlightNumber":                        aircraft.AircraftTailCode,
		"tar1090AircraftHexCode":                             aircraft.Hex,
		"tar1090AircraftType":                                aircraft.AircraftType,
		"tar1090AircraftDescription":                         aircraft.AircraftDescription,
		"tar1090AircraftYearOfManufacture":                   aircraft.AircraftManufactureYear,
		"tar1090AircraftADSBMessageCount":                    aircraft.MessageCount,
		"tar1090AircraftRSSIdBm":                             aircraft.RSSISignalPowerdBm,
		"tar1090AircraftNavModes":                            navmodes,
	}

	return event
}

// Interface function to satisfy ACARSHandler
func (a Tar1090Handler) AnnotateVDLM2Message(m VDLM2Message) (annotation Annotation) {
	if config.ADSBExchangeReferenceGeolocation == "" {
		log.Info("tar1090 enabled but geolocation not set, using '0,0'")
		config.ADSBExchangeReferenceGeolocation = "0,0"
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
	o := geopoint.NewGeoPoint(lat, lon)

	aircraft, err := a.SingleAircraftPositionByRegistration(NormalizeAircraftRegistration(m.VDL2.AVLC.ACARS.Registration))
	if err != nil {
		log.Warnf("error getting aircraft position: %v", err)
		return annotation
	}

	airlat := geopoint.Degrees(aircraft.Latitude)
	airlon := geopoint.Degrees(aircraft.Longitude)
	airgeo := fmt.Sprintf("%f,%f", aircraft.Latitude, aircraft.Longitude)
	air := geopoint.NewGeoPoint(airlat, airlon)

	var navmodes string
	for i, mode := range aircraft.NavModes {
		if i != 0 {
			navmodes = mode + ","
		}
		navmodes = navmodes + mode
	}
	event := Annotation{
		"tar1090OriginGeolocation":                           config.TAR1090ReferenceGeolocation,
		"tar1090OriginGeolocationLatitude":                   flat,
		"tar1090OriginGeolocationLongitude":                  flon,
		"tar1090AircraftGeolocation":                         airgeo,
		"tar1090AircraftLatitude":                            aircraft.Latitude,
		"tar1090AircraftLongitude":                           aircraft.Longitude,
		"tar1090AircraftHaversineDistanceKm":                 float64(air.DistanceTo(o, geopoint.Haversine)),
		"tar1090AircraftHaversineDistanceMi":                 float64(air.DistanceTo(o, geopoint.Haversine).Miles()),
		"tar1090AircraftDistanceNm":                          aircraft.DistanceFromReceiverNm,
		"tar1090AircraftDirectionDegrees":                    aircraft.DirectionFromReceiverDegrees,
		"tar1090AircraftAltimeterBarometerFeet":              aircraft.AltimeterBarometerFeet,
		"tar1090AircraftAltimeterGeometricFeet":              aircraft.AltimeterGeometricFeet,
		"tar1090AircraftAltimeterBarometerRateFeetPerSecond": aircraft.AltimeterBarometerRateFeet,
		"tar1090AircraftOwnerOperator":                       aircraft.AircraftOwnerOperator,
		"tar1090AircraftFlightNumber":                        aircraft.AircraftTailCode,
		"tar1090AircraftHexCode":                             aircraft.Hex,
		"tar1090AircraftType":                                aircraft.AircraftType,
		"tar1090AircraftDescription":                         aircraft.AircraftDescription,
		"tar1090AircraftYearOfManufacture":                   aircraft.AircraftManufactureYear,
		"tar1090AircraftADSBMessageCount":                    aircraft.MessageCount,
		"tar1090AircraftRSSIdBm":                             aircraft.RSSISignalPowerdBm,
		"tar1090AircraftNavModes":                            navmodes,
	}

	return event
}
