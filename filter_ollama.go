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

var OllamaSystemPrompt string = `You are an API that only responds in
valid, compact JSON objects without newlines. You will be provided with 
a message and criteria to evaluate it against.

You will respond with a JSON object without newlines that has:

- a "decision" key which is a boolean of true (if the message
matches the criteria) or false (if the message does not match the criteria) 
- a "reasoning" key which must be a simple explanation of the reasoning
behind your decision. You will always provide your reasoning for your decision
in the response under the "reasoning" key.

Here is the message you will evaluate:
%s

Here is the criteria:
%s
`

type OllamaResponse struct {
	Decision  any    `json:"decision"`
	Reasoning string `json:"reasoning"`
}

// Return true if a message passes a filter, false otherwise
func OllamaFilter(m string) bool {
	// If message is blank, return
	if regexp.MustCompile(`^\s*$`).MatchString(m) {
		log.Debug("message was blank, filtering without calling ollama")
		return false
	}
	if config.OllamaModel == "" || config.OllamaPrompt == "" {
		log.Warn("Ollama model and prompt are required to use the ollama filter")
		return true
	}
	url, err := url.Parse(config.OllamaURL)
	if err != nil {
		log.Errorf("Ollama url could not be parsed: %s", err)
		return true
	}
	client := api.NewClient(url, &http.Client{})
	if err != nil {
		log.Errorf("error initializing Ollama: %s", err)
		return true
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
			Content: fmt.Sprintf(OllamaSystemPrompt, m, config.OllamaPrompt),
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
		// Parse the JSON payload (hopefully)
		rex := regexp.MustCompile(`\{[^{}]+\}`)
		matches := rex.FindAllStringIndex(resp.Message.Content, -1)

		// Find the last json payload in case the model reasons about
		// one in the middle of thinking
		if len(matches) == 0 {
			err = fmt.Errorf("did not find a json object in response: %s",
				resp.Message.Content)
			return err
		}
		start, end := matches[len(matches)-1][0], matches[len(matches)-1][1]
		content := resp.Message.Content[start:end]
		err = json.Unmarshal([]byte(content), &r)

		log.Debugf("ollama parsed json response: %s", content)
		if err != nil {
			err = fmt.Errorf("%w, ollama full response: %s", err, resp.Message.Content)
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
	return r.Decision == "true" || r.Decision == true
}
