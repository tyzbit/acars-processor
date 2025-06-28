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
			log.Info(Content("new acars message received"))
			if (next == ACARSMessage{}) {
				log.Error(Attention("json message did not match expected structure, we got: "),
					Emphasised("%+v", next))
				continue
			} else {
				queueLength := db.Find(&[]ACARSMessage{{Model: gorm.Model{DeletedAt: gorm.DeletedAt{Valid: false}}}}).RowsAffected
				log.Debug(Content("new acars message content "),
					Note("(%d already in queue)", queueLength),
					Content(": "), Aside("%+v", next))
				db.Create(&next)
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
	if config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests != 0 {
		ACARSHubMaxConcurrentRequests = config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests
	}
	for range ACARSHubMaxConcurrentRequests {
		go HandleVDLM2JSONMessages(VDLM2MessageQueue)
	}

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
			log.Info(Content("new vdlm2 message received"))
			if (next == VDLM2Message{}) {
				log.Error(Attention("json message did not match expected structure, we got: %+v", next))
				continue
			} else {
				queueLength := db.Find(&[]VDLM2Message{{Model: gorm.Model{DeletedAt: gorm.DeletedAt{Valid: false}}}}).RowsAffected
				log.Debug(Content("new vdlm2 message content "),
					Note("(%d already in queue)", queueLength),
					Content(": "), Aside("%+v", next))
				db.Create(&next)
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
	if config.ACARSProcessorSettings.ACARSHub.ACARS.Host != "" && config.ACARSProcessorSettings.ACARSHub.ACARS.Port != 0 {
		go ReadACARSHubACARSMessages()
		launched = true
	}
	if config.ACARSProcessorSettings.ACARSHub.VDLM2.Host != "" && config.ACARSProcessorSettings.ACARSHub.VDLM2.Port != 0 {
		go ReadACARSHubVDLM2Messages()
		launched = true
	}
	if !launched {
		log.Warn(Attention("no acarshub subscribers set, please check configuration (%s)()", configFilePath))
	} else {
		log.Debug(Aside("launched acarshub subscribers"))
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
			log.Error(Attention("couldn't find message with id %d", id))
		}
		message.ProcessingStartedAt = time.Now()
		annotations := map[string]any{}
		ok, filters := ACARSCriteriaFilter{}.Filter(message)
		if !ok {
			fd := strings.Join(filters, ",")
			log.Info(Content("message was filtered out by %s", fd))
			log.Debug(Content("message ending in \""),
				Note(Last20Characters(message.MessageText)),
				Content("\" took "),
				Note("%.2f seconds", time.Since(message.ProcessingStartedAt).Seconds()),
				Content(" to process and was ingested "),
				Note("%.2f seconds ago", time.Since(message.CreatedAt).Seconds()))
			db.Delete(&message)
			continue
		}
		// Annotate the message via all enabled annotators
		for _, h := range enabledACARSAnnotators {
			log.Debug(Content("annotating message with annotator"), Note(" %s", h.Name()), Content(": "), Aside("%+v", message))
			result := h.AnnotateACARSMessage(message)
			if result != nil {
				result = h.SelectFields(result)
				annotations = MergeMaps(result, annotations)
			}
		}
		if len(annotations) == 0 {
			log.Info(Attention("no annotations were produced, not calling any receivers"))
		} else {
			for _, r := range enabledReceivers {
				log.Debug(Content("sending acars event to reciever "),
					Note(r.Name()),
					Content(": "),
					Aside("%+v", annotations))
				err := r.SubmitACARSAnnotations(annotations)
				if err != nil {
					log.Error(Attention("error submitting to %s, err: %v", r.Name(), err))
				}
			}
		}
		log.Debug(Content("message ending in \""),
			Note(Last20Characters(message.MessageText)),
			Content("\" took "),
			Note("%.2f seconds", time.Since(message.ProcessingStartedAt).Seconds()),
			Content(" to process and was ingested "),
			Note("%.2f seconds ago", time.Since(message.CreatedAt).Seconds()))
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
		db.Find(&message)
		if (time.Time{}.Equal(message.CreatedAt)) {
			log.Error(Attention("couldn't find message with id %d", id))
		}
		message.ProcessingStartedAt = time.Now()
		annotations := map[string]any{}
		ok, filters := VDLM2CriteriaFilter{}.Filter(message)
		if !ok {
			log.Info(Content("message was filtered out by %s", strings.Join(filters, ",")))
			log.Debug(Content("message ending in \""),
				Note(Last20Characters(message.VDL2.AVLC.ACARS.MessageText)),
				Content("\" took "),
				Note("%.2f seconds", time.Since(message.ProcessingStartedAt).Seconds()),
				Content(" to process and was ingested "),
				Note("%.2f seconds ago", time.Since(message.CreatedAt).Seconds()))
			db.Delete(&message)
			continue
		}
		// Annotate the message via all enabled annotators
		for _, h := range enabledVDLM2Annotators {
			log.Debug(Content("annotating message with annotator"), Note(" %s", h.Name()), Content(": "), Aside("%+v", message))
			result := h.AnnotateVDLM2Message(message)
			if result != nil {
				result = h.SelectFields(result)
				annotations = MergeMaps(result, annotations)
			}
		}
		if len(annotations) == 0 {
			log.Info(Attention("no annotations were produced, not calling any receivers"))
		} else {
			for _, r := range enabledReceivers {
				log.Debug(Content("sending vdlm2 event to reciever "),
					Note("%s", r.Name()),
					Content(": "),
					Aside("%+v", annotations))
				err := r.SubmitACARSAnnotations(annotations)
				if err != nil {
					log.Error(Attention("error submitting to %s, err: %v", r.Name(), err))
				}
			}
		}
		log.Debug(Content("message ending in \""),
			Note(Last20Characters(message.VDL2.AVLC.ACARS.MessageText)),
			Content("\" took "),
			Note("%.2f seconds", time.Since(message.ProcessingStartedAt).Seconds()),
			Content(" to process and was ingested "),
			Note("%.2f seconds ago", time.Since(message.CreatedAt).Seconds()))
		db.Delete(&message)
	}
}
