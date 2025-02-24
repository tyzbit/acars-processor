package main

import (
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
			log.Info("new acars message received")
			log.Debugf("new acars message content: %+v", next)
			ok, filters := ACARSCriteriaFilter{}.Filter(next)
			if !ok {
				log.Infof("message was filtered out by %s", strings.Join(filters, ","))
				continue
			}
			annotations := map[string]any{}
			// Annotate the message via all enabled annotators
			for _, h := range enabledAnnotators {
				log.Debugf("sending event to annotator %s: %+v", h.Name(), next)
				result := h.AnnotateACARSMessage(next)
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
			var annotators, receivers string
			for i, ann := range enabledAnnotators {
				if i == 0 {
					annotators = ann.Name()
				} else {
					annotators = annotators + ", " + ann.Name()
				}
			}
			for i, rec := range enabledReceivers {
				if i == 0 {
					receivers = rec.Name()
				} else {
					receivers = receivers + ", " + rec.Name()
				}
			}
			log.Infof("used annotators: %s", annotators)
			log.Infof("sent to recievers: %s", receivers)
		}
	}
}
