package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
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
	subtractions from the original text. Do not replace or transform words such
	as to modify case unless specifically instructed to.

	If asked to evaluate the message numerically, use your skills and any
	examples rules or criteria given to calculate a numerical result for the
	message.

	Here's the criteria:
	`
	OllamaAnnotatorFinalInstructions = `
	Return true or false corresponding to the answer in the 'question' field.
	Return the processed text in the 'processed_text' field.
	Return any numerical evaluation in the 'processed_value' field.
	Provide feedback summarizing your actions or commentary in the
	'model_feedback' field.
	`
	OllamaAnnotatorTimeout           = 120
	OllamaAnnotatorMaxRetryAttempts  = 6
	OllamaAnnotatorRetryDelaySeconds = 5
)

type OllamaAnnotatorResponse struct {
	Question       bool   `json:"question"`
	ModelFeedback  string `json:"model_feedback"`
	ProcessedText  string `json:"processed_text"`
	ProcessedValue int    `json:"processed_value"`
}

type OllamaAnnotatorResponseFormat struct {
	Type       string                                           `json:"type"`
	Properties OllamaAnnotatorResponseFormatRequestedProperties `json:"properties"`
	Required   []string                                         `json:"required"`
}

type OllamaAnnotatorResponseFormatRequestedProperties struct {
	ProcessedText  OllamaAnnotatorResponseFormatRequestedProperty `json:"processed_text"`
	ProcessedValue OllamaAnnotatorResponseFormatRequestedProperty `json:"processed_value"`
	ModelFeedback  OllamaAnnotatorResponseFormatRequestedProperty `json:"model_feedback"`
	Question       OllamaAnnotatorResponseFormatRequestedProperty `json:"question"`
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
		ProcessedValue: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "integer",
		},
	},
	Required: []string{"question", "model_feedback", "processed_text", "processed_value"},
}

type OllamaAnnotatorHandler struct {
	OllamaAnnotatorResponse
}

func (a OllamaAnnotatorHandler) Name() string {
	return "Ollama"
}

func (a OllamaAnnotatorHandler) SelectFields(annotation Annotation) Annotation {
	if config.Annotators.Ollama.SelectedFields == nil {
		return annotation
	}
	selectedFields := Annotation{}
	for field, value := range annotation {
		if slices.Contains(config.Annotators.Ollama.SelectedFields, field) {
			selectedFields[field] = value
		}
	}
	return selectedFields
}

func (o OllamaAnnotatorHandler) DefaultFields() []string {
	// ACARS
	fields := []string{}
	for field := range o.AnnotateACARSMessage(ACARSMessage{}) {
		fields = append(fields, field)
	}
	slices.Sort(fields)
	return fields
}

func (o OllamaAnnotatorHandler) AnnotateACARSMessage(m ACARSMessage) (annotation Annotation) {
	return o.AnnotateMessage(m.MessageText)
}

func (o OllamaAnnotatorHandler) AnnotateVDLM2Message(m VDLM2Message) (annotation Annotation) {
	return o.AnnotateMessage(m.VDL2.AVLC.ACARS.MessageText)
}

func (o OllamaAnnotatorHandler) AnnotateMessage(m string) (annotation Annotation) {
	enabled := config.Annotators.Ollama.Enabled
	// If message is blank, return
	if enabled && (regexp.MustCompile(`^\s*$`).MatchString(m)) {
		log.Debug(Aside("message was blank, not annotating with Ollama"))
		return
	}
	if enabled && config.Annotators.Ollama.Model == "" || enabled && config.Annotators.Ollama.UserPrompt == "" {
		log.Warn(Attention("OllamaAnnotator model and prompt are required to use the Ollama annotator"))
		return
	}
	url, err := url.Parse(config.Annotators.Ollama.URL)
	if enabled && err != nil {
		log.Error(Attention("OllamaAnnotator url could not be parsed: %s", err))
		return
	}
	client := api.NewClient(url, &http.Client{})
	if enabled && err != nil {
		log.Error(Attention("error initializing OllamaAnnotator: %s", err))
		return
	}

	if config.Annotators.Ollama.SystemPrompt != "" {
		OllamaAnnotatorFirstInstructions = config.Annotators.Ollama.SystemPrompt
	}
	if config.Annotators.Ollama.Timeout != 0 {
		OllamaAnnotatorTimeout = config.Annotators.Ollama.Timeout
	}
	if config.Annotators.Ollama.MaxRetryAttempts != 0 {
		OllamaAnnotatorMaxRetryAttempts = config.Annotators.Ollama.MaxRetryAttempts
	}
	if config.Annotators.Ollama.MaxRetryDelaySeconds != 0 {
		OllamaAnnotatorRetryDelaySeconds = config.Annotators.Ollama.MaxRetryDelaySeconds
	}

	stream := false
	requestedFormatJson, err := json.Marshal(OllamaAnnotatorResponseRequestedFormat)
	if err != nil {
		log.Error(Attention("error setting Ollama response format: %s", err))
		return
	}

	opts := map[string]any{}
	for _, opt := range config.Annotators.Ollama.Options {
		opts[opt.Name] = opt.Value
	}

	req := &api.GenerateRequest{
		Model:  config.Annotators.Ollama.Model,
		Format: requestedFormatJson,
		System: OllamaAnnotatorFirstInstructions + config.Annotators.Ollama.UserPrompt +
			OllamaAnnotatorFinalInstructions,
		Stream:  &stream,
		Prompt:  `Here is the message to evaluate:\n` + m,
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
	log.Debug(Aside("calling OllamaAnnotator, model "), Note(config.Annotators.Ollama.Model))
	err = retry.Do(func() error {
		err = client.Generate(ctx, req, respFunc)
		if err != nil {
			return &RetriableError{
				Err:        fmt.Errorf("error using OllamaAnnotator: %s", err),
				RetryAfter: time.Duration(OllamaAnnotatorRetryDelaySeconds) * time.Second,
			}
		}
		return nil
	},
		retry.Attempts(uint(OllamaAnnotatorMaxRetryAttempts)),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			log.Error(Attention("OllamaAnnotator attempt #%d failed: %v", n+1, err))
		}),
	)

	if config.Annotators.Ollama.FilterWithQuestion {
		if !r.Question {
			log.Info(Note("Ollama annotation response did not qualify according to " +
				"user requirements for the question, not returning any output"))
			return annotation
		}
	}
	if config.Annotators.Ollama.FilterGreaterThan > 0 {
		if r.ProcessedValue <= config.Annotators.Ollama.FilterGreaterThan {
			log.Info(Note("Ollama annotation response did not qualify according to "+
				"user requirements for %d being greater than %d, "+
				"not returning any output", r.ProcessedValue, config.Annotators.Ollama.FilterGreaterThan))
			return annotation
		}
	}
	if config.Annotators.Ollama.FilterLessThan != 0 {
		if r.ProcessedValue >= config.Annotators.Ollama.FilterLessThan {
			log.Info(Note("Ollama annotation response did not qualify according to "+
				"user requirements for %d being less than %d, "+
				"not returning any output", r.ProcessedValue, config.Annotators.Ollama.FilterLessThan))
			return annotation
		}
	}
	if enabled && (r == OllamaAnnotatorResponse{}) {
		log.Info(Attention("Ollama annotator response was empty"))
	} else {
		// Please update config example values if changed
		annotation = Annotation{
			"ollamaProcessedText":     r.ProcessedText,
			"ollamaProcessedValue":    r.ProcessedValue,
			"ollamaModelFeedbackText": r.ModelFeedback,
			"ollamaQuestion":          r.Question,
		}
	}

	return annotation
}
