package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/avast/retry-go"
	api "github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
)

var (
	OllamaAnnotatorSystemPrompt = `You are an expert at transforming messages
	according to rules that are provided to you. You will be provided
	a communication message. You will use your skills and any examples provided
	to transform the message according to the rules.

	The rules you should use to transform the message are:
	`
	OllamaAnnotatorFinalInstructions = `
	If the message definitely matches the criteria, 
	return 'true' in the 'message_matches_criteria' field.

	If the message definitely does not match the criteria, 
	return 'false' in the 'message_matches_criteria' field. 
	
	Briefly provide your reasoning regarding your
	decision in the 'reasoning' field, pointing out specific evidence that 
	factored in your decision on whether the message matches the criteria.
	`
	OllamaAnnotatorTimeout             = 120
	OllamaAnnotatorMaxPredictionTokens = 512
	OllamaAnnotatorMaxRetryAttempts    = 6
	OllamaAnnotatorRetryDelaySeconds   = 5
)

type OllamaAnnotatorResponse struct {
	MessageMatchesCriteria bool   `json:"message_matches_criteria"`
	Reasoning              string `json:"reasoning"`
}

type OllamaAnnotatorResponseFormat struct {
	Type       string                                           `json:"type"`
	Properties OllamaAnnotatorResponseFormatRequestedProperties `json:"properties"`
	Required   []string                                         `json:"required"`
}

type OllamaAnnotatorResponseFormatRequestedProperties struct {
	MessageMatchesCriteria OllamaAnnotatorResponseFormatRequestedProperty `json:"message_matches_criteria"`
	Reasoning              OllamaAnnotatorResponseFormatRequestedProperty `json:"reasoning"`
}

type OllamaAnnotatorResponseFormatRequestedProperty struct {
	Type string `json:"type"`
}

var OllamaAnnotatorResponseRequestedFormat = OllamaAnnotatorResponseFormat{
	Type: "object",
	Properties: OllamaAnnotatorResponseFormatRequestedProperties{
		MessageMatchesCriteria: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "boolean",
		},
		Reasoning: OllamaAnnotatorResponseFormatRequestedProperty{
			Type: "string",
		},
	},
	Required: []string{"message_matches_criteria", "reasoning"},
}

// Return true if a message passes a filter, false otherwise
func OllamaAnnotator(m string) bool {
	// If message is blank, return
	if regexp.MustCompile(`^\s*$`).MatchString(m) {
		log.Debug("message was blank, filtering without calling ollama")
		return false
	}
	if config.OllamaModel == "" || config.OllamaUserPrompt == "" {
		log.Warn("Ollama model and prompt are required to use the ollama filter")
		return true
	}
	url, err := url.Parse(config.OllamaURL)
	if err != nil {
		log.Errorf("Ollama url could not be parsed: %s", err)
		return true
	}
	client := api.NewClient(url, &http.Client{})
	if err != nil {
		log.Errorf("error initializing Ollama: %s", err)
		return true
	}

	if config.OllamaAnnotatorMaxPredictionTokens != 0 {
		OllamaAnnotatorMaxPredictionTokens = config.OllamaAnnotatorMaxPredictionTokens
	}
	if config.OllamaAnnotatorSystemPrompt != "" {
		OllamaAnnotatorSystemPrompt = config.OllamaAnnotatorSystemPrompt
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
		return true
	}

	req := &api.GenerateRequest{
		Model:  config.OllamaModel,
		Format: requestedFormatJson,
		System: OllamaAnnotatorSystemPrompt + config.OllamaUserPrompt +
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
		if err != nil {
			err = fmt.Errorf("%w, ollama full response: %s", err, resp.Response)
			return err
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(OllamaAnnotatorTimeout)*time.Second)
	defer cancel()
	log.Debugf("calling Ollama, model %s", config.OllamaModel)
	err = retry.Do(func() error {
		err = client.Generate(ctx, req, respFunc)
		if err != nil {
			return &RetriableError{
				Err:        fmt.Errorf("error using Ollama: %s", err),
				RetryAfter: time.Duration(OllamaAnnotatorRetryDelaySeconds) * time.Second,
			}
		}
		return nil
	},
		retry.Attempts(uint(OllamaAnnotatorMaxRetryAttempts)),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			log.Errorf("Ollama attempt #%d failed: %v", n+1, err)
		}),
	)

	action := map[bool]string{
		true:  "allow",
		false: "filter",
	}

	if err != nil {
		log.Errorf("too many failures calling Ollama, giving up and %sing: %s", action[!config.OllamaAnnotatorOnFailure], err)
		return !config.OllamaAnnotatorOnFailure
	}

	log.Infof("ollama decision: %s, message ending in: %s, reasoning: %s", action[r.MessageMatchesCriteria], Last20Characters(m), r.Reasoning)
	return r.MessageMatchesCriteria
}
