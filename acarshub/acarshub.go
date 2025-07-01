package acarshub

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	. "github.com/tyzbit/acars-processor/config"
	. "github.com/tyzbit/acars-processor/database"
	. "github.com/tyzbit/acars-processor/decorate"
	"gorm.io/gorm"
)

var (
	ACARSHubMaxConcurrentRequests = 1
	// These are just channels of IDs
	ACARSMessageQueue = make(chan uint, 10000)
	VDLM2MessageQueue = make(chan uint, 10000)
	ACARSQueue        []ACARSMessage
	VDLM2Queue        []VDLM2Message
)

func AutoMigrate() {
	if err := DB.AutoMigrate(ACARSMessage{}); err != nil {
		log.Fatal(Attention("Unable to automigrate annotator.ACARSMessage type: %s", err))
	}
	if err := DB.AutoMigrate(VDLM2Message{}); err != nil {
		log.Fatal(Attention("Unable to automigrate VDLM2Message type: %s", err))
	}
}

func ReadACARSHubACARSMessages() {
	if Config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests != 0 {
		ACARSHubMaxConcurrentRequests = Config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests
	}
	if Config.ACARSProcessorSettings.Database.Enabled {
		am := []ACARSMessage{}
		DB.Find(&am, ACARSMessage{Model: gorm.Model{DeletedAt: gorm.DeletedAt{Valid: false}}})
		for _, a := range am {
			ACARSMessageQueue <- a.ID
		}

		log.Info(Content("Loaded %d ACARS messages from the db", len(am)))
	}

	address := fmt.Sprintf("%s:%d", Config.ACARSProcessorSettings.ACARSHub.ACARS.Host, Config.ACARSProcessorSettings.ACARSHub.ACARS.Port)
	for {
		log.Debug(Aside("connecting to "), Note(Config.ACARSProcessorSettings.ACARSHub.ACARS.Host), Aside(" on acars json port "), Note(fmt.Sprint(Config.ACARSProcessorSettings.ACARSHub.ACARS.Port)))
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
			log.Info(Content("new acars message received"))
			if (next == ACARSMessage{}) {
				log.Error(Attention("json message did not match expected structure, we got: "),
					Emphasised("%+v", next))
				continue
			} else {
				queueLength := DB.Find(&[]ACARSMessage{{Model: gorm.Model{DeletedAt: gorm.DeletedAt{Valid: false}}}}).RowsAffected
				log.Debug(Content("new acars message content "),
					Note("(%d already in queue)", queueLength),
					Content(": "), Aside("%+v", next))
				DB.Create(&next)
				ACARSMessageQueue <- next.ID
				continue
			}
		}

		log.Warn(Attention("acars handler exited, reconnecting"))
		s.Close()
		time.Sleep(time.Second * 1)
	}
}

func ReadACARSHubVDLM2Messages() {
	if Config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests != 0 {
		ACARSHubMaxConcurrentRequests = Config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests
	}
	if Config.ACARSProcessorSettings.Database.Enabled {
		vm := []VDLM2Message{}
		DB.Find(&vm, VDLM2Message{Model: gorm.Model{DeletedAt: gorm.DeletedAt{Valid: false}}})
		for _, v := range vm {
			VDLM2MessageQueue <- v.ID
		}
		log.Info(Content("Loaded %d VDLM2 messages from the db", len(vm)))
	}

	address := fmt.Sprintf("%s:%d", Config.ACARSProcessorSettings.ACARSHub.VDLM2.Host, Config.ACARSProcessorSettings.ACARSHub.VDLM2.Port)
	for {
		log.Debug(Aside("connecting to "), Note(Config.ACARSProcessorSettings.ACARSHub.VDLM2.Host), Aside(" on vdlm2 json port "), Note(fmt.Sprint(Config.ACARSProcessorSettings.ACARSHub.VDLM2.Port)))
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
			log.Info(Content("new vdlm2 message received"))
			if (next == VDLM2Message{}) {
				log.Error(Attention("json message did not match expected structure, we got: %+v", next))
				continue
			} else {
				queueLength := DB.Find(&[]VDLM2Message{{Model: gorm.Model{DeletedAt: gorm.DeletedAt{Valid: false}}}}).RowsAffected
				log.Debug(Content("new vdlm2 message content "),
					Note("(%d already in queue)", queueLength),
					Content(": "), Aside("%+v", next))
				DB.Create(&next)
				VDLM2MessageQueue <- next.ID
				continue
			}
		}

		log.Warn(Attention("vdlm2 handler exited, reconnecting"))
		s.Close()
		time.Sleep(time.Second * 1)
	}
}

// Connects to ACARS and starts listening to messages
func SubscribeToACARSHub() {
	launched := false
	if Config.ACARSProcessorSettings.ACARSHub.ACARS.Host != "" && Config.ACARSProcessorSettings.ACARSHub.ACARS.Port != 0 {
		go ReadACARSHubACARSMessages()
		launched = true
	}
	if Config.ACARSProcessorSettings.ACARSHub.VDLM2.Host != "" && Config.ACARSProcessorSettings.ACARSHub.VDLM2.Port != 0 {
		go ReadACARSHubVDLM2Messages()
		launched = true
	}
	if !launched {
		log.Warn(Attention("no acarshub subscribers set, please check configuration"))
	} else {
		log.Debug(Aside("launched acarshub subscribers"))
	}
}
