package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"time"

	"github.com/avast/retry-go"
	api "github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	OllamaFilterFirstInstructions = `You are an AI that is an expert at 
	reasoning about text content. You will be provided criteria and then a 
	communication message. You will use your reasoning and any examples or rules
	provided to determine if the message positively matches the provided
	criteria.

	Here's the criteria:
	`
	OllamaFilterFinalInstructions = `
	If the message definitely matches the criteria, 
	return 'true' in the 'message_matches_criteria' field.

	If the message definitely does not match the criteria, 
	return 'false' in the 'message_matches_criteria' field. 
	
	Provide a very short, high-level explanation with 
	the reasoning for your decision in the "reasoning" field.
	`
	OllamaFilterTimeout             = 120
	OllamaFilterMaxPredictionTokens = 512
	OllamaFilterMaxRetryAttempts    = 6
	OllamaFilterRetryDelaySeconds   = 5
)

type OllamaFilterer struct {
	Filterer
	// Whether to filter messages where Ollama itself fails. Recommended if your ollama instance sometimes returns errors.
	FilterOnFailure bool `default:"false"`
	// Inverse logic (for example, Inverse: true, HasText: true means messages with text are FILTERED)
	Invert bool `json:",omitempty" default:"false"`
	OllamaCommonConfig
}

func (o OllamaFilterer) Name() string {
	return reflect.TypeOf(o).Name()
}

func (f OllamaFilterer) Configured() bool {
	return !reflect.DeepEqual(f, OllamaFilterer{})
}

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

type OllamaFilterRequest struct {
	Model                               string
	OllamaSystemPromptFirstInstructions string
	OllamaUserPrompt                    string
	OllamaSystemPromptFinalInstructions string
	ACARSMessage                        string
}

type OllamaFilterResult struct {
	gorm.Model
	ACARSMessage         string
	OllamaFilterRequest  `gorm:"embedded"`
	OllamaFilterResponse `gorm:"embedded"`
}

// Return true if a message passes a filter, false otherwise
func (o OllamaFilterer) Filter(m APMessage) (filterThisMessage bool, reason string, err error) {
	messageText := GetAPMessageCommonFieldAsString(m, "MessageText")
	if o.Model == "" || o.UserPrompt == "" {
		return false, "", fmt.Errorf("model and prompt are required")
	}
	// If message is blank, return
	if regexp.MustCompile(emptyStringRegex).MatchString(messageText) {
		return true, "message blank", nil
	}
	url, err := url.Parse(o.URL)
	if err != nil {
		return o.FilterOnFailure, "", fmt.Errorf("url could not be parsed: %w", err)
	}
	httpClient := &http.Client{}
	if o.APIKey != "" {
		httpClient = &http.Client{
			Transport: &apiHeaderTransport{
				key:  o.APIKey,
				base: http.DefaultTransport,
			},
		}
	}
	client := api.NewClient(url, httpClient)
	if err != nil {
		return o.FilterOnFailure, "", fmt.Errorf("error initializing: %w", err)
	}

	// There are defaults for these which is why we're not using
	// the settings directly.
	if o.SystemPrompt != "" {
		OllamaFilterFirstInstructions = o.SystemPrompt
	}
	if o.Timeout != 0 {
		OllamaFilterTimeout = o.Timeout
	}
	if o.MaxRetryAttempts != 0 {
		OllamaFilterMaxRetryAttempts = o.MaxRetryAttempts
	}
	if o.MaxRetryDelaySeconds != 0 {
		OllamaFilterRetryDelaySeconds = o.MaxRetryDelaySeconds
	}

	stream := false
	requestedFormatJson, err := json.Marshal(OllamaFilterResponseRequestedFormat)
	if err != nil {
		return o.FilterOnFailure, "", fmt.Errorf("error setting response format: %w", err)
	}

	opts := map[string]any{}
	for _, opt := range o.Options {
		opts[opt.Name] = opt.Value
	}

	systemPrompt := OllamaFilterFirstInstructions + o.UserPrompt +
		OllamaFilterFinalInstructions
	req := &api.GenerateRequest{
		Model:   o.Model,
		Format:  requestedFormatJson,
		System:  systemPrompt,
		Stream:  &stream,
		Prompt:  messageText,
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
			err = fmt.Errorf("%s, full response: %s", err, resp.Response)
			return err
		}
		ofr := OllamaFilterResult{
			ACARSMessage: messageText,
			OllamaFilterRequest: OllamaFilterRequest{
				Model:                               o.Model,
				OllamaSystemPromptFirstInstructions: OllamaFilterFirstInstructions,
				OllamaUserPrompt:                    o.UserPrompt,
				OllamaSystemPromptFinalInstructions: OllamaFilterFinalInstructions,
				ACARSMessage:                        messageText,
			},
			OllamaFilterResponse: r,
		}
		db.Create(&ofr)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(OllamaFilterTimeout)*time.Second)
	defer cancel()
	log.Debug(Aside("%s: considering message ending in \"", o.Name()),
		Note(Last20Characters(messageText)),
		Aside("\", model "),
		Note(o.Model))
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
	)

	if err != nil {
		return o.FilterOnFailure, "too many failures", err
	}
	filterThisMessage = !r.MessageMatchesCriteria
	var inverted string
	if o.Invert {
		inverted = "(INVERTED)"
		filterThisMessage = !filterThisMessage
	}
	return filterThisMessage, fmt.Sprintf("Decision: %s", r.Reasoning) + inverted, nil
}
