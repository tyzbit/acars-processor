package receiver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/tyzbit/acars-processor/annotator"
	. "github.com/tyzbit/acars-processor/config"
	. "github.com/tyzbit/acars-processor/decorate"
)

type DiscordHandlerReciever struct {
	Payload any
}

type DiscordWebhookMessage struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordEmbed struct {
	Title       string           `json:"title,omitempty"`
	Description string           `json:"description,omitempty"`
	URL         string           `json:"url,omitempty"`
	Color       string           `json:"color,omitempty"`
	Thumbnail   DiscordThumbnail `json:"thumbnail,omitempty"`
	Fields      []DiscordField   `json:"fields"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordThumbnail struct {
	URL string `json:"url,omitempty"`
}

func (d DiscordHandlerReciever) Name() string {
	return "discord"
}

func (d DiscordHandlerReciever) SubmitACARSAnnotations(a annotator.Annotation) error {
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var content string
	requiredFieldsPresent := true
	for _, requiredField := range Config.Receivers.DiscordWebhook.RequiredFields {
		if !slices.Contains(keys, requiredField) && a[requiredField] != "" {
			requiredFieldsPresent = false
		}
	}
	if !requiredFieldsPresent {
		log.Info(Content("one or more required fields not present, not calling discord"))
		return nil
	}
	for _, key := range keys {
		r, _ := regexp.Compile(".*Text")
		textField := r.MatchString(key)
		v := a[key]
		if Config.Receivers.DiscordWebhook.FormatText &&
			v != "" && textField {
			v = fmt.Sprintf("```%s```", v)
		}
		content = fmt.Sprintf("%s\n**%s**: %v", content, key, v)
	}

	message := DiscordWebhookMessage{
		Content: "# ACARS Message\n" + content,
	}

	buff := new(bytes.Buffer)
	err := json.NewEncoder(buff).Encode(message)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", Config.Receivers.DiscordWebhook.URL, buff)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", WebhookUserAgent)
	// Hardcoded for now because most webhooks will be JSON
	req.Header.Add("Content-Type", "application/json")

	log.Info(Content("calling discord receiver"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if response := string(body); response != "" {
		log.Debug(Aside("discord api returned: %s", response))
	}
	return err
}
