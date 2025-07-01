package handler

import (
	"slices"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tyzbit/acars-processor/acarshub"
	"github.com/tyzbit/acars-processor/annotator"
	. "github.com/tyzbit/acars-processor/config"
	. "github.com/tyzbit/acars-processor/database"
	. "github.com/tyzbit/acars-processor/decorate"
	"github.com/tyzbit/acars-processor/filter"
	"github.com/tyzbit/acars-processor/receiver"
	"github.com/tyzbit/acars-processor/util"
	"gorm.io/gorm"
)

func SelectFields(annotation annotator.Annotation) annotator.Annotation {
	if Config.Annotators.ACARS.SelectedFields == nil {
		return annotation
	}
	selectedFields := annotator.Annotation{}
	for field, value := range annotation {
		if slices.Contains(Config.Annotators.ACARS.SelectedFields, field) {
			selectedFields[field] = value
		}
	}
	return selectedFields
}

func HandleACARSHub() {
	for range Config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests {
		go HandleACARSJSONMessages(acarshub.ACARSMessageQueue)
	}

	for range Config.ACARSProcessorSettings.ACARSHub.MaxConcurrentRequests {
		go HandleVDLM2JSONMessages(acarshub.VDLM2MessageQueue)
	}
}

// Reads messages in the channel from ReadACARSHubVDLM2Messages, annotates and
// sends off to configured receivers
func HandleACARSJSONMessages(ACARSMessageQueue chan uint) {
	for id := range ACARSMessageQueue {
		// Create a message with the ID we're looking for
		message := acarshub.ACARSMessage{
			Model: gorm.Model{
				ID: id,
			},
		}
		// Find that message
		DB.Where(&message).Find(&message)
		if (time.Time{}.Equal(message.CreatedAt)) {
			log.Error(Attention("couldn't find message with id %d", id))
		}
		message.ProcessingStartedAt = time.Now()
		annotations := map[string]any{}
		ok, filters := filter.ACARSCriteriaFilter{}.Filter(message)
		if !ok {
			fd := strings.Join(filters, ",")
			log.Info(Content("message was filtered out by %s", fd))
			log.Debug(Content("message ending in \""),
				Note(util.Last20Characters(message.MessageText)),
				Content("\" took "),
				Note("%.2f seconds", time.Since(message.ProcessingStartedAt).Seconds()),
				Content(" to process and was ingested "),
				Note("%.2f seconds ago", time.Since(message.CreatedAt).Seconds()))
			DB.Delete(&message)
			continue
		}
		// Annotate the message via all enabled annotators
		for _, h := range annotator.EnabledAnnotators.ACARS {
			log.Debug(Content("annotating message with annotator"), Note(" %s", h.Name()), Content(": "), Aside("%+v", message))
			result := h.AnnotateACARSMessage(message)
			if result != nil {
				result = SelectFields(result)
				annotations = util.MergeMaps(result, annotations)
			}
		}
		if len(annotations) == 0 {
			log.Info(Attention("no annotations were produced, not calling any receivers"))
		} else {
			for _, r := range *receiver.EnabledReceivers {
				log.Debug(Content("sending acars event to reciever "),
					Note(r.Name()),
					Content(": "),
					Aside("%+v", annotations))
				err := r.SubmitACARSAnnotations(annotations)
				if err != nil {
					log.Error(Attention("error submitting to %s, err: %v", r.Name(), err))
				}
			}
		}
		log.Debug(Content("message ending in \""),
			Note(util.Last20Characters(message.MessageText)),
			Content("\" took "),
			Note("%.2f seconds", time.Since(message.ProcessingStartedAt).Seconds()),
			Content(" to process and was ingested "),
			Note("%.2f seconds ago", time.Since(message.CreatedAt).Seconds()))
		DB.Delete(&message)
	}
}

// Reads messages in the channel from ACARSMessages, annotates and
// sends off to configured receivers
func HandleVDLM2JSONMessages(VDLM2MessageQueue chan uint) {
	for id := range VDLM2MessageQueue {
		// Create a message with the ID we're looking for
		message := acarshub.VDLM2Message{
			Model: gorm.Model{
				ID: id,
			},
		}
		// Find that message
		DB.Find(&message)
		if (time.Time{}.Equal(message.CreatedAt)) {
			log.Error(Attention("couldn't find message with id %d", id))
		}
		message.ProcessingStartedAt = time.Now()
		annotations := map[string]any{}
		ok, filters := filter.VDLM2CriteriaFilter{}.Filter(message)
		if !ok {
			log.Info(Content("message was filtered out by %s", strings.Join(filters, ",")))
			log.Debug(Content("message ending in \""),
				Note(util.Last20Characters(message.VDL2.AVLC.ACARS.MessageText)),
				Content("\" took "),
				Note("%.2f seconds", time.Since(message.ProcessingStartedAt).Seconds()),
				Content(" to process and was ingested "),
				Note("%.2f seconds ago", time.Since(message.CreatedAt).Seconds()))
			DB.Delete(&message)
			continue
		}
		// Annotate the message via all enabled annotators
		for _, h := range annotator.EnabledAnnotators.VDLM2 {
			log.Debug(Content("annotating message with annotator"), Note(" %s", h.Name()), Content(": "), Aside("%+v", message))
			result := h.AnnotateVDLM2Message(message)
			if result != nil {
				result = SelectFields(result)
				annotations = util.MergeMaps(result, annotations)
			}
		}
		if len(annotations) == 0 {
			log.Info(Attention("no annotations were produced, not calling any receivers"))
		} else {
			for _, r := range *receiver.EnabledReceivers {
				log.Debug(Content("sending vdlm2 event to reciever "),
					Note("%s", r.Name()),
					Content(": "),
					Aside("%+v", annotations))
				err := r.SubmitACARSAnnotations(annotations)
				if err != nil {
					log.Error(Attention("error submitting to %s, err: %v", r.Name(), err))
				}
			}
		}
		log.Debug(Content("message ending in \""),
			Note(util.Last20Characters(message.VDL2.AVLC.ACARS.MessageText)),
			Content("\" took "),
			Note("%.2f seconds", time.Since(message.ProcessingStartedAt).Seconds()),
			Content(" to process and was ingested "),
			Note("%.2f seconds ago", time.Since(message.CreatedAt).Seconds()))
		DB.Delete(&message)
	}
}
