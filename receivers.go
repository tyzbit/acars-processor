package main

import log "github.com/sirupsen/logrus"

func ConfigureReceivers() {
	// Add receivers based on what's enabled
	if config.WebhookURL != "" {
		log.Info("Webhook receiver enabled")
		enabledReceivers = append(enabledReceivers, WebhookHandlerReciever{})
	}
	if config.NewRelicLicenseKey != "" {
		log.Info("New Relic reciever enabled")
		enabledReceivers = append(enabledReceivers, NewRelicHandlerReciever{})
	}
	if config.DiscordWebhookURL != "" {
		log.Info("Discord reciever enabled")
		enabledReceivers = append(enabledReceivers, DiscordHandlerReciever{})
	}
	if len(enabledReceivers) == 0 {
		log.Warn("no receivers are enabled")
	}
}
