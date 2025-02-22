package main

import (
	"bytes"
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
	Payload any
}

func (w WebhookHandlerReciever) Name() string {
	return "Webhook Handler"
}

// Submits an ACARS Message to a webhook after transforming via template
func (w WebhookHandlerReciever) SubmitACARSMessage(m AnnotatedACARSMessage) error {
	// Hardcoded to be a simple payload compatible with Discord at this time
	msgTemplate := `{"content": "# [{{ .ACARSMessage.AircraftTailCode }}](` + FlightAwareRoot + m.AircraftTailCode + `)\n` +
		`**Flight Number**: {{ .ACARSMessage.FlightNumber }}\n` +
		`**Signal**: {{ .ACARSMessage.SignaldBm }} dBm\n` +
		`**Distance**: {{ (index .Annotations 0).Annotation.adsbAircraftDistanceMi }} miles\n` +
		`"}`
	// Initialize the template
	t, err := template.New("webhook").Parse(msgTemplate)
	if err != nil {
		return err
	}

	// Create a destination for the result
	var b bytes.Buffer
	// Evaluate the template
	err = t.Execute(&b, m)
	if err != nil {
		return err
	}

	log.Debugf("webhook payload: %s", b.String())

	method := "GET"
	if config.WebhookMethod != "" {
		method = config.WebhookMethod
	}

	req, err := http.NewRequest(method, config.WebhookURL, &b)
	req.Header.Add("User-Agent", WebhookUserAgent)
	// Hardcoded for now because most webhooks will be JSON
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
		log.Error(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Debugf("webhook returned: %s", string(body))
	return nil
}
