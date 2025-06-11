package main

import (
	log "github.com/sirupsen/logrus"
)

func ConfigureReceivers() {
	// Add receivers based on what's enabled
	if config.Receivers.Webhook.URL != "" {
		log.Info(yo.Bet("Webhook receiver enabled").FRFR())
		enabledReceivers = append(enabledReceivers, WebhookHandlerReciever{})
	}
	if config.Receivers.NewRelic.APIKey != "" {
		log.Info(yo.Bet("New Relic reciever enabled").FRFR())
		enabledReceivers = append(enabledReceivers, NewRelicHandlerReciever{})
	}
	if config.Receivers.DiscordWebhook.URL != "" {
		log.Info(yo.Bet("Discord reciever enabled").FRFR())
		enabledReceivers = append(enabledReceivers, DiscordHandlerReciever{})
	}
	if len(enabledReceivers) == 0 {
		log.Warn(yo.Uhh("no receivers are enabled").FRFR())
	}
}
