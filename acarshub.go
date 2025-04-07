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

func ReadACARSHubACARSMessages() {
	if !config.AnnotateACARS {
		return
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
		r := io.Reader(s)
		log.Debug("handling acars json messages")
		HandleACARSJSONMessages(&r)
		log.Warn("acars handler exited, reconnecting")
		s.Close()
		time.Sleep(time.Second * 1)
	}
}

func ReadACARSHubVDLM2Messages() {
	if !config.AnnotateVDLM2 {
		return
	}
	address := fmt.Sprintf("%s:%d", config.ACARSHubVDLM2Host, config.ACARSHubVDLM2Port)
	for {
		log.Debugf("connecting to %s vdlm2 json port", address)
		s, err := net.Dial("tcp", address)
		if err != nil {
			log.Errorf("error connecting to vdlm2 json: %v", err)
			time.Sleep(time.Second * 1)
			continue
		}
		log.Info("connected to acarshub vdlm2 json port successfully")
		r := io.Reader(s)
		log.Debug("handling vdlm2 json messages")
		HandleVDLM2JSONMessages(&r)
		log.Warn("vdlm2 handler exited, reconnecting")
		s.Close()
		time.Sleep(time.Second * 1)
	}
}

// Connects to ACARS and starts listening to messages
func SubscribeToACARSHub() {
	go ReadACARSHubACARSMessages()
	go ReadACARSHubVDLM2Messages()
}

// Reads messages from the ACARSHub connection and annotates, then sends to
// receivers. Called again if returned from.
func HandleACARSJSONMessages(r *io.Reader) {
	for {
		readJson := json.NewDecoder(*r)
		annotations := map[string]any{}
		var next ACARSMessage
		if err := readJson.Decode(&next); err != nil {
			// Might have connection issues, exit to reconnect
			log.Errorf("error decoding acars message: %s", err)
			return
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
			for _, h := range enabledACARSAnnotators {
				log.Debugf("sending event to annotator %s: %+v", h.Name(), next)
				result := h.AnnotateACARSMessage(next)
				if result != nil {
					result = h.SelectFields(result)
					annotations = MergeMaps(result, annotations)
				}
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

// Reads messages from the ACARSHub connection and annotates, then sends
func HandleVDLM2JSONMessages(r *io.Reader) {
	for {
		readJson := json.NewDecoder(*r)
		annotations := map[string]any{}
		var next VDLM2Message
		// Decode consumes the buffer, so we use a second decoder
		if err := readJson.Decode(&next); err != nil {
			// Might have connection issues, exit to reconnect
			log.Errorf("error decoding vdlm2 message: %s", err)
			return
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
		} // Annotate the message via all enabled VDLM2 annotators
		for _, h := range enabledVDLM2Annotators {
			log.Debugf("sending event to annotator %s: %+v", h.Name(), next)
			result := h.AnnotateVDLM2Message(next)
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
