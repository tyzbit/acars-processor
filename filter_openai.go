package main

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	log "github.com/sirupsen/logrus"
)

var (
	OpenAISystemPrompt = `You are an AI that is an expert at logical 
	reasoning. You will be provided criteria and then a communication message. 
	You will use your skills and any examples provided to evaluate determine 
	if the message positively matches the provided criteria.

	Here's the criteria:
	`
	OpenAIFinalInstructions = `If the message affirmatively matches the criteria, 
	return 'true' in the 'message_matches' field.

	If the message definitely does not match the criteria, return 'false' in the
	'message_matches' field. 
	
	Briefly provide your reasoning regarding your 
	decision in the 'reasoning' field, pointing out specific evidence that 
	factored in your decision on whether the message matches the criteria.`
)

type OpenAIResponse struct {
	MessageMatches any    `json:"message_matches"`
	Reasoning      string `json:"reasoning"`
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
				openai.SystemMessage(OpenAIFinalInstructions),
				openai.UserMessage(m),
			}),
			Model: openai.F(openAIModel),
			// Make it deterministic
			Temperature: openai.Float(0),
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
	content = SanitizeJSONString(content)
	err = json.Unmarshal([]byte(content), &r)

	if err != nil {
		log.Warnf("error unmarshaling response from OpenAI: %s", err)
		log.Debugf("OpenAI full response: %s", chatCompletion.Choices[0].Message.Content)
		return true
	}
	decision := r.MessageMatches == "true" || r.MessageMatches == true
	action := map[bool]string{
		true:  "allow",
		false: "filter",
	}
	log.Infof("openai decision: %s, message ending in: %s, reasoning: %s", action[decision], Last20Characters(m), r.Reasoning)
	return decision
}
