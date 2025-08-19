package main

import (
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
)

// Reads ACARS-Processor messages, filters annotates and
// sends off to configured receivers
func HandleAPMessageQueue(apm chan APMessageQeueueItem) {
	workerCount := config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests
	if workerCount <= 0 {
		workerCount = 1 // fallback safety
	}

	// Worker function
	worker := func() {
		for message := range apm {
			start := time.Now()
			// Iterate through each step and execute every filter, annotator
			// and receiver.
			for stepNum, s := range config.Steps {
				// Make it human-friendly
				stepNum++
				if !reflect.DeepEqual(s.Filter, FilterStep{}) {
					name, filter, err := s.Filter.Filter(message.APMessage)
					if err != nil {
						log.Warn(Attention("error filtering with %s in step number %d: %s", name, stepNum, err))
						continue
					}
					if filter {
						log.Info(Content("message was filtered out by %s filterer in step number %d", name, stepNum))
						break
					}
				}
				if !reflect.DeepEqual(s.Annotate, AnnotateStep{}) {
					message.APMessage = s.Annotate.Annotate(message.APMessage)
				}
				if !reflect.DeepEqual(s.Send, ReceiverStep{}) {
					err := s.Send.Send(message.APMessage)
					if err != nil {
						log.Warn(Attention("error when calling receiver(s): %s", err))
					}
				}
			}
			mt := GetAPMessageCommonFieldAsString(message.APMessage, "message_text")
			ts := GetAPMessageCommonFieldAsFloat64(message.APMessage, "timestamp")
			msgts := time.UnixMilli(int64(ts))
			log.Info(Content("message ending in \""),
				Note(Last20Characters(mt)),
				Content("\" took "),
				Note("%.2f seconds", time.Since(start).Seconds()),
				Content(" to process and was ingested "),
				Note("%.2f seconds ago", time.Since(msgts).Seconds()))
			if !(reflect.DeepEqual(message.ACARSMessage, ACARSMessage{})) {
				message.ACARSMessage.ProcessingFinishedAt = time.Now()
				message.ACARSMessage.Processed = true
				db.Updates(&message.ACARSMessage)
			}
			if !(reflect.DeepEqual(message.VDLM2Message, VDLM2Message{})) {
				message.VDLM2Message.ProcessingFinishedAt = time.Now()
				message.VDLM2Message.Processed = true
				db.Updates(&message.VDLM2Message)
			}
		}
	}
	// Start workers (they run forever, waiting for channel messages)
	for i := 0; i < workerCount; i++ {
		go worker()
	}
}
