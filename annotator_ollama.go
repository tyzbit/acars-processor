package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/avast/retry-go"
	api "github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
)

var (
	OllamaAnnotatorFirstInstructions = `You are an expert at precisely
	processing text according to instructions. You will be provided 
	instructions and then a communication message. You may also be asked a
	question about the message.

	You will answer any questions that may have been asked about the message.
	You will use your skills and any examples or rules provided to edit, select,
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
	OllamaAnnotatorTimeout             = 120
	OllamaAnnotatorMaxPredictionTokens = 512
	OllamaAnnotatorMaxRetryAttempts    = 6
	OllamaAnnotatorRetryDelaySeconds   = 5
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

type OllamaHandler struct {
	OllamaAnnotatorResponse
}

func (a OllamaHandler) Name() string {
	return "ollama"
}

func (a OllamaHandler) SelectFields(annotation Annotation) Annotation {
	if config.Annotators.Ollama.SelectedFields == "" {
		return annotation
	}
	selectedFields := Annotation{}
	for field, value := range annotation {
		if strings.Contains(config.Annotators.Ollama.SelectedFields, field) {
			selectedFields[field] = value
		}
	}
	return selectedFields
}

func (o OllamaHandler) AnnotateACARSMessage(m ACARSMessage) (annotation Annotation) {
	return o.AnnotateMessage(m.MessageText)
}

func (o OllamaHandler) AnnotateVDLM2Message(m VDLM2Message) (annotation Annotation) {
	return o.AnnotateMessage(m.VDL2.AVLC.ACARS.MessageText)
}

func (o OllamaHandler) AnnotateMessage(m string) (annotation Annotation) {
	// If message is blank, return
	if regexp.MustCompile(`^\s*$`).MatchString(m) {
		log.Debug("message was blank, not annotating with ollama")
		return
	}
	if config.Annotators.Ollama.Model == "" || config.Annotators.Ollama.UserPrompt == "" {
		log.Warn("OllamaAnnotator model and prompt are required to use the ollama annotator")
		return
	}
	url, err := url.Parse(config.Annotators.Ollama.URL)
	if err != nil {
		log.Errorf("OllamaAnnotator url could not be parsed: %s", err)
		return
	}
	client := api.NewClient(url, &http.Client{})
	if err != nil {
		log.Errorf("error initializing OllamaAnnotator: %s", err)
		return
	}

	if config.Annotators.Ollama.MaxPredictionTokens != 0 {
		OllamaAnnotatorMaxPredictionTokens = config.Annotators.Ollama.MaxPredictionTokens
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
		log.Errorf("error setting ollama response format: %s", err)
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
		log.Debugf("ollama annotator done reason: %s, response: %s", resp.DoneReason, resp.Response)
		if err != nil {
			err = fmt.Errorf("%w, ollama full response: %s", err, resp.Response)
			return err
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(OllamaAnnotatorTimeout)*time.Second)
	defer cancel()
	log.Debugf("calling OllamaAnnotator, model %s", config.Annotators.Ollama.Model)
	err = retry.Do(func() error {
		err = client.Generate(ctx, req, respFunc)
		if err != nil {
			return &RetriableError{
				Err: fmt.Errorf("error using OllamaAnnotator: %s", err)}
		}
		return nil
	},
		retry.Attempts(uint(OllamaAnnotatorMaxRetryAttempts)),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(time.Second*time.Duration(config.Annotators.Ollama.MaxRetryDelaySeconds)),
		retry.OnRetry(func(n uint, err error) {
			log.Errorf("OllamaAnnotator attempt #%d failed: %v", n+1, err)
		}),
	)

	if config.Annotators.Ollama.FilterWithQuestion {
		if !r.Question {
			log.Info("ollama annotation response did not qualify according to " +
				"user requirements, not returning any output")
			return annotation
		}
	}
	if r.ProcessedText == "" && len(r.EditActions) == 0 {
		log.Info("ollama annotator response was empty")
	} else {
		// Please update config example values if changed
		annotation = Annotation{
			"ollamaProcessedText": r.ProcessedText,
			"ollamaEditActions":   r.EditActions,
			"ollamaQuestion":      r.Question,
		}
	}

	return annotation
}
