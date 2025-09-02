package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	STDINAPMessageQueue = make(chan APMessageQeueueItem, 10000)
)

// Connects to ACARS and starts listening to messages
func SubscribeToStandardIn() {
	go ReadSTDINAMessages()
	go HandleAPMessageQueue(STDINAPMessageQueue)
	log.Debug(Aside("launched stdin subscriber"))
}

func ReadSTDINAMessages() {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			var anext ACARSMessage
			var vnext VDLM2Message
			err := json.Unmarshal([]byte(line), &anext)
			if err != nil {
				log.Fatal(Attention("error unmarshalling: %s", err))
			}
			err = json.Unmarshal([]byte(line), &vnext)
			if err != nil {
				log.Fatal(Attention("error unmarshalling: %s", err))
			}
			if (vnext != VDLM2Message{}) {
				log.Info(Content("new vdlm2 message received ending in \""),
					Note(Last20Characters(vnext.VDL2.AVLC.ACARS.MessageText)),
					Content("\""))
				queueLength := len(STDINAPMessageQueue)
				nextap := vnext.Prepare()
				if msgJson, err := json.Marshal(vnext); err == nil {
					log.Debug(Emphasised("new vdlm2 message content "),
						Note("(%d already in queue)", queueLength),
						Content(": "),
						Aside("%s", strings.ReplaceAll(string(msgJson), "\n", "\t")))
				}
				db.Create(&vnext)
				STDINAPMessageQueue <- APMessageQeueueItem{
					VDLM2Message: vnext,
					APMessage:    nextap,
				}
				continue
			}
			if (anext != ACARSMessage{}) {
				log.Info(Content("new acars message received ending in \""),
					Note(Last20Characters(anext.MessageText)),
					Content("\""))
				queueLength := len(STDINAPMessageQueue)
				nextap := anext.Prepare()
				if msgJson, err := json.Marshal(anext); err == nil {
					log.Debug(Emphasised("new acars message content "),
						Note("(%d already in queue)", queueLength),
						Content(": "),
						Aside("%s", strings.ReplaceAll(string(msgJson), "\n", "\t")))
				}
				db.Create(&anext)
				STDINAPMessageQueue <- APMessageQeueueItem{
					ACARSMessage: anext,
					APMessage:    nextap,
				}
				continue
			}
			log.Warn("error reading message as acars or vdlm2")
		}

		// Check for any errors that occurred during scanning
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading from stdin:", err)
		}

		log.Warn(Attention("stdin handler exited, restarting"))
	}
}
