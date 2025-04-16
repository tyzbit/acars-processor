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
	OllamaSystemPrompt = `You are an AI that is an expert at logical 
	reasoning. You will be provided criteria and then a communication message. 
	You will use your skills and any examples provided to evaluate determine 
	if the message positively matches the provided criteria. 
	
	If the message affirmatively matches the criteria, 
	return 'true' in the 'message_matches' field.

	If the message definitely does not match the criteria, return 'false' in the
	'message_matches' field. 
	
	Briefly provide your reasoning regarding your 
	decision in the 'decision' field, pointing out specific evidence that 
	factored in your decision on whether the message matches the criteria.

	Here's the criteria:
	`
	OllamaTimeout             = 120
	OllamaMaxPredictionTokens = 512
	OllamaMaxRetryAttempts    = 6
	OllamaRetryDelaySeconds   = 5
)

type OllamaResponse struct {
	MessageMatches bool   `json:"message_matches"`
	Reasoning      string `json:"reasoning"`
}

type OllamaResponseFormat struct {
	Type       string                                  `json:"type"`
	Properties OllamaResponseFormatRequestedProperties `json:"properties"`
	Required   []string                                `json:"required"`
}

type OllamaResponseFormatRequestedProperties struct {
	MessageMatches OllamaResponseFormatRequestedProperty `json:"matches"`
	Reasoning      OllamaResponseFormatRequestedProperty `json:"reasoning"`
}

type OllamaResponseFormatRequestedProperty struct {
	Type string `json:"type"`
}

var OllamaResponseRequestedFormat = OllamaResponseFormat{
	Type: "object",
	Properties: OllamaResponseFormatRequestedProperties{
		MessageMatches: OllamaResponseFormatRequestedProperty{
			Type: "boolean",
		},
		Reasoning: OllamaResponseFormatRequestedProperty{
			Type: "string",
		},
	},
	Required: []string{"matches", "reasoning"},
}

// Return true if a message passes a filter, false otherwise
func OllamaFilter(m string) bool {
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

	if config.OllamaMaxPredictionTokens != 0 {
		OllamaMaxPredictionTokens = config.OllamaMaxPredictionTokens
	}
	if config.OllamaSystemPrompt != "" {
		OllamaSystemPrompt = config.OllamaSystemPrompt
	}
	if config.OllamaTimeout != 0 {
		OllamaTimeout = config.OllamaTimeout
	}
	if config.OllamaMaxRetryAttempts != 0 {
		OllamaMaxRetryAttempts = config.OllamaMaxRetryAttempts
	}
	if config.OllamaRetryDelaySeconds != 0 {
		OllamaRetryDelaySeconds = config.OllamaRetryDelaySeconds
	}

	stream := false
	requestedFormatJson, err := json.Marshal(OllamaResponseRequestedFormat)
	if err != nil {
		log.Errorf("error setting ollama response format: %s", err)
		return true
	}

	req := &api.GenerateRequest{
		Model:  config.OllamaModel,
		Format: requestedFormatJson,
		System: OllamaSystemPrompt + config.OllamaUserPrompt,
		Stream: &stream,
		Prompt: `Here is the message to evaluate:\n` + m,
		Options: map[string]interface{}{
			// Hopefully minimizes the model timing out
			"num_predict": OllamaMaxPredictionTokens,
			// Make output deterministic
			"temperature": 0,
		},
	}

	var r OllamaResponse
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(OllamaTimeout)*time.Second)
	defer cancel()
	log.Debugf("calling Ollama, model %s", config.OllamaModel)
	err = retry.Do(func() error {
		err = client.Generate(ctx, req, respFunc)
		if err != nil {
			return &RetriableError{
				Err:        fmt.Errorf("error using Ollama: %s", err),
				RetryAfter: time.Duration(OllamaRetryDelaySeconds) * time.Second,
			}
		}
		return nil
	},
		retry.Attempts(uint(OllamaMaxRetryAttempts)),
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
		log.Errorf("too many failures calling Ollama, giving up and %sing: %s", action[!config.OllamaFilterOnFailure], err)
		return !config.OllamaFilterOnFailure
	}

	log.Infof("ollama decision: %s, message ending in: %s, reasoning: %s", action[r.MessageMatches], Last20Characters(m), r.Reasoning)
	return r.MessageMatches
}
