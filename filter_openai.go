package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	log "github.com/sirupsen/logrus"
)

var OpenAIPromptTemplate string = `You are an API that only responds in
valid, compact JSON objects without newlines. You will be provided with 
a message and criteria to evaluate it against.

You will respond with a JSON object without newlines that has:

- a "decision" key which is a boolean of true (if the message
matches the criteria) or false (if the message does not match the criteria) 
- a "reasoning" key which must be a simple explanation of the reasoning
behind your decision. You will always provide your reasoning for your decision
in the response under the "reasoning" key.

Here is the criteria:
%s
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
	if config.OpenAICustomPreamble != "" {
		OpenAIPromptTemplate = config.OpenAICustomPreamble
	}
	openAIModel := openai.ChatModelGPT4o
	if config.OpenAIModel != "" {
		openAIModel = config.OpenAIModel
	}
	log.Debugf("calling OpenAI model %s, prompt: %s", openAIModel, config.OpenAIPrompt)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(),
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(fmt.Sprintf(OpenAIPromptTemplate, config.OpenAIPrompt)),
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

	log.Debugf("OpenAI parsed json response: %s", content)
	if err != nil {
		log.Warnf("error unmarshaling response from OpenAI: %s", err)
		log.Debugf("OpenAI full response: %s", chatCompletion.Choices[0].Message.Content)
		return true
	}
	return r.Decision == "true" || r.Decision == true
}
