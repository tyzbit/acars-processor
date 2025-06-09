package main

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/fatih/color"
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
		log.Info(yo().FYI("message was blank, filtering without calling OpenAI").FRFR())
		return false
	}
	client := openai.NewClient(
		option.WithAPIKey(config.Filters.OpenAI.APIKey),
	)
	if config.Filters.OpenAI.SystemPrompt != "" {
		OpenAISystemPrompt = config.Filters.OpenAI.SystemPrompt
	}
	openAIModel := openai.ChatModelGPT4o
	if config.Filters.OpenAI.Model != "" {
		openAIModel = config.Filters.OpenAI.Model
	}

	log.Debug(yo().INFODUMP("calling OpenAI model ").Hmm(openAIModel).FRFR())
	chatCompletion, err := client.Chat.Completions.New(context.TODO(),
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(OpenAISystemPrompt),
				openai.SystemMessage(config.Filters.OpenAI.UserPrompt),
				openai.SystemMessage(OpenAIFinalInstructions),
				openai.UserMessage(m),
			}),
			Model: openai.F(openAIModel),
			// Make it deterministic
			Temperature: openai.Float(0),
		})
	if err != nil {
		log.Error(yo().Uhh("error using OpenAI: %s", err).FRFR())
		return true
	}

	var r OpenAIResponse
	// Parse the JSON payload (hopefully)
	rex := regexp.MustCompile(`\{[^{}]+\}`)
	matches := rex.FindAllStringIndex(chatCompletion.Choices[0].Message.Content, -1)

	// Find the last json payload in case the model reasons about
	// one in the middle of thinking
	if len(matches) == 0 {
		log.Error(yo().Uhh("did not find a json object in response: %s",
			chatCompletion.Choices[0].Message.Content).FRFR())
		return true
	}
	start, end := matches[len(matches)-1][0], matches[len(matches)-1][1]
	content := chatCompletion.Choices[0].Message.Content[start:end]
	content = SanitizeJSONString(content)
	err = json.Unmarshal([]byte(content), &r)

	if err != nil {
		log.Warn(yo().Uhh("error unmarshaling response from OpenAI: %s", err).FRFR())
		log.Debug(yo().INFODUMP("OpenAI full response: %s", chatCompletion.Choices[0].Message.Content))
		return true
	}
	decision := r.MessageMatches == "true" || r.MessageMatches == true
	action := map[bool]DM{
		true: {
			Color:   *color.New(color.FgCyan),
			Message: "allow",
		},
		false: {
			Color:   *color.New(color.FgYellow),
			Message: "filter",
		},
	}
	log.Info(
		yo().FYI("openai decision: ").
			GlowUp(action[decision]).
			FYI("message ending in: ").
			INFODUMP(Last20Characters(m)).
			FYI(", reasoning:").
			INFODUMP(r.Reasoning).FRFR())
	return decision
}
