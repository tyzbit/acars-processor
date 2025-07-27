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

const (
	defaultEmbedColor = "1150122" // cyan, base 10
	footer            = "-# Message generated with [acars-processor](<https://github.com/tyzbit/acars-processor>)"
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

// Color represents an RGB color.
type Color struct {
	R, G, B uint8
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
	content = content + footer
	if config.Receivers.DiscordWebhook.Embed {
		var url, transmitter, thumbnail, embedColorString string
		var embedColorValue int

		for _, key := range keys {
			v := fmt.Sprintf("%v", a[key])
			if key == "acarsAircraftTailCode" {
				transmitter = v
			}
			if key == "acarsExtraThumbnailLink" {
				thumbnail = v
			}
			if key == "acarsMessageFrom" {
				direction := " from "
				if v == "Tower" {
					direction = " to "
				}
				// If the tailcode field isn't present, do not append this to the title
				if transmitter != "" {
					transmitter = direction + transmitter
				}
			}
			if slices.Contains(config.Receivers.DiscordWebhook.EmbedColorFacetFields, key) {
				embedColorString = embedColorString + v
			}
			if key == config.Receivers.DiscordWebhook.EmbedColorGradientField {
				embedColorValue = a[key].(int)
			}
		}

		var color string
		if config.Receivers.DiscordWebhook.EmbedColorFacetFields != nil {
			color = GetColorForString(embedColorString)
		}
		if embedColorValue != 0 {
			color = fmt.Sprintf("%d", GetColorForInt(embedColorValue))
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

// Takes a string and returns a deterministic RGB value for it, expressed as
// an integer string.
// No collision avoidance or distribution considerations.
func GetColorForString(s string) (rgb string) {
	sha := sha256.New()
	sha.Write([]byte(s))
	hash := sha.Sum(nil)
	return fmt.Sprintf("%d", int(big.NewInt(0).SetBytes(hash[0:3]).Uint64()))
}

// interpolateLinearly interpolates between two colors.
func interpolateLinearly(c1, c2 Color, t float64) Color {
	return Color{
		R: uint8(float64(c1.R)*(1-t) + float64(c2.R)*t),
		G: uint8(float64(c1.G)*(1-t) + float64(c2.G)*t),
		B: uint8(float64(c1.B)*(1-t) + float64(c2.B)*t),
	}
}

// getColorInt returns a smoothly interpolated color (as int) for input in range 1–100.
func GetColorForInt(value int) int {
	if value < 1 {
		value = 1
	} else if value > 100 {
		value = 100
	}

	colors := []Color{
		{0x64, 0x8F, 0xFF}, // #648FFF
		{0x78, 0x5E, 0xF0}, // #785EF0
		{0xFF, 0xB0, 0x00}, // #FFB000
		{0xFE, 0x61, 0x00}, // #FE6100
		{0xDC, 0x26, 0x7F}, // #DC267F
	}

	// Total number of segments is 4 (between 5 colors)
	segmentSize := 100 / (len(colors) - 1)
	segment := (value - 1) / segmentSize
	t := float64((value-1)%segmentSize) / float64(segmentSize)

	c1 := colors[segment]
	c2 := colors[segment+1]
	interp := interpolateLinearly(c1, c2, t)

	// Return color as 0xRRGGBB integer
	return (int(interp.R) << 16) | (int(interp.G) << 8) | int(interp.B)
}
