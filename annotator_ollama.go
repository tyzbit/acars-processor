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
	OllamaAnnotatorFirstInstructions = `You are an AI that is an expert at 
	precisely processing text according to instructions. You will be provided 
	instructions and then a communication message. You may also be asked a
	qualifying question about the message.

	You will use your skills and any examples or rules provided to edit, select
	transform or otherwise process the text strictly according to the directions
	given. Preserve the words from the original text exactly unless specifically
	instructed otherwise.

	Here's the criteria:
	`
	OllamaAnnotatorFinalInstructions = `
	Return the text after making the requested changes in the 'processed_text'
	field.

	Note each action you took to process the text in 'edit_actions'.

	Return true or false based on any qualifying questions provided in 
	'qualifier'
	`
	OllamaAnnotatorTimeout             = 120
	OllamaAnnotatorMaxPredictionTokens = 512
	OllamaAnnotatorMaxRetryAttempts    = 6
	OllamaAnnotatorRetryDelaySeconds   = 5
)

type OllamaAnnotatorResponse struct {
	ProcessedText string `json:"processed_text"`
	EditActions   string `json:"edit_actions"`
	Qualifier     bool   `json:"qualifier"`
}

type OllamaAnnotatorResponseFormat struct {
	Type       string                                           `json:"type"`
	Properties OllamaAnnotatorResponseFormatRequestedProperties `json:"properties"`
	Required   []string                                         `json:"required"`
}

type OllamaAnnotatorResponseFormatRequestedProperties struct {
	ProcessedText OllamaAnnotatorResponseFormatRequestedProperty `json:"processed_text"`
	EditActions   OllamaAnnotatorResponseFormatRequestedProperty `json:"edit_actions"`
	Qualifier     OllamaAnnotatorResponseFormatRequestedProperty `json:"qualifier"`
}

type OllamaAnnotatorResponseFormatRequestedProperty struct {
	Type string `json:"type"`
}

var OllamaAnnotatorResponseRequestedFormat = OllamaAnnotatorResponseFormat{
	Type: "object",
	Properties: OllamaAnnotatorResponseFormatRequestedProperties{
		ProcessedText: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "string",
		},
		EditActions: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "string",
		},
		Qualifier: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "boolean",
		},
	},
	Required: []string{"processed_text", "edit_actions"},
}

type OllamaHandler struct{}

func (a OllamaHandler) Name() string {
	return "acars"
}

func (a OllamaHandler) SelectFields(annotation Annotation) Annotation {
	if config.ACARSAnnotatorSelectedFields == "" {
		return annotation
	}
	selectedFields := Annotation{}
	for field, value := range annotation {
		if strings.Contains(config.OllamaAnnotatorSelectedFields, field) {
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
	if config.OllamaAnnotatorModel == "" || config.OllamaAnnotatorUserPrompt == "" {
		log.Warn("OllamaAnnotator model and prompt are required to use the ollama filter")
		return
	}
	url, err := url.Parse(config.OllamaAnnotatorURL)
	if err != nil {
		log.Errorf("OllamaAnnotator url could not be parsed: %s", err)
		return
	}
	client := api.NewClient(url, &http.Client{})
	if err != nil {
		log.Errorf("error initializing OllamaAnnotator: %s", err)
		return
	}

	if config.OllamaAnnotatorMaxPredictionTokens != 0 {
		OllamaAnnotatorMaxPredictionTokens = config.OllamaAnnotatorMaxPredictionTokens
	}
	if config.OllamaAnnotatorSystemPrompt != "" {
		OllamaAnnotatorFirstInstructions = config.OllamaAnnotatorSystemPrompt
	}
	if config.OllamaAnnotatorTimeout != 0 {
		OllamaAnnotatorTimeout = config.OllamaAnnotatorTimeout
	}
	if config.OllamaAnnotatorMaxRetryAttempts != 0 {
		OllamaAnnotatorMaxRetryAttempts = config.OllamaAnnotatorMaxRetryAttempts
	}
	if config.OllamaAnnotatorRetryDelaySeconds != 0 {
		OllamaAnnotatorRetryDelaySeconds = config.OllamaAnnotatorRetryDelaySeconds
	}

	stream := false
	requestedFormatJson, err := json.Marshal(OllamaAnnotatorResponseRequestedFormat)
	if err != nil {
		log.Errorf("error setting ollama response format: %s", err)
		return
	}

	req := &api.GenerateRequest{
		Model:  config.OllamaAnnotatorModel,
		Format: requestedFormatJson,
		System: OllamaAnnotatorFirstInstructions + config.OllamaAnnotatorUserPrompt +
			OllamaAnnotatorFinalInstructions,
		Stream: &stream,
		Prompt: `Here is the message to evaluate:\n` + m,
		Options: map[string]interface{}{
			// Hopefully minimizes the model timing out
			"num_predict": OllamaAnnotatorMaxPredictionTokens,
			// Make output deterministic
			"temperature": 0,
		},
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
		log.Debugf("ollama annotator response: %s", resp.Response)
		if err != nil {
			err = fmt.Errorf("%w, ollama full response: %s", err, resp.Response)
			return err
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(OllamaAnnotatorTimeout)*time.Second)
	defer cancel()
	log.Debugf("calling OllamaAnnotator, model %s", config.OllamaAnnotatorModel)
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
			log.Errorf("OllamaAnnotator attempt #%d failed: %v", n+1, err)
		}),
	)

	if r.ProcessedText == "" && r.EditActions == "" {
		log.Info("ollama annotator response was empty")
	} else {
		annotation = Annotation{
			"ollamaProcessedText": r.ProcessedText,
			"ollamaEditActions":   r.EditActions,
			"ollamaQualifier":     r.Qualifier,
		}
	}

	return annotation
}
