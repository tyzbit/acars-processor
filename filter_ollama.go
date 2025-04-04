package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	api "github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
)

var OllamaSystemPrompt string = `You will evaluate a message to decide if it 
matches the provided criteria.
You MUST respond in valid JSON with the format:

{"decision": boolean, "reasoning": string}

with the unquoted data in that object being results from your evaluation.

"decision" should ONLY be true or false.
"reasoning" should ONLY be a string, and it should be a terse explanation of
why, based on the criteria, you made the decision you made.
Do not use backticks.

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

	ctx := context.Background()
	req := &api.ChatRequest{
		Model:    config.OllamaModel,
		Messages: messages,
	}

	var r OllamaResponse
	respFunc := func(resp api.ChatResponse) error {
		err = json.Unmarshal([]byte(resp.Message.Content), &r)
		if err != nil {
			return err
		}
		return nil
	}

	err = client.Chat(ctx, req, respFunc)
	if err != nil {
		log.Errorf("error using Ollama: %s", err)
		return true
	}
	log.Debugf("Ollama response: %+v", r)
	return r.Decision
}
