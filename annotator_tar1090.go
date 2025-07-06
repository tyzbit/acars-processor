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
	"time"

	"github.com/jftuga/geodist"
	log "github.com/sirupsen/logrus"
)

func (a Tar1090AnnotatorHandler) Name() string {
	return "tar1090"
}

func (a Tar1090AnnotatorHandler) SelectFields(annotation Annotation) Annotation {
	// If no fields are being selected, return annotation unchanged
	if config.Annotators.Tar1090.SelectedFields == nil {
		return annotation
	}
	selectedFields := Annotation{}
	for field, value := range annotation {
		if slices.Contains(config.Annotators.Tar1090.SelectedFields, field) {
			selectedFields[field] = value
		}
	}
	return selectedFields
}

func (t Tar1090AnnotatorHandler) DefaultFields() []string {
	// ACARS
	fields := []string{}
	for field := range t.AnnotateACARSMessage(ACARSMessage{}) {
		fields = append(fields, field)
	}
	for field := range t.AnnotateVDLM2Message(VDLM2Message{}) {
		fields = append(fields, field)
	}
	slices.Sort(fields)
	return fields
}

type Tar1090AnnotatorHandler struct {
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

// Wrapper around the SingleAircraftQueryByRegistration API
func (a Tar1090AnnotatorHandler) SingleAircraftQueryByRegistration(reg string) (aircraft TJSONAircraft, err error) {
	reg = NormalizeAircraftRegistration(reg)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/data/aircraft.json?_=%d/", config.Annotators.Tar1090.URL, time.Now().Unix()), nil)
	if err != nil {
		return aircraft, err
	}
	client := &http.Client{}

	log.Debug(Aside("making call to tar1090"))
	resp, err := client.Do(req)
	if err != nil {
		return aircraft, err
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
				log.Debug(Aside("returning data from tar1090"))
				return aircraft, nil
			}
		}
		log.Debug(Aside("aircraft not found in tar1090 response"))
		return aircraft, errors.New("aircraft not found in tar1090 response")
	} else {
		return aircraft, errors.New("unable to parse returned aircraft position")
	}
}

// Interface function to satisfy ACARSHandler
func (a Tar1090AnnotatorHandler) AnnotateACARSMessage(m ACARSMessage) (annotation Annotation) {
	enabled := config.Annotators.Tar1090.Enabled
	if config.Annotators.Tar1090.ReferenceGeolocation == "" {
		log.Info(Note("tar1090 enabled but geolocation not set, using '0,0'"))
		config.Annotators.Tar1090.ReferenceGeolocation = "0,0"
	}
	coords := strings.Split(config.Annotators.Tar1090.ReferenceGeolocation, ",")
	if enabled && len(coords) != 2 {
		log.Warn(Attention("tar1090 geolocation coordinates are not in the format 'LAT,LON'"))
		return annotation
	}
	olat, _ := strconv.ParseFloat(coords[0], 64)
	olon, _ := strconv.ParseFloat(coords[1], 64)
	origin := geodist.Coord{Lat: olat, Lon: olon}

	aircraftInfo := TJSONAircraft{}
	var err error
	if enabled {
		aircraftInfo, err = a.SingleAircraftQueryByRegistration(m.AircraftTailCode)
		if err != nil {
			log.Warn(Attention("error getting aircraft position from tar1090: %v", err))
			return annotation
		}
	}

	aircraft := geodist.Coord{Lat: aircraftInfo.Latitude, Lon: aircraftInfo.Longitude}
	mi, km, err := geodist.VincentyDistance(origin, aircraft)
	if err != nil {
		log.Warn(Attention("error calculating distance: %s", err))
	}

	var navmodes string
	for i, mode := range aircraftInfo.NavModes {
		if i != 0 {
			navmodes = mode + ","
		}
		navmodes = navmodes + mode
	}
	event := Annotation{
		"tar1090ReferenceGeolocation":                        config.Annotators.Tar1090.ReferenceGeolocation,
		"tar1090ReferenceGeolocationLatitude":                olat,
		"tar1090ReferenceGeolocationLongitude":               olon,
		"tar1090AircraftEmergency":                           aircraftInfo.Emergency,
		"tar1090AircraftGeolocation":                         aircraftInfo,
		"tar1090AircraftLatitude":                            aircraftInfo.Latitude,
		"tar1090AircraftLongitude":                           aircraftInfo.Longitude,
		"tar1090AircraftDistanceKm":                          km,
		"tar1090AircraftDistanceMi":                          mi,
		"tar1090AircraftDistanceNm":                          aircraftInfo.DistanceFromReceiverNm,
		"tar1090AircraftDirectionDegrees":                    aircraftInfo.DirectionFromReceiverDegrees,
		"tar1090AircraftAltimeterBarometerFeet":              aircraftInfo.AltimeterBarometerFeet,
		"tar1090AircraftAltimeterGeometricFeet":              aircraftInfo.AltimeterGeometricFeet,
		"tar1090AircraftAltimeterBarometerRateFeetPerSecond": aircraftInfo.AltimeterBarometerRateFeet,
		"tar1090AircraftOwnerOperator":                       aircraftInfo.AircraftOwnerOperator,
		"tar1090AircraftFlightNumber":                        aircraftInfo.AircraftTailCode,
		"tar1090AircraftHexCode":                             aircraftInfo.Hex,
		"tar1090AircraftType":                                aircraftInfo.AircraftType,
		"tar1090AircraftDescription":                         aircraftInfo.AircraftDescription,
		"tar1090AircraftYearOfManufacture":                   aircraftInfo.AircraftManufactureYear,
		"tar1090AircraftADSBMessageCount":                    aircraftInfo.MessageCount,
		"tar1090AircraftRSSIdBm":                             aircraftInfo.RSSISignalPowerdBm,
		"tar1090AircraftNavModes":                            navmodes,
	}

	return event
}

// Interface function to satisfy ACARSHandler
func (a Tar1090AnnotatorHandler) AnnotateVDLM2Message(m VDLM2Message) (annotation Annotation) {
	if config.Annotators.Tar1090.ReferenceGeolocation == "" {
		log.Info(Note("tar1090 enabled but geolocation not set, using '0,0'"))
		config.Annotators.Tar1090.ReferenceGeolocation = "0,0"
	}
	coords := strings.Split(config.Annotators.Tar1090.ReferenceGeolocation, ",")
	if len(coords) != 2 {
		log.Warn(Attention("geolocation coordinates are not in the format 'LAT,LON'"))
		return annotation
	}
	olat, _ := strconv.ParseFloat(coords[0], 64)
	olon, _ := strconv.ParseFloat(coords[1], 64)
	origin := geodist.Coord{Lat: olat, Lon: olon}

	aircraftInfo, err := a.SingleAircraftQueryByRegistration(NormalizeAircraftRegistration(m.VDL2.AVLC.ACARS.Registration))
	if err != nil {
		log.Warn(Attention("error getting aircraft position: %v", err))
		return annotation
	}

	alat, alon := aircraftInfo.Latitude, aircraftInfo.Longitude
	aircraft := geodist.Coord{Lat: alat, Lon: alon}
	mi, km, err := geodist.VincentyDistance(origin, aircraft)
	if err != nil {
		log.Warn(Attention("error calculating distance: %s", err))
	}

	var navmodes string
	for i, mode := range aircraftInfo.NavModes {
		if i != 0 {
			navmodes = mode + ","
		}
		navmodes = navmodes + mode
	}
	event := Annotation{
		"tar1090ReferenceGeolocation":                        config.Annotators.Tar1090.ReferenceGeolocation,
		"tar1090ReferenceGeolocationLatitude":                olat,
		"tar1090ReferenceGeolocationLongitude":               olon,
		"tar1090AircraftGeolocation":                         fmt.Sprintf("%f,%f", alat, alon),
		"tar1090AircraftLatitude":                            alat,
		"tar1090AircraftLongitude":                           alon,
		"tar1090AircraftDistanceKm":                          km,
		"tar1090AircraftDistanceMi":                          mi,
		"tar1090AircraftDistanceNm":                          aircraftInfo.DistanceFromReceiverNm,
		"tar1090AircraftDirectionDegrees":                    aircraftInfo.DirectionFromReceiverDegrees,
		"tar1090AircraftAltimeterBarometerFeet":              aircraftInfo.AltimeterBarometerFeet,
		"tar1090AircraftAltimeterGeometricFeet":              aircraftInfo.AltimeterGeometricFeet,
		"tar1090AircraftAltimeterBarometerRateFeetPerSecond": aircraftInfo.AltimeterBarometerRateFeet,
		"tar1090AircraftOwnerOperator":                       aircraftInfo.AircraftOwnerOperator,
		"tar1090AircraftFlightNumber":                        aircraftInfo.AircraftTailCode,
		"tar1090AircraftHexCode":                             aircraftInfo.Hex,
		"tar1090AircraftType":                                aircraftInfo.AircraftType,
		"tar1090AircraftDescription":                         aircraftInfo.AircraftDescription,
		"tar1090AircraftYearOfManufacture":                   aircraftInfo.AircraftManufactureYear,
		"tar1090AircraftADSBMessageCount":                    aircraftInfo.MessageCount,
		"tar1090AircraftRSSIdBm":                             aircraftInfo.RSSISignalPowerdBm,
		"tar1090AircraftNavModes":                            navmodes,
	}

	return event
}
