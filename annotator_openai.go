package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	log "github.com/sirupsen/logrus"
)

var (
	OpenAIAnnotatorFirstInstructions = `You will be provided 
	instructions and then a communication message.

	Answer any questions that may have been asked about the message.

	If asked to process the message, you will use your skills and any examples
	or rules provided to edit, select, transform, evaluate or otherwise process
	the text strictly according to the directions given. Only make additions or
	subtractions from the original text. Do not replace or transform words such
	as to modify case unless specifically instructed to.

	If asked to evaluate the message numerically, use your skills and any
	examples rules or criteria given to calculate a numerical result for the
	message.

	Here's the criteria:
	`
	OpenAIAnnotatorFinalInstructions = `
	Return true or false corresponding to the answer in the 'question' field.
	Return the processed text in the 'processed_text' field.
	Return any numerical evaluation in the 'processed_number' field.
	Provide feedback summarizing your actions or commentary in the
	'model_feedback' field.
	`
)

type OpenAIAnnotator struct {
	Annotator
	Module
	// Only provide these fields to future steps.
	SelectedFields []string
	// Whether to filter messages where the OpenAI annotator itself fails. Recommended if your OpenAI instance sometimes returns errors.
	FilterOnFailure bool `default:"true"`
	APIKey          string
	// Model to use.
	Model string `jsonschema:"required,default=gpt-4o" default:"gpt-4o"`
	// URL to the OpenAI-compatible API endpoint (include protocol and port).
	URL string `jsonschema:"required,example=http://llama-server:8080/v1" default:"http://localhost:8080/v1"`
	// Override the built-in system prompt to instruct the model on how to behave for requests (not usually necessary).
	SystemPrompt string `default:"Answer like a pirate"`
	// Instructions for OpenAI model to use when annotating messages. More detail is better.
	UserPrompt string `jsonschema:"required,example=Rate the emotional content of this message between 1-100." default:"Tell the LLM how to handle the message"`
	// How long to wait until giving up on any request to OpenAI.
	Timeout int `default:"60"`
}

type OpenAIAnnotatorResponse struct {
	ModelFeedbackText   string `json:"model_feedback" ap:"LLMModelFeedbackText"`
	ProcessedNumber     int    `json:"processed_number" ap:"LLMProcessedNumber"`
	ProcessedText       string `json:"processed_text" ap:"LLMProcessedText"`
	YesNoQuestionAnswer bool   `json:"question_answer" ap:"LLMYesNoQuestionAnswer"`
}

func (a OpenAIAnnotator) Name() string {
	return reflect.TypeOf(a).Name()
}

func (a OpenAIAnnotator) GetDefaultFields() (s []string) {
	for f := range FormatAsAPMessage(OpenAIAnnotatorResponse{}, a.Name()) {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
}

func (a OpenAIAnnotator) Configured() bool {
	return !reflect.DeepEqual(a, OpenAIAnnotator{})
}

func (a OpenAIAnnotator) Annotate(m APMessage) (APMessage, error) {
	msg := GetAPMessageCommonFieldAsString(m, "MessageText")
	// If message is blank, return
	if regexp.MustCompile(emptyStringRegex).MatchString(msg) {
		log.Debug(Aside("%s: message was blank, not annotating", a.Name()))
		return m, nil
	}
	if a.Model == "" || a.UserPrompt == "" {
		return m, fmt.Errorf("model and prompt are required to use the OpenAI annotator")
	}

	baseURL := a.URL
	if baseURL == "" {
		baseURL = "http://localhost:8080/v1"
	}

	opts := []option.RequestOption{}
	if a.APIKey != "" {
		opts = append(opts, option.WithAPIKey(a.APIKey))
	}
	opts = append(opts, option.WithBaseURL(baseURL))

	client := openai.NewClient(opts...)

	if a.SystemPrompt != "" {
		OpenAIAnnotatorFirstInstructions = a.SystemPrompt
	}

	log.Debug(Aside("%s: annotating message ending in \"", a.Name()),
		Note(Last20Characters(msg)),
		Aside("\", model "),
		Note(a.Model))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout)*time.Second)
	defer cancel()

	chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(OpenAIAnnotatorFirstInstructions + a.UserPrompt + OpenAIAnnotatorFinalInstructions),
			openai.UserMessage("Here is the message to evaluate:\n" + msg),
		}),
		Model:       openai.F(a.Model),
		Temperature: openai.Float(0),
	})

	if err != nil {
		return m, fmt.Errorf("error calling OpenAI API: %s", err)
	}

	var r OpenAIAnnotatorResponse
	content := chatCompletion.Choices[0].Message.Content
	// Parse the JSON payload (hopefully)
	rex := regexp.MustCompile(`\{[^{}]+\}`)
	matches := rex.FindAllStringIndex(content, -1)

	// Find the last json payload in case the model reasons about
	// one in the middle of thinking
	if len(matches) == 0 {
		return m, fmt.Errorf("did not find a json object in response: %s", Aside(content))
	}
	start, end := matches[len(matches)-1][0], matches[len(matches)-1][1]
	content = content[start:end]
	content = SanitizeJSONString(content)
	err = json.Unmarshal([]byte(content), &r)

	if err != nil {
		log.Debug(Aside("%s: full response: %s", a.Name(), content))
		return m, fmt.Errorf("did not find a valid json object in response: %s", Aside(content))
	}

	if (r == OpenAIAnnotatorResponse{}) {
		log.Debug(Aside("%s: response was empty", a.Name()))
		return m, nil
	} else {
		// This ensures the field is never zero
		r.ProcessedNumber = min(100, max(1, r.ProcessedNumber))
		return MergeAPMessages(FormatAsAPMessage(r, a.Name()), m), nil
	}
}
