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
	// Only provide these fields to future steps.
	SelectedFields []string
}

func (a Tar1090Annotator) Name() string {
	return reflect.TypeOf(a).Name()
}

func (a Tar1090Annotator) GetDefaultFields() (s []string) {
	taj := FormatAsAPMessage(Tar1090AircraftJSON{}, "Tar1090")
	c := FormatAsAPMessage(TAR1090Calculated{}, "Tar1090")
	for f := range MergeAPMessages(taj, c) {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
}

func (a Tar1090Annotator) Configured() bool {
	return !reflect.DeepEqual(a, Tar1090Annotator{})
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
	AltimeterBarometerFeet     int64    `json:"alt_baro,omitempty" ap:"AircraftBarometerAltitudeFeet"`
	AltimeterBarometerRateFeet float64  `json:"baro_rate,omitempty" ap:"AircraftBarometerAltitudeRateFeet"`
	Squawk                     string   `json:"squawk,omitempty"`
	Emergency                  string   `json:"emergency,omitempty"`
	NavQNH                     float64  `json:"nav_qnh,omitempty"`
	NavAltitudeMCP             int64    `json:"nav_altitude_mcp,omitempty"`
	NavModes                   []string `json:"nav_modes,omitempty"`

	AltimeterGeometricFeet       float64 `json:"alt_geom,omitempty" ap:"AircraftAltitudeFeet"`
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
	AircraftGeolocation          string  `ap:"AircraftGeolocation"`
	AircraftGeolocationLatitude  float64 `ap:"AircraftLatitude"`
	AircraftGeolocationLongitude float64 `ap:"AircraftLongitude"`
	AircraftDistanceKm           float64 `ap:"AircraftDistanceKm"`
	AircraftDistanceMi           float64 `ap:"AircraftDistanceMi"`
}

func NormalizeAircraftRegistration(reg string) string {
	s := []string{
		".",
		" ",
		"-",
	}
	for _, r := range s {
		reg = strings.ReplaceAll(reg, r, "")
	}
	return strings.ToLower(reg)
}

// Wrapper around the SingleAircraftQueryByRegistration API
func (a Tar1090Annotator) SingleAircraftQueryByRegistration(reg string) (aircraft TJSONAircraft, err error) {
	reg = NormalizeAircraftRegistration(reg)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/data/aircraft.json?_=%d/", a.URL, time.Now().Unix()), nil)
	if err != nil {
		return aircraft, err
	}
	client := &http.Client{}

	log.Debug(Aside("%s: making call to tar1090", a.Name()))
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
				return aircraft, nil
			}
		}
		log.Debug(Aside("%s: aircraft not found", a.Name()))
		return aircraft, nil
	} else {
		return aircraft, errors.New("unable to parse returned aircraft position")
	}
}

func (a Tar1090Annotator) Annotate(m APMessage) (APMessage, error) {
	if a.ReferenceGeolocation == "" {
		log.Info(Note("%s enabled but geolocation not set, using '0,0'", a.Name()))
		a.ReferenceGeolocation = "0,0"
	}
	coords := strings.Split(a.ReferenceGeolocation, ",")
	if len(coords) != 2 {
		return m, fmt.Errorf("geolocation coordinates are not in the format 'LAT,LON'")
	}
	olat, _ := strconv.ParseFloat(coords[0], 64)
	olon, _ := strconv.ParseFloat(coords[1], 64)
	origin := geodist.Coord{Lat: olat, Lon: olon}

	aircraftInfo := TJSONAircraft{}
	var err error
	tailcode := GetAPMessageCommonFieldAsString(m, "TailCode")
	if tailcode == "" {
		log.Debug(Aside("%s: did not find a tail code in message, this is not unusual", a.Name()))
		return m, nil
	}
	aircraftInfo, err = a.SingleAircraftQueryByRegistration(tailcode)
	if err != nil {
		return m, fmt.Errorf("error finding aircraft position from tar1090: %v", err)
	}

	aircraft := geodist.Coord{Lat: aircraftInfo.Latitude, Lon: aircraftInfo.Longitude}
	var mi,km float64
	if aircraft.Lat == 0.0 && aircraft.Lon == 0.0 {
		// 0 distance but 0,0 geolocation should indicate geolocation failed.
		mi, km = 0,0
	} else {
		mi, km, err = geodist.VincentyDistance(origin, aircraft)
		if err != nil {
			return m, fmt.Errorf("error calculating distance: %s", err)
		}
	}

	c := TAR1090Calculated{
		AircraftGeolocation:          fmt.Sprintf("%f,%f", aircraftInfo.Latitude, aircraftInfo.Longitude),
		AircraftGeolocationLatitude:  aircraft.Lat,
		AircraftGeolocationLongitude: aircraft.Lon,
		AircraftDistanceKm:           km,
		AircraftDistanceMi:           mi,
	}
	// Create AP Messages from the API response and the calculated type above
	// then merge them.
	acinfo := FormatAsAPMessage(aircraftInfo, a.Name())
	calc := FormatAsAPMessage(c, a.Name())
	ac := MergeAPMessages(acinfo, calc)

	// Merge with the main AP Message and return
	return MergeAPMessages(m, ac), nil
}
