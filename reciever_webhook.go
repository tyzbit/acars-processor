package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"
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
		log.Errorf("error executing template, err: %v", err)
		return
	}

	h, err := http.NewRequest(config.WebhookMethod, config.WebhookURL, &b)
	for _, header := range strings.Split(config.WebhookHeaders, ",") {
		if header != "" {
			key := strings.Split(header, "=")[0]
			value := strings.Split(header, "=")[1]
			h.Header.Add(key, value)
		}
	}

	c := http.Client{}
	resp, err := c.Do(h)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Debugf("webhook returned: %s", string(body))
	return nil
}
