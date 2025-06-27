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
	OllamaAnnotatorFirstInstructions = `you are an expert at precisely
	processing text according to instructions. you will be provided 
	instructions and then a communication message. you may also be asked a
	question about the message.

	you will answer any questions that may have been asked about the message.
	you will use your skills and any examples or rules provided to edit, select,
	transform or otherwise process the text strictly according to the directions
	given. Only make additions or subtractions from the original text. Do not
	replace or transform words such as to modify case unless specifically
	instructed to.

	Here's the criteria:
	`
	OllamaAnnotatorFinalInstructions = `
	If you were asked a question, return true or false corresponding to the
	answer in the 'question' field. If you weren't asked a question,
	return true in the 'question' field.

	Note each action you took to process the text in 'edit_actions'.

	Return the processed text in the 'processed_text' field.
	`
	OllamaAnnotatorTimeout           = 120
	OllamaAnnotatorMaxRetryAttempts  = 6
	OllamaAnnotatorRetryDelaySeconds = 5
)

type OllamaAnnotatorResponse struct {
	Question      bool   `json:"question"`
	EditActions   string `json:"edit_actions"`
	ProcessedText string `json:"processed_text"`
}

type OllamaAnnotatorResponseFormat struct {
	Type       string                                           `json:"type"`
	Properties OllamaAnnotatorResponseFormatRequestedProperties `json:"properties"`
	Required   []string                                         `json:"required"`
}

type OllamaAnnotatorResponseFormatRequestedProperties struct {
	ProcessedText OllamaAnnotatorResponseFormatRequestedProperty `json:"processed_text"`
	EditActions   OllamaAnnotatorResponseFormatRequestedProperty `json:"edit_actions"`
	Question      OllamaAnnotatorResponseFormatRequestedProperty `json:"question"`
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
		EditActions: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "string",
		},
		ProcessedText: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "string",
		},
	},
	Required: []string{"question", "edit_actions", "processed_text"},
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
		log.Debug(yo.INFODUMP("message was blank, not annotating with Ollama").FRFR())
		return
	}
	if enabled && config.Annotators.Ollama.Model == "" || enabled && config.Annotators.Ollama.UserPrompt == "" {
		log.Warn(yo.Uhh("OllamaAnnotator model and prompt are required to use the Ollama annotator").FRFR())
		return
	}
	url, err := url.Parse(config.Annotators.Ollama.URL)
	if enabled && err != nil {
		log.Error(yo.Uhh("OllamaAnnotator url could not be parsed: %s", err).FRFR())
		return
	}
	client := api.NewClient(url, &http.Client{})
	if enabled && err != nil {
		log.Error(yo.Uhh("error initializing OllamaAnnotator: %s", err).FRFR())
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
		log.Error(yo.Uhh("error setting Ollama response format: %s", err).FRFR())
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
		log.Debug(yo.INFODUMP("Ollama annotator done reason: %s, response: %s", resp.DoneReason, resp.Response).FRFR())
		if err != nil {
			err = fmt.Errorf("%s, Ollama full response: %s", err, resp.Response)
			return err
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(OllamaAnnotatorTimeout)*time.Second)
	defer cancel()
	log.Debug(yo.INFODUMP("calling OllamaAnnotator, model ").Hmm(config.Annotators.Ollama.Model).FRFR())
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
			log.Error(yo.Uhh("OllamaAnnotator attempt #%d failed: %v", n+1, err).FRFR())
		}),
	)

	if config.Annotators.Ollama.FilterWithQuestion {
		if !r.Question {
			log.Info(yo.Hmm("Ollama annotation response did not qualify according to " +
				"user requirements, not returning any output"))
			return annotation
		}
	}
	if enabled && r.ProcessedText == "" && len(r.EditActions) == 0 {
		log.Info(yo.Uhh("Ollama annotator response was empty").FRFR())
	} else {
		// Please update config example values if changed
		annotation = Annotation{
			"OllamaProcessedText": r.ProcessedText,
			"OllamaEditActions":   r.EditActions,
			"OllamaQuestion":      r.Question,
		}
	}

	return annotation
}
