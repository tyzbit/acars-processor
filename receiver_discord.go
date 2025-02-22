package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

const MessageDescription = `# [%s](<%s>)
**Photo**: %s
**Flight Number**: %s
**Signal**: %s dBm
**Message Text**:` +
	"```%s```" +
	`**Further annotations**:
%s
`

func (d DiscordHandlerReciever) Name() string {
	return "Discord Handler"
}

func (d DiscordHandlerReciever) SubmitACARSAnnotations(a Annotation) error {
	// var fields []DiscordField
	// for key, value := range m {
	// 	fields = append(fields, DiscordField{
	// 		Name:   key,
	// 		Value:  fmt.Sprintf("%s", value),
	// 		Inline: false,
	// 	})
	// }

	// if len(fields) > 25 {
	// 	log.Warn("too many fields for Discord, truncating")
	// 	fields = fields[:25]
	// }

	// embed := DiscordWebhookMessage{
	// 	Embeds: []DiscordEmbed{{
	// 		Title:  "ACARS",
	// 		Fields: fields,
	// 	}},
	// }

	// Create a slice to hold the keys
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}

	// Sort the keys
	sort.Strings(keys)

	var content string
	for _, key := range keys {
		content = fmt.Sprintf("%s\n**%s**: %v", content, key, a[key])
	}

	message := DiscordWebhookMessage{
		Content: content,
	}

	buff := new(bytes.Buffer)
	json.NewEncoder(buff).Encode(message)
	req, err := http.NewRequest("POST", config.DiscordWebhookURL, buff)
	req.Header.Add("User-Agent", WebhookUserAgent)
	// Hardcoded for now because most webhooks will be JSON
	req.Header.Add("Content-Type", "application/json")

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

	// embed := discord.NewEmbed("ACARS Message", "Description", FlightAwareRoot+m.AircraftTailCode)
	// embed.Content = m.MessageText

	// distance := strconv.FormatFloat(m.Annotations[0].Annotation["adsbAircraftDistanceMi"].(float64), 'f', 2, 64) +
	// 	" mi"
	// embed.SetAuthor("ACARS Message", "", "")
	// //Creates a new field and adds it to the embed
	// //boolean represents whether the field is inline or not
	// embed.AddField("Flight Number", m.FlightNumber, false)
	// embed.AddField("Signal Strength", fmt.Sprintf("%f", m.SignaldBm), true)
	// embed.AddField("Distance", distance, true)

	// //Sets image of embed
	// embed.AddField("Aircraft photo", JetPhotosRoot+m.AircraftTailCode, true)

	// //Sets color of embed given hexcode
	// // embed.SetColor("#F1B379")

	// log.Debugf("payload to Discord webhook: %+v", embed)
	// //Send embed to given webhook
	// err := embed.SendToWebhook(config.DiscordWebhookURL)
	return err
}
