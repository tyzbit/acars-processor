// TODO: Feature parity with Ollama
package main

import (
	"context"
	"encoding/json"
	"reflect"
	"regexp"

	"github.com/fatih/color"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	log "github.com/sirupsen/logrus"
)

var (
	OpenAISystemPrompt = `you are an AI that is an expert at cal 
	reasoning. you will be provided criteria and then a communication message. 
	you will use your skills and any examples provided to evaluate determine 
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

type OpenAIFilterer struct {
	Filterer
	// Whether to filter messages where the OpenAI filter itself fails. Recommended if your ollama instance sometimes returns errors.
	FilterOnFailure bool   `default:"true"`
	APIKey          string `jsonschema:"required" default:"example_key"`
	// Model to use.
	Model string `jsonschema:"required,default=gpt-4o" default:"gpt-4o"`
	// Instructions for OpenAI model to use when filtering messages. More detail is better.
	UserPrompt string `jsonschema:"required,example=Does this message talk about coffee makers or lavatories (shortand LAV is sometimes used)?" default:"Does this message talk about coffee makers or lavatories (shortand LAV is sometimes used)?"`
	// Override the built-in system prompt to instruct the model on how to behave for requests (not usually necessary).
	SystemPrompt string `default:"Answer like a pirate"`
	// How long to wait until giving up on any request to OpenAI.
	Timeout int `default:"5"`
}

func (o OpenAIFilterer) Name() string {
	return reflect.TypeOf(o).Name()
}

type OpenAIResponse struct {
	MessageMatches any    `json:"message_matches"`
	Reasoning      string `json:"reasoning"`
}

// Return true if a message passes a filter, false otherwise
func (o OpenAIFilterer) Filter(m APMessage) (filter bool, reason string, err error) {
	if reflect.DeepEqual(o, OpenAIFilterer{}) {
		return false, "", nil
	}
	ms := GetAPMessageCommonFieldAsString(m, "message_text")

	log.Debug(Emphasised("submitting message ending in \""),
		Note(Last20Characters(ms)),
		Content("\" for filtering with OpenAI"))

	// If message is blank, return
	if regexp.MustCompile(`^\s*$`).MatchString(ms) {
		log.Info(Content("message was blank, filtering without calling OpenAI"))
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

	log.Debug(Aside("calling OpenAI model "), Note(openAIModel))
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
		log.Error(Attention("error using OpenAI: %s", err))
		return o.FilterOnFailure, "", err
	}

	var r OpenAIResponse
	// Parse the JSON payload (hopefully)
	rex := regexp.MustCompile(`\{[^{}]+\}`)
	matches := rex.FindAllStringIndex(chatCompletion.Choices[0].Message.Content, -1)

	// Find the last json payload in case the model reasons about
	// one in the middle of thinking
	if len(matches) == 0 {
		log.Error(Attention("did not find a json object in response: %s",
			chatCompletion.Choices[0].Message, Content))
		return o.FilterOnFailure, "", err
	}
	start, end := matches[len(matches)-1][0], matches[len(matches)-1][1]
	content := chatCompletion.Choices[0].Message.Content[start:end]
	content = SanitizeJSONString(content)
	err = json.Unmarshal([]byte(content), &r)

	if err != nil {
		log.Warn(Attention("error unmarshaling response from OpenAI: %s", err))
		log.Debug(Aside("OpenAI full response: %s", chatCompletion.Choices[0].Message, Content))
		return o.FilterOnFailure, "", err
	}
	decision := r.MessageMatches == "true" || r.MessageMatches == true
	action := map[bool]string{
		true:  Custom(*color.New(color.FgCyan), "allow"),
		false: Custom(*color.New(color.FgYellow), "filter"),
	}
	log.Info(Content("OpenAI decision: "),
		action[decision],
		Content(" for message ending in: \""),
		Note(Last20Characters(ms)),
		Content("\", reasoning: "),
		Emphasised(r.Reasoning))
	return !decision, "final decision", nil
}

func (f OpenAIFilterer) Configured() bool {
	return !reflect.DeepEqual(f, OpenAIFilterer{})
}
