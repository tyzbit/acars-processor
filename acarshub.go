package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Connects to ACARS and starts listening to messages
func SubscribeToACARSHub() {
	address := fmt.Sprintf("%s:%d", config.ACARSHubHost, config.ACARSHubPort)
	log.Debugf("connecting to %s", address)
	s, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("error connecting: %v", err)
	}
	defer s.Close()
	log.Info("connected successfully")
	r := io.Reader(s)
	HandleACARSJSONMessages(&r)
}

// Reads messages from the ACARSHub connection and annotates, then sends
func HandleACARSJSONMessages(r *io.Reader) {
	var readBuff bytes.Buffer
	nextPreview := io.TeeReader(*r, &readBuff)
	previewJson := json.NewDecoder(nextPreview)
	readJson := json.NewDecoder(&readBuff)
	log.Debug("handling acarshub json messages")
	for {
		annotations := map[string]any{}
		var next map[string]any
		// Decode consumes the buffer, so we use a second decoder
		err := previewJson.Decode(&next)
		if err != nil {
			log.Warnf("error decoding acars message: %s", err)
			// Empty buffer for nextJson
			readJson.Decode(&next)
			continue
		}
		switch next["vdl2"] {
		case nil:
			log.Info("new acars message received")
			var nextAcars ACARSMessage
			readJson.Decode(&nextAcars)
			if (nextAcars == ACARSMessage{}) {
				log.Errorf("json message did not match expected structure, we got: %+v", nextAcars)
				continue
			} else {

				log.Debugf("new acars message content: %+v", nextAcars)
				ok, filters := ACARSCriteriaFilter{}.Filter(nextAcars)
				if !ok {
					log.Infof("message was filtered out by %s", strings.Join(filters, ","))
					continue
				}
				// Annotate the message via all enabled annotators
				for _, h := range enabledAnnotators {
					log.Debugf("sending event to annotator %s: %+v", h.Name(), nextAcars)
					result := h.AnnotateACARSMessage(nextAcars)
					if result != nil {
						result = h.SelectFields(result)
						annotations = MergeMaps(result, annotations)
					}
				}
			}
		default:
			var nextVdlm2 VDLM2Message
			log.Info("new vdlm2 message received")
			err := readJson.Decode(&nextVdlm2)
			if err != nil {
				log.Errorf("error decoding vdlm2 message: %v", err)
				continue
			}
			log.Debugf("new vdlm2 message content: %+v", nextVdlm2)
			ok, filters := VDLM2CriteriaFilter{}.Filter(nextVdlm2)
			if !ok {
				log.Infof("message was filtered out by %s", strings.Join(filters, ","))
				continue
			} // Annotate the message via all enabled annotators
			for _, h := range enabledAnnotators {
				log.Debugf("sending event to annotator %s: %+v", h.Name(), nextVdlm2)
				result := h.AnnotateVDLM2Message(nextVdlm2)
				if result != nil {
					result = h.SelectFields(result)
					annotations = MergeMaps(result, annotations)
				}
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
