package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	api "github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
)

var OLLamaSystemPrompt string = `You will evaluate a message to decide if it 
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
	if config.OLLamaModel == "" {
		log.Warn("ollama model not specified, this is required to use the ollama filter")
		return true
	}
	if match, err := regexp.Match(`\S*`, []byte(m)); !match || err != nil {
		log.Info("message was blank, filtering without calling Ollama")
		return true
	}
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	if config.OLLamaSystemPrompt != "" {
		OLLamaSystemPrompt = config.OLLamaSystemPrompt
	}
	messages := []api.Message{
		{
			Role:    "system",
			Content: OLLamaSystemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf(OLLamaSystemPrompt, config.OLLamaPrompt, m),
		},
	}

	ctx := context.Background()
	req := &api.ChatRequest{
		Model:    config.OLLamaModel,
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
