package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"text/template"

	log "github.com/sirupsen/logrus"
)

type WebHookReceiver struct {
	Module
	Receiver
	// URL, including port and params, to the desired webhook.
	URL string `jsonschema:"required,example=https://webhook:8443/webhook/?enable_feature=yes" default:"https://webhook:8443/webhook/?enable_feature=yes"`
	// Method when calling webhook (GET,POST,PUT etc).
	Method string `jsonschema:"required,default=POST" default:"POST"`
	// Additional headers to send along with the request.
	Headers []WebHookReceiverHeaders // The default for this is set in schema.go
}

type WebHookReceiverHeaders struct {
	// Header name.
	Name string `jsonschema:"required"` // NOTE: defaults are set in the function in schema.go
	// Header value.
	Value string `jsonschema:"required"` // NOTE: defaults are set in the function in schema.go
}

// Must satisfy Receiver interface
func (w WebHookReceiver) Name() string {
	return reflect.TypeOf(w).Name()
}

// Must satisfy Receiver interface
func (w WebHookReceiver) SubmitACARSAnnotations(a APMessage) (err error) {
	if w.URL == "" {
		log.Panic(Attention("Webhook URL not specified"))
	}
	t, err := template.ParseFiles("receiver_webhook.tpl")
	if err != nil {
		log.Fatal(Attention(fmt.Sprintf("%s", err)))
	}

	var b bytes.Buffer
	err = t.Execute(&b, a)
	if err != nil {
		log.Error(Attention("error executing template, err: %v", err))
		return
	}

	h, err := http.NewRequest(w.Method, w.URL, &b)
	if err != nil {
		return err
	}
	for _, header := range w.Headers {
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

func (f WebHookReceiver) Configured() bool {
	return !reflect.DeepEqual(f, WebHookReceiver{})
}

func (f WebHookReceiver) GetDefaultFields() (s []string) {
	for f := range FormatAsAPMessage(WebHookReceiver{}) {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
}
