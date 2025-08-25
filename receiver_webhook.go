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
	Headers []WebHookReceiverHeaders
	// Go template for the post. Use dot notation with double curly braces to insert fields (`{{ .ACARSProcessor.MessageText }}`)
	PayloadGoTemplate string `jsonschema:"required,example={\"tail_code\": \"{{ index . \"ACARSProcessor.TailCode\" }}\"}" default:"{\n \"tail_code\": \"{{ index . \"\\\"ACARSProcessor.TailCode\\\"\" }}\"\n}"`
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

func (f WebHookReceiver) Configured() bool {
	return !reflect.DeepEqual(f, WebHookReceiver{})
}

func (f WebHookReceiver) GetDefaultFields() (s []string) {
	for f := range FormatAsAPMessage(WebHookReceiver{}, f.Name()) {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
}

func (w WebHookReceiver) Send(a APMessage) (err error) {
	if w.URL == "" {
		return fmt.Errorf("Webhook URL not specified")
	}
	t := template.New(w.Name())
	if err != nil {
		log.Fatal(Attention(fmt.Sprintf("%s", err)))
	}
	t, err = t.Parse(w.PayloadGoTemplate)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	err = t.Execute(&b, a)
	if err != nil {
		return fmt.Errorf("error executing template, err: %v", err)
	}

	h, err := http.NewRequest(w.Method, w.URL, &b)
	if err != nil {
		return fmt.Errorf("error preparing new webhook request: %w", err)
	}
	for _, header := range w.Headers {
		h.Header.Add(header.Name, header.Value)
	}

	c := http.Client{}
	log.Debug(Content("%s: calling receiver", w.Name()))
	resp, err := c.Do(h)
	if err != nil {
		return fmt.Errorf("error making webhook request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading webhook response: %w", err)
	}

	log.Debug(Aside("%s: webhook returned: %s", w.Name(), string(body)))
	return nil
}
