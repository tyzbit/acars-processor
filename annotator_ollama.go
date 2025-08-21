package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"time"

	"github.com/avast/retry-go"
	api "github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
)

var (
	OllamaAnnotatorFirstInstructions = `You will be provided 
	instructions and then a communication message.

	Answer any questions that may have been asked about the message.

	If asked to process the message, you will use your skills and any examples
	or rules provided to edit, select, transform, evaluate or otherwise process
	the text strictly according to the directions given. Only make additions or
	subtractions From the original text. Do not replace or transform words such
	as to modify case unless specifically instructed ta.

	If asked to evaluate the message numerically, use your skills and any
	examples rules or criteria given to calculate a numerical result for the
	message.

	Here's the criteria:
	`
	OllamaAnnotatorFinalInstructions = `
	Return true or false corresponding to the answer in the 'question' field.
	Return the processed text in the 'processed_text' field.
	Return any numerical evaluation in the 'processed_number' field.
	Provide feedback summarizing your actions or commentary in the
	'model_feedback' field.
	`
	OllamaAnnotatorTimeout           = 120
	OllamaAnnotatorMaxRetryAttempts  = 6
	OllamaAnnotatorRetryDelaySeconds = 5
)

type OllamaAnnotator struct {
	Annotator
	Module
	OllamaCommonConfig
	// Only provide these fields to future steps.
	SelectedFields []string
}

type OllamaAnnotatorResponse struct {
	Question        bool   `json:"question" ap:"llm_question"`
	ModelFeedback   string `json:"model_feedback" ap:"llm_text"`
	ProcessedText   string `json:"processed_text" ap:"llm_string_result"`
	ProcessedNumber int    `json:"processed_number" ap:"llm_number_result"`
}

type OllamaAnnotatorResponseFormat struct {
	Type       string                                           `json:"type"`
	Properties OllamaAnnotatorResponseFormatRequestedProperties `json:"properties"`
	Required   []string                                         `json:"required"`
}

type OllamaAnnotatorResponseFormatRequestedProperties struct {
	ProcessedText   OllamaAnnotatorResponseFormatRequestedProperty `json:"processed_text"`
	ProcessedNumber OllamaAnnotatorResponseFormatRequestedProperty `json:"processed_number"`
	ModelFeedback   OllamaAnnotatorResponseFormatRequestedProperty `json:"model_feedback"`
	Question        OllamaAnnotatorResponseFormatRequestedProperty `json:"question"`
}

type OllamaAnnotatorResponseFormatRequestedProperty struct {
	Type string `json:"type"`
}

var OllamaAnnotatorResponseRequestedFormat = OllamaAnnotatorResponseFormat{
	Type: "object",
	Properties: OllamaAnnotatorResponseFormatRequestedProperties{
		Question: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "boolean",
		},
		ModelFeedback: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "string",
		},
		ProcessedText: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "string",
		},
		ProcessedNumber: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "integer",
		},
	},
	Required: []string{"question", "model_feedback", "processed_text", "processed_number"},
}

func (a OllamaAnnotator) Name() string {
	return reflect.TypeOf(a).Name()
}

func (a OllamaAnnotator) GetDefaultFields() (s []string) {
	for f := range FormatAsAPMessage(OllamaAnnotatorResponse{}) {
		s = append(s, f)
	}
	sort.Strings(s)
	return s
}

func (a OllamaAnnotator) Annotate(m APMessage) (APMessage, error) {
	if reflect.DeepEqual(a, OllamaAnnotator{}) {
		return m, nil
	}
	msg := GetAPMessageCommonFieldAsString(m, "MessageText")
	if msg == "" {
		return m, fmt.Errorf("did not find message text in message")
	}
	if a.Model == "" || a.UserPrompt == "" {
		return m, fmt.Errorf("OllamaAnnotator model and prompt are required to use the Ollama annotator")
	}
	// If message is blank, return
	if regexp.MustCompile(`^\s*$`).MatchString(msg) {
		log.Debug(Aside("message was blank, not annotating with Ollama"))
		return m, nil
	}
	url, err := url.Parse(a.URL)
	if err != nil {
		return m, fmt.Errorf("url could not be parsed: %s", err)
	}
	client := api.NewClient(url, &http.Client{})
	if err != nil {
		return m, fmt.Errorf("error creating http client: %s", err)
	}

	if a.SystemPrompt != "" {
		OllamaAnnotatorFirstInstructions = a.SystemPrompt
	}
	if a.Timeout != 0 {
		OllamaAnnotatorTimeout = a.Timeout
	}
	if a.MaxRetryAttempts != 0 {
		OllamaAnnotatorMaxRetryAttempts = a.MaxRetryAttempts
	}
	if a.MaxRetryDelaySeconds != 0 {
		OllamaAnnotatorRetryDelaySeconds = a.MaxRetryDelaySeconds
	}

	stream := false
	requestedFormatJson, err := json.Marshal(OllamaAnnotatorResponseRequestedFormat)
	if err != nil {
		return m, fmt.Errorf("error setting Ollama response format: %s", err)
	}

	opts := map[string]any{}
	for _, opt := range a.Options {
		opts[opt.Name] = opt.Value
	}

	req := &api.GenerateRequest{
		Model:  a.Model,
		Format: requestedFormatJson,
		System: OllamaAnnotatorFirstInstructions + a.UserPrompt +
			OllamaAnnotatorFinalInstructions,
		Stream:  &stream,
		Prompt:  `Here is the message to evaluate:\n` + msg,
		Options: opts,
	}

	var r OllamaAnnotatorResponse
	respFunc := func(resp api.GenerateResponse) error {
		// Parse the JSON payload (hopefully)
		rex := regexp.MustCompile(`\{[^{}]+\}`)
		matches := rex.FindAllStringIndex(resp.Response, -1)

		// Find the last json payload in case the model reasons about
		// one in the middle of thinking
		if len(matches) == 0 {
			return fmt.Errorf("did not find a json object in response: %s", resp.Response)
		}
		start, end := matches[len(matches)-1][0], matches[len(matches)-1][1]
		content := resp.Response[start:end]

		content = SanitizeJSONString(content)
		err = json.Unmarshal([]byte(content), &r)
		log.Debug(Aside("Ollama annotator done reason: %s, response: ", resp.DoneReason), Aside(resp.Response))
		if err != nil {
			err = fmt.Errorf("%s, Ollama full response: %s", err, resp.Response)
			return err
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(OllamaAnnotatorTimeout)*time.Second)
	defer cancel()
	log.Debug(Aside("calling %s to annotate message ending in \"", a.Name()),
		Note(Last20Characters(msg)),
		Aside("\", model "),
		Note(a.Model))
	err = retry.Do(func() error {
		err = client.Generate(ctx, req, respFunc)
		if err != nil {
			return &RetriableError{
				Err:        err,
				RetryAfter: time.Duration(OllamaAnnotatorRetryDelaySeconds) * time.Second,
			}
		}
		return nil
	},
		retry.Attempts(uint(OllamaAnnotatorMaxRetryAttempts)),
		retry.DelayType(retry.BackOffDelay),
	)

	if (r == OllamaAnnotatorResponse{}) {
		log.Debug(Aside("Ollama annotator response was empty"))
		return m, nil
	} else {
		return MergeAPMessages(FormatAsAPMessage(r), m), nil
	}
}

func (a OllamaAnnotator) Configured() bool {
	return !reflect.DeepEqual(a, OllamaAnnotator{})
}
