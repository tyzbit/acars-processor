package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jftuga/geodist"
	log "github.com/sirupsen/logrus"
)

type Tar1090Annotator struct {
	Annotator
	Module
	// URL to your tar1090 instance
	URL string `jsonschema:"required,example:http://tar1090/" default:"http://tar1090/"`
	// Geolocation to use for distance calculations (LAT,LON).
	ReferenceGeolocation string `default:"35.6244416,139.7753782"`
	// Only return these fields when done.
	SelectedFields []string
}

func (a Tar1090Annotator) Name() string {
	return reflect.TypeOf(a).Name()
}

func (a Tar1090Annotator) GetDefaultFields() (s []string) {
	taj := FormatAsAPMessage(Tar1090AircraftJSON{})
	c := FormatAsAPMessage(TAR1090Calculated{})
	for f := range MergeMaps(taj, c) {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
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

	AltimeterGeometricFeet       float64 `json:"alt_geom,omitempty" acars:"aircraft_altitude_feet"`
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

type TAR1090Calculated struct {
	ReferenceGeolocation          string
	ReferenceGeolocationLatitude  float64
	ReferenceGeolocationLongitude float64
	AircraftGeolocation           string  `acars:"aircraft_geolocation"`
	AircraftGeolocationLatitude   float64 `acars:"aircraft_latitude"`
	AircraftGeolocationLongitude  float64 `acars:"aircraft_longitude"`
	AircraftDistanceKm            float64 `acars:"aircraft_distance_km"`
	AircraftDistanceMi            float64 `acars:"aircraft_distance_mi"`
}

type MLAT struct {
}

type TISB struct {
}

// Wrapper around the SingleAircraftQueryByRegistration API
func (a Tar1090Annotator) SingleAircraftQueryByRegistration(reg string) (aircraft TJSONAircraft, err error) {
	reg = NormalizeAircraftRegistration(reg)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/data/aircraft.json?_=%d/", a.URL, time.Now().Unix()), nil)
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
		return aircraft, nil
	} else {
		return aircraft, errors.New("unable to parse returned aircraft position")
	}
}

func (a Tar1090Annotator) Annotate(m APMessage) (APMessage, error) {
	if reflect.DeepEqual(a, Tar1090Annotator{}) {
		return m, nil
	}
	if a.ReferenceGeolocation == "" {
		log.Info(Note("tar1090 enabled but geolocation not set, using '0,0'"))
		a.ReferenceGeolocation = "0,0"
	}
	coords := strings.Split(a.ReferenceGeolocation, ",")
	if len(coords) != 2 {
		return m, fmt.Errorf("tar1090 geolocation coordinates are not in the format 'LAT,LON'")
	}
	olat, _ := strconv.ParseFloat(coords[0], 64)
	olon, _ := strconv.ParseFloat(coords[1], 64)
	origin := geodist.Coord{Lat: olat, Lon: olon}

	aircraftInfo := TJSONAircraft{}
	var err error
	tailcode := GetAPMessageCommonFieldAsString(m, "tail_code")
	if tailcode == "" {
		return m, fmt.Errorf("did not find a tail code in message")
	}
	aircraftInfo, err = a.SingleAircraftQueryByRegistration(tailcode)
	if err != nil {
		return m, fmt.Errorf("error finding aircraft position from tar1090: %v", err)
	}

	aircraft := geodist.Coord{Lat: aircraftInfo.Latitude, Lon: aircraftInfo.Longitude}
	mi, km, err := geodist.VincentyDistance(origin, aircraft)
	if err != nil {
		return m, fmt.Errorf("error calculating distance: %s", err)
	}

	var navmodes string
	for i, mode := range aircraftInfo.NavModes {
		if i != 0 {
			navmodes = mode + ","
		}
		navmodes = navmodes + mode
	}
	c := TAR1090Calculated{
		ReferenceGeolocation:          a.ReferenceGeolocation,
		ReferenceGeolocationLatitude:  olat,
		ReferenceGeolocationLongitude: olon,
		AircraftGeolocation:           fmt.Sprintf("%s,%s", aircraftInfo.Latitude, aircraftInfo.Longitude),
		AircraftGeolocationLatitude:   aircraft.Lat,
		AircraftGeolocationLongitude:  aircraft.Lon,
		AircraftDistanceKm:            km,
		AircraftDistanceMi:            mi,
	}
	m = MergeMaps(FormatAsAPMessage(c), m)
	return m, nil
}

func (a Tar1090Annotator) Configured() bool {
	return !reflect.DeepEqual(a, Tar1090Annotator{})
}
