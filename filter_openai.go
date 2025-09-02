// TODO: Feature parity with Ollama
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	log "github.com/sirupsen/logrus"
)

var (
	OpenAISystemPrompt = `You are an AI that is an expert at 
	reasoning about text content. You will be provided criteria and then a 
	communication message. You will use your reasoning and any examples provided
	to determine if the message positively matches the provided criteria.

	Here's the criteria:
	`
	OpenAIFinalInstructions = `
	If the message definitely matches the criteria, 
	return 'true' in the 'message_matches_criteria' field.

	If the message definitely does not match the criteria, 
	return 'false' in the 'message_matches_criteria' field. 
	
	Provide a very short, high-level explanation with 
	the reasoning for your decision in the "reasoning" field.`
)

type OpenAIFilterer struct {
	Filterer
	// Whether to filter messages where the OpenAI filter itself fails. Recommended if your ollama instance sometimes returns errors.
	FilterOnFailure bool `default:"true"`
	// Inverse logic (for example, Invert: true, HasText: true means messages with text are FILTERED)
	Invert bool   `json:",omitempty" default:"false"`
	APIKey string `jsonschema:"required" default:"example_key"`
	// Model to use.
	Model string `jsonschema:"required,default=gpt-4o" default:"gpt-4o"`
	// Instructions for OpenAI model to use when filtering messages. More detail is better.
	UserPrompt string `jsonschema:"required,example=Does this message talk about coffee makers or lavatories (shortand LAV is sometimes used)?" default:"Does this message talk about coffee makers or lavatories (shorthand LAV is sometimes used)?"`
	// Override the built-in system prompt to instruct the model on how to behave for requests (not usually necessary).
	SystemPrompt string `default:"Answer like a pirate"`
	// How long to wait until giving up on any request to OpenAI.
	Timeout int `default:"5"`
}

func (o OpenAIFilterer) Name() string {
	return reflect.TypeOf(o).Name()
}

func (f OpenAIFilterer) Configured() bool {
	return !reflect.DeepEqual(f, OpenAIFilterer{})
}

type OpenAIResponse struct {
	MessageMatches any    `json:"message_matches_criteria"`
	Reasoning      string `json:"reasoning"`
}

// Return true if a message passes a filter, false otherwise
func (o OpenAIFilterer) Filter(m APMessage) (filterThisMessage bool, reason string, err error) {
	ms := GetAPMessageCommonFieldAsString(m, "MessageText")

	// If message is blank, return
	if regexp.MustCompile(emptyStringRegex).MatchString(ms) {
		return true, "message blank", nil
	}
	client := openai.NewClient(
		option.WithAPIKey(o.APIKey),
	)
	if o.SystemPrompt != "" {
		OpenAISystemPrompt = o.SystemPrompt
	}
	openAIModel := openai.ChatModelGPT4o
	if o.Model != "" {
		openAIModel = o.Model
	}

	log.Debug(Aside("%s considering message ending in \"", o.Name()),
		Note(Last20Characters(ms)),
		Aside("\", model "),
		Note(openAIModel))

	chatCompletion, err := client.Chat.Completions.New(context.TODO(),
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(OpenAISystemPrompt),
				openai.SystemMessage(o.UserPrompt),
				openai.SystemMessage(OpenAIFinalInstructions),
				openai.UserMessage(ms),
			}),
			Model: openai.F(openAIModel),
			// Make it deterministic
			Temperature: openai.Float(0),
		})
	if err != nil {
		return o.FilterOnFailure, "", err
	}

	var r OpenAIResponse
	// Parse the JSON payload (hopefully)
	rex := regexp.MustCompile(`\{[^{}]+\}`)
	matches := rex.FindAllStringIndex(chatCompletion.Choices[0].Message.Content, -1)

	// Find the last json payload in case the model reasons about
	// one in the middle of thinking
	if len(matches) == 0 {
		return o.FilterOnFailure, "",
			fmt.Errorf("did not find a json object in response, message: %s, content: %s",
				chatCompletion.Choices[0].Message, Content)
	}
	start, end := matches[len(matches)-1][0], matches[len(matches)-1][1]
	content := chatCompletion.Choices[0].Message.Content[start:end]
	content = SanitizeJSONString(content)
	err = json.Unmarshal([]byte(content), &r)

	if err != nil {
		log.Debug(Aside("%s: full response: %s", o.Name(), chatCompletion.Choices[0].Message, Content))
		return o.FilterOnFailure, "", err
	}
	filterThisMessage = r.MessageMatches == "false" || r.MessageMatches == false
	var inverted string
	if o.Invert {
		inverted = "(INVERTED)"
		filterThisMessage = !filterThisMessage
	}
	return filterThisMessage, fmt.Sprintf("Decision: %s", r.Reasoning) + inverted, nil
}
