package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"text/template"

	"github.com/mattn/go-mastodon"
	log "github.com/sirupsen/logrus"
)

type MastodonReceiver struct {
	Module
	Receiver
	// Full URL to the Mastodon server
	Server string `jsonschema:"required,example=https://mastodon.social,default=https://mastodon.social" default:"https://mastodon.social"`
	// Get this from your Mastodon server
	ClientID string `jsonschema:"required,default=Get this from your Mastodon server" default:"1f2a3c4e-9b83736f"`
	// Get this from your Mastodon server
	ClientSecret string `jsonschema:"required,default=Get this from your Mastodon server" default:"1f2a3c4e-9b83736f"`
	// Get this from your Mastodon server
	AccessToken string `jsonschema:"required,default=Get this from your Mastodon server" default:"1f2a3c4e-9b83736f"`
	// Visibility for posts. MUST BE ONE OF: public,unlisted,private,direct
	Visibility string `jsonschema:"required,example=public,example=unlisted,example=private,example=direct,default=unlisted" default:"unlisted"`
	// Go template for the post. Insert fields like this: `{{ index . "ACARSProcessor.TailCode" }}`
	PostGoTemplate string `jsonschema:"example=New message from aircraft! Message is {{ index . \"ACARSProcessor.MessageText\" }}" default:"New message from aircraft! Message is: {{ index . \"ACARSMessage.MessageText\" }}"`
}

func (mr MastodonReceiver) Name() string {
	return reflect.TypeOf(mr).Name()
}

func (f MastodonReceiver) Configured() bool {
	return !reflect.DeepEqual(f, MastodonReceiver{})
}

func (f MastodonReceiver) GetDefaultFields() (s []string) {
	for f := range FormatAsAPMessage(MastodonReceiver{}, "") {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
}

func (mr MastodonReceiver) Send(m APMessage) error {
	mstdn := mastodon.NewClient(&mastodon.Config{
		Server:       mr.Server,
		ClientID:     mr.ClientID,
		ClientSecret: mr.ClientSecret,
		AccessToken:  mr.AccessToken,
	})
	if mr.Server == "" {
		return fmt.Errorf("Mastodon server URL not specified")
	}
	var status string
	t := template.New(mr.Name())
	tp, err := t.Parse(mr.PostGoTemplate)
	if err != nil {
		return err
	}
	b := bytes.Buffer{}
	err = tp.Execute(&b, m)
	if err != nil {
		return err
	}
	br, err := io.ReadAll(&b)
	if err != nil {
		return err
	}
	status = string(br)
	toot := mastodon.Toot{
		Status:     status,
		Visibility: mr.Visibility,
	}
	log.Debug(Aside("%s: calling receiver", mr.Name()))
	post, err := mstdn.PostStatus(context.Background(), &toot)
	if err != nil {
		return fmt.Errorf("posting failed, err: %s", err)
	}
	postBytes, _ := json.Marshal(post)
	log.Debug(Aside("%s: status posted, status info: %s", mr.Name(), string(postBytes)))
	return nil
}
