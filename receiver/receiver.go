package receiver

import (
	log "github.com/sirupsen/logrus"
	"github.com/tyzbit/acars-processor/annotator"
	. "github.com/tyzbit/acars-processor/config"
	. "github.com/tyzbit/acars-processor/decorate"
)

type Receiver interface {
	SubmitACARSAnnotations(annotator.Annotation) error
	Name() string
}

var EnabledReceivers = new([]Receiver)

func ConfigureReceivers() {
	// Add receivers based on what's enabled
	if Config.Receivers.Webhook.URL != "" {
		log.Info(Success("Webhook receiver enabled"))
		*EnabledReceivers = append(*EnabledReceivers, WebhookHandlerReciever{})
	}
	if Config.Receivers.NewRelic.APIKey != "" {
		log.Info(Success("New Relic reciever enabled"))
		*EnabledReceivers = append(*EnabledReceivers, &NewRelicHandlerReciever{})
	}
	if Config.Receivers.DiscordWebhook.URL != "" {
		log.Info(Success("Discord reciever enabled"))
		*EnabledReceivers = append(*EnabledReceivers, &DiscordHandlerReciever{})
	}
	if len(*EnabledReceivers) == 0 {
		log.Warn(Attention("no receivers are enabled"))
	}
}
