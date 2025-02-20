package main

type WebhookHandlerReciever struct {
	Payload interface{}
}

func (w WebhookHandlerReciever) SubmitACARSMessage(m AnnotatedACARSMessage) error {
	// TODO: fill out
	return nil
}

func (w WebhookHandlerReciever) Name() string {
	return "Webhook Handler"
}
