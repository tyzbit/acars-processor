package receiver

import (
	"bytes"
	"io"
	"net/http"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/tyzbit/acars-processor/annotator"
	. "github.com/tyzbit/acars-processor/config"
	. "github.com/tyzbit/acars-processor/decorate"
)

type WebhookHandlerReciever struct {
	Payload any
}

// Must satisfy Receiver interface
func (n WebhookHandlerReciever) Name() string {
	return "webhook"
}

// Must satisfy Receiver interface
func (n WebhookHandlerReciever) SubmitACARSAnnotations(a annotator.Annotation) (err error) {
	t, err := template.ParseFiles("webhook.tpl")
	if err != nil {
		log.Fatal(Attention(err))
	}

	var b bytes.Buffer
	err = t.Execute(&b, a)
	if err != nil {
		log.Error(Attention("error executing template, err: %v", err))
		return
	}

	h, err := http.NewRequest(Config.Receivers.Webhook.Method, Config.Receivers.Webhook.URL, &b)
	if err != nil {
		return err
	}
	for _, header := range Config.Receivers.Webhook.Headers {
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
