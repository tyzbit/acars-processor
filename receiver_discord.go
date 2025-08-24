package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"

	hue "codeberg.org/tyzbit/huenique"
	log "github.com/sirupsen/logrus"
)

const (
	defaultEmbedColor = "1150122" // cyan, base 10
	footer            = "-# Message generated with [acars-processor](<https://github.com/tyzbit/acars-processor>)"
)

type DiscordReceiver struct {
	Module
	Receiver
	// Full URL to the Discord webhook for a channel (edit a channel in the Discord UI for the option to create a webhook).
	URL string `jsonschema:"required" default:"https://discord.com/api/webhooks/1234321/unique_id1234"`
	// Should an embed be sent instead of a simpler message?
	Embed bool `jsonschema:"default=true" default:"true"`
	// Pick one or more fields that deterministically determines the embed color
	EmbedColorFacetFields []string `default:"[acarsAircraftTailCode]"`
	// Pick one or more fields that determines the embed color according to this field, which should be an integer between 1 and 100
	EmbedColorGradientField string `default:"ollamaProcessedNumber"`
	// An array of colors that corresponds with EmbedColorGradientField values
	EmbedColorGradientSteps []hue.Color `default:"{{R:0,G:100,B:0}}"`
	// Surround fields with message content with backticks so they are monospaced and stand out.
	FormatText bool `jsonschema:"default=true" default:"true"`
	// Add Discord-specific formatting to show human-readable instants from timestamps
	FormatTimestamps bool `jsonschema:"default=true" default:"true"`
	// Go template for the message. Insert fields like this: `{{ index . "ACARSProcessor.TailCode" }}`
	MessageGoTemplate string `jsonschema:"example=New message from aircraft! Message is {{ index . \"ACARSProcessor.MessageText\" }}" default:"New message from aircraft! Message is: {{ index . \"ACARSMessage.MessageText\" }}"`
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

func (d DiscordReceiver) Name() string {
	return reflect.TypeOf(d).Name()
}

func (f DiscordReceiver) Configured() bool {
	return !reflect.DeepEqual(f, DiscordReceiver{})
}

func (f DiscordReceiver) GetDefaultFields() (s []string) {
	for f := range FormatAsAPMessage(DiscordReceiver{}, "") {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
}

func (d DiscordReceiver) Send(m APMessage) error {
	if d.URL == "" {
		return fmt.Errorf("Discord webhook URL not specified")
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	buff := new(bytes.Buffer)
	r, _ := regexp.Compile(".*[Tt]ext")
	l, _ := regexp.Compile(".*[Ll]ink")
	var embeds []DiscordEmbed
	var title, content string
	if d.MessageGoTemplate != "" {
		t := template.New(d.Name())
		tp, err := t.Parse(d.MessageGoTemplate)
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
		content = string(br)
	} else {
		for _, key := range keys {
			textField := r.MatchString(key)
			linkField := l.MatchString(key)
			v := m[key]
			if d.FormatText &&
				v != "" && textField {
				v = fmt.Sprintf("```%s```", v)
			}
			if linkField && v != "" {
				///////////////////////// kawaii ^.*
				re := regexp.MustCompile(`^.*[._]`)
				linkType := re.ReplaceAllString(key, "")
				linkType = strings.TrimSuffix(linkType, "Link")
				v = fmt.Sprintf("[%s](%s)", linkType, v)
			}
			if ts, _ := regexp.Compile(".*[Tt]imestamp"); ts.Match([]byte(key)) {
				v = v.(int64)
				if d.FormatTimestamps {
					v = fmt.Sprintf("<t:%d:R>", v)
				}
			}
			content = fmt.Sprintf("%s**%s**: %v\n", content, key, v)
		}
		content = content + footer
	}
	if d.Embed {
		var url, transmitter, thumbnail, embedColorString string
		var embedColorValue int

		for _, key := range keys {
			v := fmt.Sprintf("%v", m[key])
			if ts, _ := regexp.Compile(".*[Tt]ail[Cc]ode"); ts.Match([]byte(key)) {
				transmitter = " from " + AircraftOrTower(v)
			}
			if ts, _ := regexp.Compile(".*[Tt]humbnail[Ll]ink"); ts.Match([]byte(key)) {
				thumbnail = v
			}
			if ts, _ := regexp.Compile(".*[Ff]rom"); ts.Match([]byte(key)) {
				transmitter = " from " + v
			}
			if slices.Contains(d.EmbedColorFacetFields, key) {
				embedColorString = embedColorString + v
			}
			if key == d.EmbedColorGradientField {
				var value int
				var err error
				// The value might be string because LLMs are fuzzybad, cast it
				// to an int
				if _, ok := m[key].(int); !ok {
					value, err = strconv.Atoi(m[key].(string))
					if err != nil {
						return fmt.Errorf("%w, field needs to be a number", err)
					}
				} else {
					value = m[key].(int)
				}
				embedColorValue = value
			}
		}

		var color string
		if d.EmbedColorFacetFields != nil {
			color = fmt.Sprintf("%d", hue.GetRGBValueForString(embedColorString))
		}
		if embedColorValue != 0 {
			color = fmt.Sprintf("%d", hue.PickRGBValueForInt(embedColorValue))
		}
		embeds = append(embeds, DiscordEmbed{
			Title:       fmt.Sprintf("ACARS Message%s", transmitter),
			Description: content,
			Color:       color,
			URL:         url,
			Thumbnail: DiscordThumbnail{
				URL: thumbnail,
			},
		})
		content = ""

	}
	// "Plain" message
	if !d.Embed {
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
	req, err := http.NewRequest("POST", d.URL, buff)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", WebhookUserAgent)
	// Hardcoded for now because most webhooks will be JSON
	req.Header.Add("Content-Type", "application/json")

	log.Debug(Content("calling discord receiver"))
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
		log.Debug(Aside("%s: api returned %s", d.Name(), response))
	}
	return err
}
