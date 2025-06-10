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

var (
	ACARSHubMaxConcurrentRequests = 1
	ACARSMessageQueue             = make(chan ACARSMessage, 10000)
	VDLM2MessageQueue             = make(chan VDLM2Message, 10000)
)

func ReadACARSHubACARSMessages() {
	if config.ACARSHub.MaxConcurrentRequests != 0 {
		ACARSHubMaxConcurrentRequests = config.ACARSHub.MaxConcurrentRequests
	}
	for range ACARSHubMaxConcurrentRequests {
		go HandleACARSJSONMessages(ACARSMessageQueue)
	}

	address := fmt.Sprintf("%s:%d", config.ACARSHub.ACARS.Host, config.ACARSHub.ACARS.Port)
	for {
		log.Debug(yo().INFODUMP("connecting to ").Hmm(config.ACARSHub.ACARS.Host).INFODUMP(" on acars json port ").Hmm(fmt.Sprint(config.ACARSHub.ACARS.Port)).FRFR())
		s, err := net.Dial("tcp", address)
		if err != nil {
			log.Error(yo().Uhh("error connecting to acars json: %v", err).FRFR())
			time.Sleep(time.Second * 1)
			continue
		}
		log.Info(yo().Bet("connected to acarshub acars json port successfully").FRFR())
		readJson := json.NewDecoder(io.Reader(s))
		log.Debug(yo().INFODUMP("handling acars json messages").FRFR())
		for {
			var next ACARSMessage
			if err := readJson.Decode(&next); err != nil {
				// Might have connection issues, exit to reconnect
				log.Error(yo().Uhh("error decoding acars message: %s", err).FRFR())
				break
			}
			log.Info(yo().FYI("new acars message received").FRFR())
			if (next == ACARSMessage{}) {
				log.Error(
					yo().Uhh("json message did not match expected structure, we got: ").
						BTW("%+v", fmt.Sprintf("%+v", next)).FRFR(),
				)
				continue
			} else {
				log.Debug(yo().FYI("new acars message content ").
					Hmm(fmt.Sprintf("(%d already in queue)", len(ACARSMessageQueue))).
					FYI(": ").INFODUMP(fmt.Sprintf("%+v", next)).FRFR())
				ACARSMessageQueue <- next
				continue
			}
		}

		log.Warn(yo().Uhh("acars handler exited, reconnecting").FRFR())
		s.Close()
		time.Sleep(time.Second * 1)
	}
}

func ReadACARSHubVDLM2Messages() {
	if config.ACARSHub.MaxConcurrentRequests != 0 {
		ACARSHubMaxConcurrentRequests = config.ACARSHub.MaxConcurrentRequests
	}
	for range ACARSHubMaxConcurrentRequests {
		go HandleVDLM2JSONMessages(VDLM2MessageQueue)
	}

	address := fmt.Sprintf("%s:%d", config.ACARSHub.VDLM2.Host, config.ACARSHub.VDLM2.Port)
	for {
		log.Debug(yo().INFODUMP("connecting to ").Hmm(config.ACARSHub.VDLM2.Host).INFODUMP(" on vdlm2 json port ").Hmm(fmt.Sprint(config.ACARSHub.VDLM2.Port)).FRFR())
		s, err := net.Dial("tcp", address)
		if err != nil {
			log.Error(yo().Uhh("error connecting to vdlm2 json: %v", err).FRFR())
			time.Sleep(time.Second * 1)
			break
		}
		log.Info(yo().Bet("connected to acarshub vdlm2 json port successfully").FRFR())
		readJson := json.NewDecoder(io.Reader(s))
		log.Debug(yo().INFODUMP("handling vdlm2 json messages").FRFR())
		for {
			var next VDLM2Message
			if err := readJson.Decode(&next); err != nil {
				// Might have connection issues, exit to reconnect
				log.Error(yo().Uhh("error decoding vdlm2 message: %s", err).FRFR())
				break
			}
			log.Info(yo().FYI("new vdlm2 message received").FRFR())
			if (next == VDLM2Message{}) {
				log.Error(yo().Uhh("json message did not match expected structure, we got: %+v", next).FRFR())
				continue
			} else {
				log.Debug(yo().FYI("new vdlm2 message content ").
					Hmm(fmt.Sprintf("(%d already in queue)", len(VDLM2MessageQueue))).
					FYI(": ").INFODUMP(fmt.Sprintf("%+v", next)).FRFR())
				VDLM2MessageQueue <- next
				continue
			}
		}

		log.Warn(yo().Uhh("vdlm2 handler exited, reconnecting").FRFR())
		s.Close()
		time.Sleep(time.Second * 1)
	}
}

// Connects to ACARS and starts listening to messages
func SubscribeToACARSHub() {
	launched := false
	if config.ACARSHub.ACARS.Host != "" && config.ACARSHub.ACARS.Port != 0 {
		go ReadACARSHubACARSMessages()
		launched = true
	}
	if config.ACARSHub.VDLM2.Host != "" && config.ACARSHub.VDLM2.Port != 0 {
		go ReadACARSHubVDLM2Messages()
		launched = true
	}
	if !launched {
		log.Warn(yo().Uhh("no acarshub subscribers set, please check configuration (%s).FRFR()", configFilePath))
	} else {
		log.Debug(yo().INFODUMP("launched acarshub subscribers").FRFR())
	}
}

// Reads messages in the channel from ReadACARSHubVDLM2Messages, annotates and
// sends off to configured receivers
func HandleACARSJSONMessages(ACARSMessageQueue chan ACARSMessage) {
	for message := range ACARSMessageQueue {
		annotations := map[string]any{}
		ok, filters := ACARSCriteriaFilter{}.Filter(message)
		if !ok {
			fd := strings.Join(filters, ",")
			log.Info(yo().FYI("message was filtered out by %s", fd).FRFR())
			continue
		}
		// Annotate the message via all enabled annotators
		for _, h := range enabledACARSAnnotators {
			log.Debug(yo().FYI("annotating message with annotator").Hmm(" %s", h.Name()).FYI(": ").INFODUMP(fmt.Sprintf("%+v", message)).FRFR())
			result := h.AnnotateACARSMessage(message)
			if result != nil {
				result = h.SelectFields(result)
				annotations = MergeMaps(result, annotations)
			}
		}
		if len(annotations) == 0 {
			log.Info(yo().Uhh("no annotations were produced, not calling any receivers").FRFR())
		} else {
			for _, r := range enabledReceivers {
				log.Debug(yo().FYI("sending acars event to reciever ").
					Hmm(r.Name()).
					FYI(": ").
					INFODUMP("%+v", annotations).FRFR(),
				)
				err := r.SubmitACARSAnnotations(annotations)
				if err != nil {
					log.Error(yo().Uhh("error submitting to %s, err: %v", r.Name(), err).FRFR())
				}
			}
		}
	}
}

// Reads messages in the channel from ReadACARSHubACARSMessages, annotates and
// sends off to configured receivers
func HandleVDLM2JSONMessages(VDLM2MessageQueue chan VDLM2Message) {
	for message := range VDLM2MessageQueue {
		annotations := map[string]any{}
		ok, filters := VDLM2CriteriaFilter{}.Filter(message)
		if !ok {
			log.Info(yo().FYI("message was filtered out by %s", strings.Join(filters, ",")).FRFR())
			continue
		}
		// Annotate the message via all enabled annotators
		for _, h := range enabledVDLM2Annotators {
			log.Debug(yo().FYI("annotating message with annotator").Hmm(" %s", h.Name()).FYI(": ").INFODUMP(fmt.Sprintf("%+v", message)).FRFR())
			result := h.AnnotateVDLM2Message(message)
			if result != nil {
				result = h.SelectFields(result)
				annotations = MergeMaps(result, annotations)
			}
		}
		if len(annotations) == 0 {
			log.Info(yo().Uhh("no annotations were produced, not calling any receivers").FRFR())
		} else {
			for _, r := range enabledReceivers {
				log.Debug(yo().INFODUMP("sending vdlm2 event to reciever ").
					Hmm("%s", r.Name()).
					INFODUMP(": ").
					INFODUMP("%s", annotations).FRFR())
				err := r.SubmitACARSAnnotations(annotations)
				if err != nil {
					log.Error(yo().Uhh("error submitting to %s, err: %v", r.Name(), err).FRFR())
				}
			}
		}
	}
}
