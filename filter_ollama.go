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
	OllamaFilterFirstInstructions = `You are an AI that is an expert at logical 
	reasoning. You will be provided criteria and then a communication message. 
	You will use your skills and any examples provided to evaluate determine 
	if the message positively matches the provided criteria. 

	Here's the criteria:
	`
	OllamaFilterFinalInstructions = `
	If the message definitely matches the criteria, 
	return 'true' in the 'message_matches_criteria' field.

	If the message definitely does not match the criteria, 
	return 'false' in the 'message_matches_criteria' field. 
	
	Provide a very short, high-level explanation as to the reasoning
	for your decision in the "reasoning" field.
	`
	OllamaFilterTimeout             = 120
	OllamaFilterMaxPredictionTokens = 512
	OllamaFilterMaxRetryAttempts    = 6
	OllamaFilterRetryDelaySeconds   = 5
)

type OllamaFilterResponse struct {
	MessageMatchesCriteria bool   `json:"message_matches_criteria"`
	Reasoning              string `json:"reasoning"`
}

type OllamaFilterResponseFormat struct {
	Type       string                                        `json:"type"`
	Properties OllamaFilterResponseFormatRequestedProperties `json:"properties"`
	Required   []string                                      `json:"required"`
}

type OllamaFilterResponseFormatRequestedProperties struct {
	MessageMatchesCriteria OllamaFilterResponseFormatRequestedProperty `json:"message_matches_criteria"`
	Reasoning              OllamaFilterResponseFormatRequestedProperty `json:"reasoning"`
}

type OllamaFilterResponseFormatRequestedProperty struct {
	Type string `json:"type"`
}

var OllamaFilterResponseRequestedFormat = OllamaFilterResponseFormat{
	Type: "object",
	Properties: OllamaFilterResponseFormatRequestedProperties{
		MessageMatchesCriteria: OllamaFilterResponseFormatRequestedProperty{
			Type: "boolean",
		},
		Reasoning: OllamaFilterResponseFormatRequestedProperty{
			Type: "string",
		},
	},
	Required: []string{"message_matches_criteria", "reasoning"},
}

// Return true if a message passes a filter, false otherwise
func OllamaFilter(m string) bool {
	// If message is blank, return
	if regexp.MustCompile(`^\s*$`).MatchString(m) {
		log.Debug("message was blank, filtering without calling ollama")
		return false
	}
	if config.Filters.Ollama.Model == "" || config.Filters.Ollama.UserPrompt == "" {
		log.Warn("OllamaFilter model and prompt are required to use the ollama filter")
		return true
	}
	url, err := url.Parse(config.Filters.Ollama.URL)
	if err != nil {
		log.Errorf("OllamaFilter url could not be parsed: %s", err)
		return true
	}
	client := api.NewClient(url, &http.Client{})
	if err != nil {
		log.Errorf("error initializing OllamaFilter: %s", err)
		return true
	}

	if config.Filters.Ollama.MaxPredictionTokens != 0 {
		OllamaFilterMaxPredictionTokens = config.Filters.Ollama.MaxPredictionTokens
	}
	if config.Filters.Ollama.SystemPrompt != "" {
		OllamaFilterFirstInstructions = config.Filters.Ollama.SystemPrompt
	}
	if config.Filters.Ollama.Timeout != 0 {
		OllamaFilterTimeout = config.Filters.Ollama.Timeout
	}
	if config.Filters.Ollama.MaxRetryAttempts != 0 {
		OllamaFilterMaxRetryAttempts = config.Filters.Ollama.MaxRetryAttempts
	}
	if config.Filters.Ollama.MaxRetryDelaySeconds != 0 {
		OllamaFilterRetryDelaySeconds = config.Filters.Ollama.MaxRetryDelaySeconds
	}

	stream := false
	requestedFormatJson, err := json.Marshal(OllamaFilterResponseRequestedFormat)
	if err != nil {
		log.Errorf("error setting ollama response format: %s", err)
		return true
	}

	opts := map[string]any{}
	for _, opt := range config.Filters.Ollama.Options {
		opts[opt.Name] = opt.Value
	}

	req := &api.GenerateRequest{
		Model:  config.Filters.Ollama.Model,
		Format: requestedFormatJson,
		System: OllamaFilterFirstInstructions + config.Filters.Ollama.UserPrompt +
			OllamaFilterFinalInstructions,
		Stream:  &stream,
		Prompt:  `Here is the message to evaluate:\n` + m,
		Options: opts,
	}

	var r OllamaFilterResponse
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(OllamaFilterTimeout)*time.Second)
	defer cancel()
	log.Debugf("calling OllamaFilter, model %s", config.Filters.Ollama.Model)
	err = retry.Do(func() error {
		err = client.Generate(ctx, req, respFunc)
		if err != nil {
			return &RetriableError{
				Err:        fmt.Errorf("error using OllamaFilter: %s", err),
				RetryAfter: time.Duration(OllamaFilterRetryDelaySeconds) * time.Second,
			}
		}
		return nil
	},
		retry.Attempts(uint(OllamaFilterMaxRetryAttempts)),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(time.Second*time.Duration(config.Filters.Ollama.MaxRetryDelaySeconds)),
		retry.OnRetry(func(n uint, err error) {
			log.Errorf("OllamaFilter attempt #%d failed: %v", n+1, err)
		}),
	)

	action := map[bool]string{
		true:  "allow",
		false: "filter",
	}

	if err != nil {
		log.Errorf("too many failures calling OllamaFilter, giving up and %sing: %s", action[!config.Filters.Ollama.FilterOnFailure], err)
		return !config.Filters.Ollama.FilterOnFailure
	}

	log.Infof("ollama decision: %s, message ending in: %s, reasoning: %s", action[r.MessageMatchesCriteria], Last20Characters(m), r.Reasoning)
	return r.MessageMatchesCriteria
}
