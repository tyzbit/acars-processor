package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
)

const (
	WebhookUserAgent = "github.com/tyzbit/acars-annotator"
)

type WebhookHandlerReciever struct {
	Payload interface{}
}

func (w WebhookHandlerReciever) Name() string {
	return "Webhook Handler"
}

// Submits an ACARS Message to a webhook after transforming via template
func (w WebhookHandlerReciever) SubmitACARSMessage(m AnnotatedACARSMessage) error {
	// Hardcoded to be a simple payload compatible with Discord
	msgTemplate := `{"content": "{{ .ACARSMessage.AircraftTailCode }}|{{ .ACARSMessage.FlightNumber }} - {{ (index .Annotations 0).Annotation.adsbAircraftDistanceMi }} miles"}`
	t, err := template.New("webhook").Parse(msgTemplate)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	err = t.Execute(&b, m)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(b.String())
	if err != nil {
		return err
	}
	log.Debugf("webhook payload: %s", b.String())

	method := "GET"
	if config.WebhookMethod != "" {
		method = config.WebhookMethod
	}

	reqBody := bytes.NewBuffer(payload)
	req, err := http.NewRequest(method, config.WebhookURL, reqBody)
	req.Header.Add("User-Agent", WebhookUserAgent)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}

	headers := []string{}
	if config.WebhookHeaders != "" {
		headers = strings.Split(config.WebhookHeaders, ",")
	}
	for _, h := range headers {
		req.Header.Add(strings.Split(h, "=")[0], strings.Split(h, "=")[1])
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Debugf("webhook returned: %s", string(body))
	return nil
}
