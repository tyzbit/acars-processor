package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

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
	r := io.ReadCloser(s)
	defer r.Close()
	j := json.NewDecoder(r)
	HandleACARSJSONMessages(j)
}

// Reads messages from the ACARSHub connection and annotates, then sends
func HandleACARSJSONMessages(j *json.Decoder) {
	log.Debug("handling acars json messages")
	var next ACARSMessage
	for {
		err := j.Decode(&next)
		if err != nil {
			log.Errorf("error decoding acars message: %v", err)
			return
		}
		if (next == ACARSMessage{}) {
			log.Errorf("json message did not match expected structure, we got: %+v", next)
		} else {
			log.Debugf("new acars message: %+v", next)
			annotations := map[string]any{}
			// Annotate the message via all enabled annotators
			for _, h := range enabledAnnotators {
				result := h.AnnotateACARSMessage(next)
				if result != nil {
					annotations = MergeMaps(result, annotations)
					log.Info(annotations)
				}
			}
			for _, r := range enabledReceivers {
				log.Debugf("submitting to %s", r.Name())
				err := r.SubmitACARSAnnotations(annotations)
				if err != nil {
					log.Errorf("error submitting to %s, err: %v", r.Name(), err)
				}
			}
		}
	}
}
