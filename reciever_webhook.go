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
	t, err := template.ParseFiles("receiver_webhook.tpl")
	if err != nil {
		log.Panic(err)
	}

	var b bytes.Buffer
	err = t.Execute(&b, a)
	if err != nil {
		log.Error(yo.Uhh("error executing template, err: %v", err).FRFR())
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
	log.Info(yo.FYI("calling custom webhook").FRFR())
	resp, err := c.Do(h)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Debug(yo.INFODUMP("webhook returned: %s", string(body)))
	return nil
}
