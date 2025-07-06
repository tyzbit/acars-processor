package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

const defaultEmbedColor = "1150122" // cyan, base 10

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
	if config.Receivers.DiscordWebhook.URL == "" {
		log.Panic(Attention("Discord webhook URL not specified"))
	}
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	requiredFieldsPresent := true
	for _, requiredField := range config.Receivers.DiscordWebhook.RequiredFields {
		if !slices.Contains(keys, requiredField) && a[requiredField] != "" {
			requiredFieldsPresent = false
		}
	}
	if !requiredFieldsPresent {
		log.Info(Content("one or more required fields not present, not calling discord"))
		return nil
	}
	buff := new(bytes.Buffer)
	r, _ := regexp.Compile(".*Text")
	l, _ := regexp.Compile(".*Link")
	var embeds []DiscordEmbed
	var title, content string
	for _, key := range keys {
		textField := r.MatchString(key)
		linkField := l.MatchString(key)
		acarsTimeField := key == "acarsTimestamp"
		vdlm2TimeField := key == "vdlm2Timestamp"
		v := a[key]
		if config.Receivers.DiscordWebhook.FormatText &&
			v != "" && textField {
			v = fmt.Sprintf("```%s```", v)
		}
		if linkField && v != "" {
			linkType := strings.TrimPrefix(key, "acarsExtra")
			linkType = strings.TrimSuffix(linkType, "Link")
			v = fmt.Sprintf("[%s](%s)", linkType, v)
		}
		if acarsTimeField {
			v = int(v.(float64))
		}
		if config.Receivers.DiscordWebhook.FormatTimestamps &&
			(vdlm2TimeField || acarsTimeField) {
			v = fmt.Sprintf("<t:%d:R>", v)
		}
		content = fmt.Sprintf("%s**%s**: %v\n", content, key, v)
	}
	if config.Receivers.DiscordWebhook.Embed {
		var tailcode, transmitter, url, thumbnail, embedColorString string

		for _, key := range keys {
			v := fmt.Sprintf("%v", a[key])
			if key == "acarsAircraftTailCode" {
				tailcode = "(" + v + ")"
			}
			if key == "acarsExtraThumbnailLink" {
				thumbnail = v
			}
			if key == "acarsMessageFrom" {
				if v == "Tower" {
					if slices.Contains(keys, "acarsFlightNumber") {
						tailcode = "(to flight " + a["acarsFlightNumber"].(string) + ")"
					}
				}
			}
			if slices.Contains(config.Receivers.DiscordWebhook.EmbedColorFacetFields, key) {
				embedColorString = embedColorString + v
			}
		}

		embeds = append(embeds, DiscordEmbed{
			Title:       fmt.Sprintf("ACARS %s Message %s", transmitter, tailcode),
			Description: content,
			Color:       GetRGBForString(embedColorString),
			URL:         url,
			Thumbnail: DiscordThumbnail{
				URL: thumbnail,
			},
		})
		content = ""

	}
	// "Plain" message
	if !config.Receivers.DiscordWebhook.Embed {
		title = "# ACARS Message\n"
	}
	message := DiscordWebhookMessage{
		Content: title + content,
		// This will be an empty slice if embeds are not enabled
		Embeds: embeds,
	}

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

func GetRGBForString(s string) (rgb string) {
	sha := sha256.New()
	sha.Write([]byte(s))
	hash := sha.Sum(nil)
	return fmt.Sprintf("%d", int(big.NewInt(0).SetBytes(hash[0:3]).Uint64()))
}
