package main

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

func (d DiscordHandlerReciever) SubmitACARSAnnotations(a Annotation) error {
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var content string
	requiredFieldsPresent := true
	for _, requiredField := range config.Receivers.DiscordWebhook.RequiredFields {
		if !slices.Contains(keys, requiredField) && a[requiredField] != "" {
			requiredFieldsPresent = false
		}
	}
	if !requiredFieldsPresent {
		log.Info(yo.FYI("one or more required fields not present, not calling discord").FRFR())
		return nil
	}
	for _, key := range keys {
		r, _ := regexp.Compile(".*Text")
		textField := r.MatchString(key)
		v := a[key]
		if config.Receivers.DiscordWebhook.FormatText &&
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
	req, err := http.NewRequest("POST", config.Receivers.DiscordWebhook.URL, buff)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", WebhookUserAgent)
	// Hardcoded for now because most webhooks will be JSON
	req.Header.Add("Content-Type", "application/json")

	log.Info(yo.FYI("calling discord receiver").FRFR())
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
		log.Debug(yo.INFODUMP("discord api returned: %s", response).FRFR())
	}
	return err
}
