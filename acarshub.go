package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"regexp"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	APMessageQueue = make(chan APMessageQeueueItem, 10000)
)

type APMessageQeueueItem struct {
	ACARSMessage
	VDLM2Message
	APMessage
}

func (ac ACARSConnectionConfig) GetDefaultFields() (s []string) {
	a := ACARSMessage{}.Prepare()
	for f := range a {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
}

func (vc VDLM2ConnectionConfig) GetDefaultFields() (s []string) {
	v := VDLM2Message{}.Prepare()
	for f := range v {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
}

// Connects to ACARS and starts listening to messages
func SubscribeToACARSHub() {
	launched := false
	if config.ACARSProcessorSettings.ACARSHub.ACARS.Host != "" && config.ACARSProcessorSettings.ACARSHub.ACARS.Port != 0 {
		go ReadACARSHubACARSMessages()
		launched = true
	}
	if config.ACARSProcessorSettings.ACARSHub.VDLM2.Host != "" && config.ACARSProcessorSettings.ACARSHub.VDLM2.Port != 0 {
		go ReadACARSHubVDLM2Messages()
		launched = true
	}
	go HandleAPMessageQueue(APMessageQueue)
	if !launched {
		log.Warn(Attention("no acarshub subscribers set, please check configuration (%s)()", configFilePath))
	} else {
		log.Debug(Aside("launched acarshub subscribers"))
	}
}

func ReadACARSHubACARSMessages() {
	address := fmt.Sprintf("%s:%d", config.ACARSProcessorSettings.ACARSHub.ACARS.Host, config.ACARSProcessorSettings.ACARSHub.ACARS.Port)
	for {
		log.Debug(Aside("connecting to "), Note(config.ACARSProcessorSettings.ACARSHub.ACARS.Host), Aside(" on acars json port "), Note(fmt.Sprint(config.ACARSProcessorSettings.ACARSHub.ACARS.Port)))
		s, err := net.Dial("tcp", address)
		if err != nil {
			log.Error(Attention("error connecting to acars json: %v", err))
			time.Sleep(time.Second * 1)
			continue
		}
		log.Info(Success("connected to acarshub acars json port successfully"))
		readJson := json.NewDecoder(io.Reader(s))
		log.Debug(Aside("handling acars json messages"))
		for {
			var next ACARSMessage
			if err := readJson.Decode(&next); err != nil {
				// Might have connection issues, exit to reconnect
				log.Error(Attention("error decoding acars message: %v", err))
				break
			}
			log.Info(Content("new acars message received ending in \""),
				Note(Last20Characters(next.MessageText)),
				Content("\""))
			if (next == ACARSMessage{}) {
				log.Error(Attention("json message did not match expected structure, we got: "),
					Emphasised("%+v", next))
				continue
			} else {
				queueLength := len(APMessageQueue)
				nextap := next.Prepare()
				next.MessageText = strings.ReplaceAll(next.MessageText, "\n", "\t")
				log.Debug(Emphasised("new acars message content "),
					Note("(%d already in queue)", queueLength),
					Content(": "), Aside("%+v", next))
				db.Create(&next)
				APMessageQueue <- APMessageQeueueItem{
					ACARSMessage: next,
					APMessage:    nextap,
				}
				continue
			}
		}

		log.Warn(Attention("acars handler exited, reconnecting"))
		s.Close()
		time.Sleep(time.Second * 1)
	}
}

func ReadACARSHubVDLM2Messages() {
	address := fmt.Sprintf("%s:%d", config.ACARSProcessorSettings.ACARSHub.VDLM2.Host, config.ACARSProcessorSettings.ACARSHub.VDLM2.Port)
	for {
		log.Debug(Aside("connecting to "), Note(config.ACARSProcessorSettings.ACARSHub.VDLM2.Host), Aside(" on vdlm2 json port "), Note(fmt.Sprint(config.ACARSProcessorSettings.ACARSHub.VDLM2.Port)))
		s, err := net.Dial("tcp", address)
		if err != nil {
			log.Error(Attention("error connecting to vdlm2 json: %v", err))
			time.Sleep(time.Second * 1)
			break
		}
		log.Info(Success("connected to acarshub vdlm2 json port successfully"))
		readJson := json.NewDecoder(io.Reader(s))
		log.Debug(Aside("handling vdlm2 json messages"))
		for {
			var next VDLM2Message
			if err := readJson.Decode(&next); err != nil {
				// Might have connection issues, exit to reconnect
				log.Error(Attention("error decoding vdlm2 message: %v", err))
				break
			}
			log.Info(Content("new vdlm2 message received ending in \""),
				Note(Last20Characters(next.VDL2.AVLC.ACARS.MessageText)),
				Content("\""))
			if (next == VDLM2Message{}) {
				log.Error(Attention("json message did not match expected structure, we got: %+v", next))
				continue
			} else {
				queueLength := len(APMessageQueue)
				nextap := next.Prepare()
				next.VDL2.AVLC.ACARS.MessageText = strings.ReplaceAll(next.VDL2.AVLC.ACARS.MessageText, "\n", "\t")
				log.Debug(Emphasised("new vdlm2 message content "),
					Note("(%d already in queue)", queueLength),
					Content(": "), Aside("%+v", next))
				db.Create(&next)
				APMessageQueue <- APMessageQeueueItem{
					VDLM2Message: next,
					APMessage:    nextap,
				}
				continue
			}
		}

		log.Warn(Attention("vdlm2 handler exited, reconnecting"))
		s.Close()
		time.Sleep(time.Second * 1)
	}
}

// Returns if the string is empty or if it only contains nonprintable characters
func AircraftOrTower(s string) (r string) {
	b, _ := regexp.Match("\\S+", []byte(s))
	if b {
		return "Aircraft"
	}
	return "Tower"
}
