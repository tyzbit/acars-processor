package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	api "github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
)

var OllamaSystemPrompt string = `You will evaluate a message to decide if it 
matches the provided criteria.
Your entire response MUST BE A VALID RAW JSON OBJECT with the format:

{"decision": DECISION, "reasoning": REASONING}

with the unquoted data in that object being results from your evaluation.

DO NOT USE BACKTICKS.
DECISION MUST BE EITHER true or false without quotes adhering to JSON standards.

REASONING MUST BE a valid quoted string adhering to JSON standards and
it should be a terse explanation of why, based on the criteria, 
you made the decision you made.

Here is the criteria:
%s

Here is the message I want you to evaluate:
%s
`

type OllamaResponse struct {
	Decision  bool   `json:"decision"`
	Reasoning string `json:"reasoning"`
}

// Return true if a message passes a filter, false otherwise
func OllamaFilter(m string) bool {
	if config.OllamaModel == "" {
		log.Warn("Ollama model not specified, this is required to use the ollama filter")
		return true
	}
	if match, err := regexp.Match(`\S*`, []byte(m)); !match || err != nil {
		log.Info("message was blank, filtering without calling Ollama")
		return true
	}
	url, err := url.Parse(config.OllamaURL)
	if err != nil {
		log.Fatalf("Ollama url could not be parsed: %s", err)
		return true
	}
	client := api.NewClient(url, &http.Client{})
	if err != nil {
		log.Fatalf("error initializing Ollama: %s", err)
	}

	if config.OllamaSystemPrompt != "" {
		OllamaSystemPrompt = config.OllamaSystemPrompt
	}
	messages := []api.Message{
		{
			Role:    "system",
			Content: OllamaSystemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf(OllamaSystemPrompt, config.OllamaPrompt, m),
		},
	}

	stream := false
	ctx := context.Background()
	req := &api.ChatRequest{
		Model:    config.OllamaModel,
		Messages: messages,
		Stream:   &stream,
	}

	var r OllamaResponse
	respFunc := func(resp api.ChatResponse) error {
		// Remove any extra formatting
		content := strings.ReplaceAll(resp.Message.Content, "```json", "")
		content = strings.ReplaceAll(content, "`", "")
		err = json.Unmarshal([]byte(content), &r)
		log.Debugf("ollama response: %s", resp.Message.Content)
		if err != nil {
			return err
		}
		return nil
	}

	log.Debugf("calling Ollama, model %s, prompt: %s", config.OllamaModel, config.OllamaPrompt)
	err = client.Chat(ctx, req, respFunc)
	if err != nil {
		log.Errorf("error using Ollama: %s", err)
		return true
	}
	return r.Decision
}
