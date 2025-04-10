package main

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	log "github.com/sirupsen/logrus"
)

var OpenAISystemPrompt string = `You are an API that only responds in
valid JSON objects. You will carefully evaluate a message to determine if it 
matches specific criteria.

If the message matches the criteria, "decision" will ALWAYS be true: 
{"decision": true, "reasoning": "REASON"}

If the message does not match the criteria, "decision" will ALWAYS be false:
{"decision": false, "reasoning": "REASON"}

Replace REASON with a summary of why "decision" was true or false.
`

type OpenAIResponse struct {
	Decision  any    `json:"decision"`
	Reasoning string `json:"reasoning"`
}

// Return true if a message passes a filter, false otherwise
func OpenAIFilter(m string) bool {
	// If message is blank, return
	if regexp.MustCompile(`^\s*$`).MatchString(m) {
		log.Info("message was blank, filtering without calling OpenAI")
		return false
	}
	client := openai.NewClient(
		option.WithAPIKey(config.OpenAIAPIKey),
	)
	if config.OpenAISystemPrompt != "" {
		OpenAISystemPrompt = config.OpenAISystemPrompt
	}
	openAIModel := openai.ChatModelGPT4o
	if config.OpenAIModel != "" {
		openAIModel = config.OpenAIModel
	}
	log.Debugf("calling OpenAI model %s", openAIModel)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(),
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(OpenAISystemPrompt),
				openai.SystemMessage(config.OpenAIUserPrompt),
				openai.UserMessage(m),
			}),
			Model: openai.F(openAIModel),
		})
	if err != nil {
		log.Errorf("error using OpenAI: %s", err)
		return true
	}

	var r OpenAIResponse
	// Parse the JSON payload (hopefully)
	rex := regexp.MustCompile(`\{[^{}]+\}`)
	matches := rex.FindAllStringIndex(chatCompletion.Choices[0].Message.Content, -1)

	// Find the last json payload in case the model reasons about
	// one in the middle of thinking
	if len(matches) == 0 {
		log.Errorf("did not find a json object in response: %s",
			chatCompletion.Choices[0].Message.Content)
		return true
	}
	start, end := matches[len(matches)-1][0], matches[len(matches)-1][1]
	content := chatCompletion.Choices[0].Message.Content[start:end]
	err = json.Unmarshal([]byte(content), &r)

	if err != nil {
		log.Warnf("error unmarshaling response from OpenAI: %s", err)
		log.Debugf("OpenAI full response: %s", chatCompletion.Choices[0].Message.Content)
		return true
	}
	decision := r.Decision == "true" || r.Decision == true
	action := map[bool]string{
		true:  "allow",
		false: "filter",
	}
	log.Infof("ollama decision: %s, reasoning: %s", action[decision], r.Reasoning)
	return decision
}
