package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	log "github.com/sirupsen/logrus"
)

func SubscribeToACARSHub() {
	address := fmt.Sprintf("%s:%s", config.ACARSHost, config.ACARSPort)
	log.Debugf("connecting to %s via %s", address, config.ACARSTransport)
	s, err := net.Dial(config.ACARSTransport, address)
	if err != nil {
		log.Fatalf("error connecting: %v", err)
	}
	log.Info("connected successfully")
	defer s.Close()
	r := io.ReadCloser(s)
	j := json.NewDecoder(r)
	HandleACARSJSONMessages(j)
}

func HandleACARSJSONMessages(j *json.Decoder) {
	log.Debug("handling acars json messages")
	var next ACARSMessage
	for {
		_ = j.Decode(&next)
		if (next == ACARSMessage{}) {
			log.Error("invalid acars message")
		} else {
			log.Debugf("new acars message: %v", next)
			for _, h := range handlers {
				result := h.HandleACARSMessage(next)
				if result != "" {
					log.Info(result)
				}
			}
		}
	}
}
