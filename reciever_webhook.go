package main

import (
	"bytes"
	"io"
	"net/http"
	"text/template"

	log "github.com/sirupsen/logrus"
)

type WebhookHandlerReciever struct {
	Payload any
}

// Must satisfy Receiver interface
func (n WebhookHandlerReciever) Name() string {
	return "webhook"
}

// Must satisfy Receiver interface
func (n WebhookHandlerReciever) SubmitACARSAnnotations(a Annotation) (err error) {
	if config.Receivers.Webhook.URL == "" {
		log.Panic(Attention("Webhook URL not specified"))
	}
	t, err := template.ParseFiles("receiver_webhook.tpl")
	if err != nil {
		log.Fatal(Attention(err))
	}

	var b bytes.Buffer
	err = t.Execute(&b, a)
	if err != nil {
		log.Error(Attention("error executing template, err: %v", err))
		return
	}

	h, err := http.NewRequest(config.Receivers.Webhook.Method, config.Receivers.Webhook.URL, &b)
	if err != nil {
		return err
	}
	for _, header := range config.Receivers.Webhook.Headers {
		h.Header.Add(header.Name, header.Value)
	}

	c := http.Client{}
	log.Info(Content("calling custom webhook"))
	resp, err := c.Do(h)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Debug(Aside("webhook returned: %s", string(body)))
	return nil
}
