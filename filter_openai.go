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

var OpenAIPromptTemplate string = `I have a message I want you to evaluate.
You MUST respond in valid JSON with the format:
{"decision": boolean, "reasoning": string} with the unquoted data in that object
being results from your evaluation.

"decision" should ONLY be true or false.
"reasoning" should ONLY be a string, and it should be a terse explanation of
why, based on the criteria, you made the decision you made.
Do not use backticks.

Here is the criteria:
%s

Here is the message I want you to evaluate:
%s
`

type OpenAIResponse struct {
	Decision  bool   `json:"decision"`
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
	log.Debugf("calling OpenAI, prompt: %s", config.OpenAIPrompt)
	chatCompletion, err := client.Chat.Completions.New(context.TODO(),
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(fmt.Sprintf(OpenAIPromptTemplate, config.OpenAIPrompt, m)),
			}),
			Model: openai.F(openAIModel),
		})
	if err != nil {
		log.Errorf("error using OpenAI: %s", err)
		return true
	}
	var r OpenAIResponse
	content := chatCompletion.Choices[0].Message.Content
	log.Debugf("response from OpenAI: %s", content)
	err = json.Unmarshal([]byte(content), &r)
	if err != nil {
		log.Warnf("error unmarshaling response from OpenAI: %s", err)
		return true
	}
	return r.Decision
}
