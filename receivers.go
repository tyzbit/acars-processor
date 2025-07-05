package main

import log "github.com/sirupsen/logrus"

func ConfigureReceivers() {
	// Add receivers based on what's enabled
	if config.Receivers.Webhook.Enabled {
		log.Info(Success("Webhook receiver enabled"))
		enabledReceivers = append(enabledReceivers, WebhookHandlerReciever{})
	}
	if config.Receivers.NewRelic.Enabled {
		log.Info(Success("New Relic reciever enabled"))
		enabledReceivers = append(enabledReceivers, NewRelicHandlerReciever{})
	}
	if config.Receivers.DiscordWebhook.Enabled {
		log.Info(Success("Discord reciever enabled"))
		enabledReceivers = append(enabledReceivers, DiscordHandlerReciever{})
	}
	if len(enabledReceivers) == 0 {
		log.Warn(Attention("no receivers are enabled"))
	}
}
