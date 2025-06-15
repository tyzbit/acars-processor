package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ACARSHubMaxConcurrentRequests = 1
	// These are just channels of IDs
	ACARSMessageQueue = make(chan uint, 10000)
	VDLM2MessageQueue = make(chan uint, 10000)
)

func ReadACARSHubACARSMessages() {
	if config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests != 0 {
		ACARSHubMaxConcurrentRequests = config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests
	}
	for range ACARSHubMaxConcurrentRequests {
		go HandleACARSJSONMessages(ACARSMessageQueue)
	}

	address := fmt.Sprintf("%s:%d", config.ACARSProcessorSettings.ACARSHub.ACARS.Host, config.ACARSProcessorSettings.ACARSHub.ACARS.Port)
	for {
		log.Debug(yo.INFODUMP("connecting to ").Hmm(config.ACARSProcessorSettings.ACARSHub.ACARS.Host).INFODUMP(" on acars json port ").Hmm(fmt.Sprint(config.ACARSProcessorSettings.ACARSHub.ACARS.Port)).FRFR())
		s, err := net.Dial("tcp", address)
		if err != nil {
			log.Error(yo.Uhh("error connecting to acars json: %v", err).FRFR())
			time.Sleep(time.Second * 1)
			continue
		}
		log.Info(yo.Bet("connected to acarshub acars json port successfully").FRFR())
		readJson := json.NewDecoder(io.Reader(s))
		log.Debug(yo.INFODUMP("handling acars json messages").FRFR())
		for {
			var next ACARSMessage
			if err := readJson.Decode(&next); err != nil {
				// Might have connection issues, exit to reconnect
				log.Error(yo.Uhh("error decoding acars message: %v", err).FRFR())
				break
			}
			log.Info(yo.FYI("new acars message received").FRFR())
			if (next == ACARSMessage{}) {
				log.Error(
					yo.Uhh("json message did not match expected structure, we got: ").
						BTW("%+v", next).FRFR(),
				)
				continue
			} else {
				queueLength := db.Find(&[]ACARSMessage{{Model: gorm.Model{DeletedAt: gorm.DeletedAt{Valid: false}}}}).RowsAffected
				log.Debug(yo.FYI("new acars message content ").
					Hmm("(%d already in queue)", queueLength).
					FYI(": ").INFODUMP("%+v", next).FRFR())
				db.Create(&next)
				ACARSMessageQueue <- next.ID
				continue
			}
		}

		log.Warn(yo.Uhh("acars handler exited, reconnecting").FRFR())
		s.Close()
		time.Sleep(time.Second * 1)
	}
}

func ReadACARSHubVDLM2Messages() {
	if config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests != 0 {
		ACARSHubMaxConcurrentRequests = config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests
	}
	for range ACARSHubMaxConcurrentRequests {
		go HandleVDLM2JSONMessages(VDLM2MessageQueue)
	}

	address := fmt.Sprintf("%s:%d", config.ACARSProcessorSettings.ACARSHub.VDLM2.Host, config.ACARSProcessorSettings.ACARSHub.VDLM2.Port)
	for {
		log.Debug(yo.INFODUMP("connecting to ").Hmm(config.ACARSProcessorSettings.ACARSHub.VDLM2.Host).INFODUMP(" on vdlm2 json port ").Hmm(fmt.Sprint(config.ACARSProcessorSettings.ACARSHub.VDLM2.Port)).FRFR())
		s, err := net.Dial("tcp", address)
		if err != nil {
			log.Error(yo.Uhh("error connecting to vdlm2 json: %v", err).FRFR())
			time.Sleep(time.Second * 1)
			break
		}
		log.Info(yo.Bet("connected to acarshub vdlm2 json port successfully").FRFR())
		readJson := json.NewDecoder(io.Reader(s))
		log.Debug(yo.INFODUMP("handling vdlm2 json messages").FRFR())
		for {
			var next VDLM2Message
			if err := readJson.Decode(&next); err != nil {
				// Might have connection issues, exit to reconnect
				log.Error(yo.Uhh("error decoding vdlm2 message: %v", err).FRFR())
				break
			}
			log.Info(yo.FYI("new vdlm2 message received").FRFR())
			if (next == VDLM2Message{}) {
				log.Error(yo.Uhh("json message did not match expected structure, we got: %+v", next).FRFR())
				continue
			} else {
				queueLength := db.Find(&[]VDLM2Message{{Model: gorm.Model{DeletedAt: gorm.DeletedAt{Valid: false}}}}).RowsAffected
				log.Debug(yo.FYI("new vdlm2 message content ").
					Hmm("(%d already in queue)", queueLength).
					FYI(": ").INFODUMP("%+v", next).FRFR())
				db.Create(&next)
				VDLM2MessageQueue <- next.ID
				continue
			}
		}

		log.Warn(yo.Uhh("vdlm2 handler exited, reconnecting").FRFR())
		s.Close()
		time.Sleep(time.Second * 1)
	}
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
	if !launched {
		log.Warn(yo.Uhh("no acarshub subscribers set, please check configuration (%s).FRFR()", configFilePath))
	} else {
		log.Debug(yo.INFODUMP("launched acarshub subscribers").FRFR())
	}
}

// Reads messages in the channel from ReadACARSHubVDLM2Messages, annotates and
// sends off to configured receivers
func HandleACARSJSONMessages(ACARSMessageQueue chan uint) {
	for id := range ACARSMessageQueue {
		// Create a message with the ID we're looking for
		message := ACARSMessage{
			Model: gorm.Model{
				ID: id,
			},
		}
		// Find that message
		db.Where(&message).Find(&message)
		if (message.CreatedAt == time.Time{}) {
			log.Error(yo.Uhh("couldn't find message with id %d", id))
		}
		message.ProcessingStartedAt = time.Now()
		annotations := map[string]any{}
		ok, filters := ACARSCriteriaFilter{}.Filter(message)
		if !ok {
			fd := strings.Join(filters, ",")
			log.Info(yo.FYI("message was filtered out by %s", fd).FRFR())
			log.Debug(
				yo.FYI("message ending in \"").
					Hmm(Last20Characters(message.MessageText)).
					FYI("\" took ").
					Hmm("%s seconds", time.Since(message.ProcessingStartedAt).Seconds()).
					FYI(" to process and was ingested ").
					Hmm("%s seconds ago", time.Since(message.CreatedAt).Seconds()).FRFR())
			db.Delete(&message)
			continue
		}
		// Annotate the message via all enabled annotators
		for _, h := range enabledACARSAnnotators {
			log.Debug(yo.FYI("annotating message with annotator").Hmm(" %s", h.Name()).FYI(": ").INFODUMP("%+v", message).FRFR())
			result := h.AnnotateACARSMessage(message)
			if result != nil {
				result = h.SelectFields(result)
				annotations = MergeMaps(result, annotations)
			}
		}
		if len(annotations) == 0 {
			log.Info(yo.Uhh("no annotations were produced, not calling any receivers").FRFR())
		} else {
			for _, r := range enabledReceivers {
				log.Debug(yo.FYI("sending acars event to reciever ").
					Hmm(r.Name()).
					FYI(": ").
					INFODUMP("%+v", annotations).FRFR(),
				)
				err := r.SubmitACARSAnnotations(annotations)
				if err != nil {
					log.Error(yo.Uhh("error submitting to %s, err: %v", r.Name(), err).FRFR())
				}
			}
		}
		log.Debug(
			yo.FYI("message ending in \"").
				Hmm(Last20Characters(message.MessageText)).
				FYI("\" took ").
				Hmm("%s seconds", time.Since(message.ProcessingStartedAt).Seconds()).
				FYI(" to process and was ingested ").
				Hmm("%s seconds ago", time.Since(message.CreatedAt).Seconds()).FRFR())
		db.Delete(&message)
	}
}

// Reads messages in the channel from ReadACARSHubACARSMessages, annotates and
// sends off to configured receivers
func HandleVDLM2JSONMessages(VDLM2MessageQueue chan uint) {
	for id := range VDLM2MessageQueue {
		// Create a message with the ID we're looking for
		message := VDLM2Message{
			Model: gorm.Model{
				ID: id,
			},
		}
		// Find that message
		db.Where(&message).Find(&message)
		if (message.CreatedAt == time.Time{}) {
			log.Error(yo.Uhh("couldn't find message with id %d", id))
		}
		message.ProcessingStartedAt = time.Now()
		annotations := map[string]any{}
		ok, filters := VDLM2CriteriaFilter{}.Filter(message)
		if !ok {
			log.Info(yo.FYI("message was filtered out by %s", strings.Join(filters, ",")).FRFR())
			log.Debug(
				yo.FYI("message ending in \"").
					Hmm(Last20Characters(message.VDL2.AVLC.ACARS.MessageText)).
					FYI("\" took ").
					Hmm("%s seconds", time.Since(message.ProcessingStartedAt).Seconds()).
					FYI(" to process and was ingested ").
					Hmm("%s seconds ago", time.Since(message.CreatedAt).Seconds()).FRFR())
			db.Delete(&message)
			continue
		}
		// Annotate the message via all enabled annotators
		for _, h := range enabledVDLM2Annotators {
			log.Debug(yo.FYI("annotating message with annotator").Hmm(" %s", h.Name()).FYI(": ").INFODUMP("%+v", message).FRFR())
			result := h.AnnotateVDLM2Message(message)
			if result != nil {
				result = h.SelectFields(result)
				annotations = MergeMaps(result, annotations)
			}
		}
		if len(annotations) == 0 {
			log.Info(yo.Uhh("no annotations were produced, not calling any receivers").FRFR())
		} else {
			for _, r := range enabledReceivers {
				log.Debug(yo.FYI("sending vdlm2 event to reciever ").
					Hmm("%s", r.Name()).
					FYI(": ").
					INFODUMP("%+v", annotations).FRFR())
				err := r.SubmitACARSAnnotations(annotations)
				if err != nil {
					log.Error(yo.Uhh("error submitting to %s, err: %v", r.Name(), err).FRFR())
				}
			}
		}
		log.Debug(
			yo.FYI("message ending in \"").
				Hmm(Last20Characters(message.VDL2.AVLC.ACARS.MessageText)).
				FYI("\" took ").
				Hmm("%s seconds", time.Since(message.ProcessingStartedAt).Seconds()).
				FYI(" to process and was ingested ").
				Hmm("%s seconds ago", time.Since(message.CreatedAt).Seconds()).FRFR())
		db.Delete(&message)
	}
}
