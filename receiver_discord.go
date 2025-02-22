package main

import (
	"fmt"

	"github.com/aiomonitors/godiscord"
)

type DiscordHandlerReciever struct {
	Payload any
}

func (d DiscordHandlerReciever) Name() string {
	return "Discord Handler"
}

func (d DiscordHandlerReciever) SubmitACARSMessage(m AnnotatedACARSMessage) error {
	//Create a new embed object
	embed := godiscord.NewEmbed("ACARS Message", m.MessageText, fmt.Sprintf("%s/%s", FlightAwareRoot, m.AircraftTailCode))

	//Creates a new field and adds it to the embed
	//boolean represents whether the field is inline or not
	embed.AddField("Flight Number", m.FlightNumber, true)
	embed.AddField("Signal Strength", fmt.Sprintf("%f", m.SignaldBm), true)
	embed.AddField("Distance", fmt.Sprintf("%d", m.Annotations[0].Annotation["adsbAircraftDistanceMi"]), true)
	embed.AddField("Flight Number", m.FlightNumber, true)
	embed.AddField("Flight Number", m.FlightNumber, true)

	// picture := d.GetURLToAircraftThumbnail(m.AircraftTailCode)
	// if picture != "" {
	// 	//Sets the thumbail of the embed
	// 	embed.SetThumbnail(picture)
	// }

	// //Sets image of embed
	// embed.SetImage(picture)

	//Sets color of embed given hexcode
	embed.SetColor("#F1B379")

	//Send embed to given webhook
	err := embed.SendToWebhook(config.DiscordWebhookURL)
	return err
}

// // <meta property="og:image" content="https://t.plnspttrs.net/16865/1692533_c484a0d563_280.jpg" >
// func (d DiscordHandlerReciever) GetURLToAircraftThumbnail(t string) (url string) {
// 	c := http.Client{}
// 	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", PlaneSpottersPhotoRoot, t), nil)
// 	if err != nil {
// 		log.Errorf("unable to create new http request: %v", err)
// 		return url
// 	}

// 	resp, err := c.Do(req)
// 	if err != nil {
// 		return url
// 	}
// 	defer resp.Body.Close()

// 	// Parse the HTML document using goquery
// 	doc, err := goquery.NewDocumentFromReader(resp.Body)
// 	if err != nil {
// 		log.Errorf("could not parse html doc: %v", err)
// 		return url
// 	}

// 	// Find the meta tag with property "og:image" and extract its content attribute
// 	metaTag := doc.Find(`meta[property="og:image"]`)
// 	if metaTag.Length() == 0 {
// 		log.Error("og:image meta tag not found")
// 	}

// 	url, exists := metaTag.Attr("content")
// 	if !exists {
// 		log.Error("content attribute not found in og:image meta tag")
// 	}

// 	return url
// }
