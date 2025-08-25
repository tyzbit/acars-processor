package main

import (
	"reflect"
	"time"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

const (
	emptyStringRegex    = `^\s*$`
	nonEmptyStringRegex = `\S`
)

var formatFilterAction = map[bool]string{
	true:  Custom(color.New(color.FgYellow), "filtered"),
	false: Custom(color.New(color.FgCyan), "finished"),
}

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
			var filter bool
			// Iterate through each step and execute every filter, annotator
			// and receiver.
			var name string
			var err error
			exitStep := len(config.Steps) + 1
			for stepNum, s := range config.Steps {
				// Make it human-friendly
				stepNum++
				if !reflect.DeepEqual(s.Filter, FilterStep{}) {
					name, filter, err = s.Filter.Filter(message.APMessage)
					if err != nil {
						// The filters take FilterOnFailure into account, so we
						// only warn here.
						log.Warn(Attention("error filtering with %s in step number %d: %s", name, stepNum, err))
					}
					if filter {
						exitStep = stepNum
						break
					}
				}
				if !reflect.DeepEqual(s.Annotate, AnnotateStep{}) {
					message.APMessage = s.Annotate.Annotate(message.APMessage)
				}
				if !reflect.DeepEqual(s.Send, ReceiverStep{}) {
					err := s.Send.Send(message.APMessage)
					if err != nil {
						log.Warn(Attention("error sending to receivers in step number %d: %s", stepNum, err))
					}
				}
			}
			mt := GetAPMessageCommonFieldAsString(message.APMessage, "MessageText")
			ts := GetAPMessageCommonFieldAsInt64(message.APMessage, "UnixTimestamp")
			msgts := time.Unix(ts, 0)
			filterActionDescription := Content(",")
			// Only include filter information in the upcoming log message
			// if the message was filtered
			if name != "" {
				filterActionDescription = Content(" by ") +
					Emphasised(name)
			}
			var textPreview string
			if len(mt) > 0 {
				textPreview = Content("message ending in \"") +
					Note(Last20Characters(mt)) +
					Content("\"")

			} else {
				textPreview = Content("message with blank or empty message text")
			}
			log.Info(textPreview,
				Content(" was "),
				formatFilterAction[filter],
				Content(" in step %d", exitStep),
				filterActionDescription,
				Content(" took "),
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
