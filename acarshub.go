package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

func ReadACARSHubACARSMessages() {
	address := fmt.Sprintf("%s:%d", config.ACARSHubHost, config.ACARSHubPort)
	log.Debugf("connecting to %s acars json port", address)
	s, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("error connecting to acars json: %v", err)
	}
	defer s.Close()
	log.Info("connected to acarshub acars json port successfully")
	r := io.Reader(s)
	HandleACARSJSONMessages(&r)
}

func ReadACARSHubVDLM2Messages() {
	address := fmt.Sprintf("%s:%d", config.ACARSHubHost, config.ACARSHubVDLM2Port)
	log.Debugf("connecting to %s vdlm2 json port", address)
	s, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("error connecting to vdlm2 json: %v", err)
	}
	defer s.Close()
	log.Info("connected to acarshub vdlm2 json port successfully")
	r := io.Reader(s)
	HandleVDLM2JSONMessages(&r)
}

// Connects to ACARS and starts listening to messages
func SubscribeToACARSHub() {
	go ReadACARSHubACARSMessages()
	go ReadACARSHubVDLM2Messages()
}

// Reads messages from the ACARSHub connection and annotates, then sends
func HandleACARSJSONMessages(r *io.Reader) {
	readJson := json.NewDecoder(*r)
	log.Debug("handling acars json messages")
	for {
		annotations := map[string]any{}
		var next ACARSMessage
		err := readJson.Decode(&next)
		if err != nil {
			log.Warnf("error decoding acars message: %s", err)
			continue
		}
		log.Info("new acars message received")
		if (next == ACARSMessage{}) {
			log.Errorf("json message did not match expected structure, we got: %+v", next)
			continue
		} else {

			log.Debugf("new acars message content: %+v", next)
			ok, filters := ACARSCriteriaFilter{}.Filter(next)
			if !ok {
				log.Infof("message was filtered out by %s", strings.Join(filters, ","))
				continue
			}
			// Annotate the message via all enabled annotators
			for _, h := range enabledAnnotators {
				log.Debugf("sending event to annotator %s: %+v", h.Name(), next)
				result := h.AnnotateACARSMessage(next)
				if result != nil {
					result = h.SelectFields(result)
					annotations = MergeMaps(result, annotations)
				}
			}
		}
	}
}

// Reads messages from the ACARSHub connection and annotates, then sends
func HandleVDLM2JSONMessages(r *io.Reader) {
	readJson := json.NewDecoder(*r)
	log.Debug("handling vdlm2 json messages")
	for {
		annotations := map[string]any{}
		var next VDLM2Message
		// Decode consumes the buffer, so we use a second decoder
		err := readJson.Decode(&next)
		if err != nil {
			log.Warnf("error decoding vdlm2 message: %s", err)
			// Empty buffer for nextJson
			readJson.Decode(&next)
			continue
		}
		log.Info("new vdlm2 message received")
		if (next == VDLM2Message{}) {
			log.Errorf("json message did not match expected structure, we got: %+v", next)
			continue
		}
		log.Debugf("new vdlm2 message content: %+v", next)
		ok, filters := VDLM2CriteriaFilter{}.Filter(next)
		if !ok {
			log.Infof("message was filtered out by %s", strings.Join(filters, ","))
			continue
		} // Annotate the message via all enabled annotators
		for _, h := range enabledAnnotators {
			log.Debugf("sending event to annotator %s: %+v", h.Name(), next)
			result := h.AnnotateVDLM2Message(next)
			if result != nil {
				result = h.SelectFields(result)
				annotations = MergeMaps(result, annotations)
			}
		}
		for _, r := range enabledReceivers {
			log.Debugf("sending event to reciever %s: %+v", r.Name(), annotations)
			err := r.SubmitACARSAnnotations(annotations)
			if err != nil {
				log.Errorf("error submitting to %s, err: %v", r.Name(), err)
			}
		}
	}
}
