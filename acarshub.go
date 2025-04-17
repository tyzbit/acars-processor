package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var ACARSHubMaxConcurrentRequests = 1

func ReadACARSHubACARSMessages() {
	var achan = make(chan ACARSMessage, 1000)
	if config.ACARSHubMaxConcurrentRequests != 0 {
		ACARSHubMaxConcurrentRequests = config.ACARSHubMaxConcurrentRequests
	}
	for i := 0; i < ACARSHubMaxConcurrentRequests; i++ {
		go HandleACARSJSONMessages(achan)
	}

	address := fmt.Sprintf("%s:%d", config.ACARSHubHost, config.ACARSHubPort)
	for {
		log.Debugf("connecting to %s acars json port", address)
		s, err := net.Dial("tcp", address)
		if err != nil {
			log.Errorf("error connecting to acars json: %v", err)
			time.Sleep(time.Second * 1)
			continue
		}
		log.Info("connected to acarshub acars json port successfully")
		readJson := json.NewDecoder(io.Reader(s))
		log.Debug("handling acars json messages")
		for {
			var next ACARSMessage
			if err := readJson.Decode(&next); err != nil {
				// Might have connection issues, exit to reconnect
				log.Errorf("error decoding acars message: %s", err)
				break
			}
			log.Info("new acars message received")
			if (next == ACARSMessage{}) {
				log.Errorf("json message did not match expected structure, we got: %+v", next)
				continue
			} else {
				log.Debugf("new acars message content (%d already in queue): %+v", len(achan), next)
				achan <- next
				continue
			}
		}

		log.Warn("acars handler exited, reconnecting")
		s.Close()
		time.Sleep(time.Second * 1)
	}
}

func ReadACARSHubVDLM2Messages() {
	var vchan = make(chan VDLM2Message, 1000)
	if config.ACARSHubMaxConcurrentRequests != 0 {
		ACARSHubMaxConcurrentRequests = config.ACARSHubMaxConcurrentRequests
	}
	for i := 0; i < ACARSHubMaxConcurrentRequests; i++ {
		go HandleVDLM2JSONMessages(vchan)
	}

	address := fmt.Sprintf("%s:%d", config.ACARSHubVDLM2Host, config.ACARSHubVDLM2Port)
	for {
		log.Debugf("connecting to %s vdlm2 json port", address)
		s, err := net.Dial("tcp", address)
		if err != nil {
			log.Errorf("error connecting to vdlm2 json: %v", err)
			time.Sleep(time.Second * 1)
			break
		}
		log.Info("connected to acarshub vdlm2 json port successfully")
		readJson := json.NewDecoder(io.Reader(s))
		log.Debug("handling vdlm2 json messages")
		for {
			var next VDLM2Message
			if err := readJson.Decode(&next); err != nil {
				// Might have connection issues, exit to reconnect
				log.Errorf("error decoding vdlm2 message: %s", err)
				break
			}
			log.Info("new vdlm2 message received")
			if (next == VDLM2Message{}) {
				log.Errorf("json message did not match expected structure, we got: %+v", next)
				continue
			} else {
				log.Debugf("new vdlm2 message content (%d already in queue): %+v", len(vchan), next)
				vchan <- next
				continue
			}
		}

		log.Warn("vdlm2 handler exited, reconnecting")
		s.Close()
		time.Sleep(time.Second * 1)
	}
}

// Connects to ACARS and starts listening to messages
func SubscribeToACARSHub() {
	if config.AnnotateACARS {
		go ReadACARSHubACARSMessages()
	}
	if config.AnnotateVDLM2 {
		go ReadACARSHubVDLM2Messages()
	}
	log.Debug("launched acarshub subscribers")
}

// Reads messages in the channel from ReadACARSHubVDLM2Messages, annotates and
// sends off to configured receivers
func HandleACARSJSONMessages(achan chan ACARSMessage) {
	for message := range achan {
		annotations := map[string]any{}
		ok, filters := ACARSCriteriaFilter{}.Filter(message)
		if !ok {
			log.Infof("message was filtered out by %s", strings.Join(filters, ","))
			continue
		}
		// Annotate the message via all enabled annotators
		for _, h := range enabledACARSAnnotators {
			log.Debugf("sending event to annotator %s: %+v", h.Name(), message)
			result := h.AnnotateACARSMessage(message)
			if result != nil {
				result = h.SelectFields(result)
				annotations = MergeMaps(result, annotations)
			}
		}
		for _, r := range enabledReceivers {
			log.Debugf("sending acars event to reciever %s: %+v", r.Name(), annotations)
			err := r.SubmitACARSAnnotations(annotations)
			if err != nil {
				log.Errorf("error submitting to %s, err: %v", r.Name(), err)
			}
		}
	}
}

// Reads messages in the channel from ReadACARSHubACARSMessages, annotates and
// sends off to configured receivers
func HandleVDLM2JSONMessages(vchan chan VDLM2Message) {
	for message := range vchan {
		annotations := map[string]any{}
		ok, filters := VDLM2CriteriaFilter{}.Filter(message)
		if !ok {
			log.Infof("message was filtered out by %s", strings.Join(filters, ","))
			continue
		}
		// Annotate the message via all enabled annotators
		for _, h := range enabledVDLM2Annotators {
			log.Debugf("sending event to annotator %s: %+v", h.Name(), message)
			result := h.AnnotateVDLM2Message(message)
			if result != nil {
				result = h.SelectFields(result)
				annotations = MergeMaps(result, annotations)
			}
		}
		for _, r := range enabledReceivers {
			log.Debugf("sending vdlm2 event to reciever %s: %+v", r.Name(), annotations)
			err := r.SubmitACARSAnnotations(annotations)
			if err != nil {
				log.Errorf("error submitting to %s, err: %v", r.Name(), err)
			}
		}
	}
}
