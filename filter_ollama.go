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
	"github.com/fatih/color"
	api "github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	OllamaFilterFirstInstructions = `You are an AI that is an expert at cal 
	reasoning. you will be provided criteria and then a communication message. 
	you will use your skills and any examples provided to evaluate determine 
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

type OllamaFilterer struct {
	Filterer
	// Whether to filter messages where Ollama itself fails. Recommended if your ollama instance sometimes returns errors.
	FilterOnFailure bool `default:"true"`
	OllamaCommonConfig
}

func (o OllamaFilterer) Name() string {
	return reflect.TypeOf(o).Name()
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
func (o OllamaFilterer) Filter(m APMessage) (filter bool, reason string, err error) {
	if reflect.DeepEqual(o, OllamaFilterer{}) {
		return false, "", nil
	}
	ms := GetAPMessageCommonFieldAsString(m, "message_text")
	log.Debug(Content("submitting message ending in \""),
		Note(Last20Characters(ms)),
		Content("\" for filtering with Ollama"))
	if o.Model == "" || o.UserPrompt == "" {
		log.Warn(Attention("OllamaFilter model and prompt are required to use the Ollama filter"))
		return false, "", nil
	}
	// If message is blank, return
	if regexp.MustCompile(`^\s*$`).MatchString(ms) {
		log.Debug(Content("message was blank, filtering without calling Ollama"))
		return true, "message blank", nil
	}
	url, err := url.Parse(o.URL)
	if err != nil {
		log.Error(Attention("OllamaFilter url could not be parsed: %s", err))
		return o.FilterOnFailure, "", err
	}
	client := api.NewClient(url, &http.Client{})
	if err != nil {
		log.Error(Attention("error initializing OllamaFilter: %s", err))
		return o.FilterOnFailure, "", err
	}

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
		log.Error(Attention("error setting Ollama response format: %s", err))
		return false, "", err
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
		Prompt:  `Here is the message to evaluate:\n` + ms,
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
			err = fmt.Errorf("%s, Ollama full response: %s", err, resp.Response)
			return err
		}
		ofr := OllamaFilterResult{
			ACARSMessage: ms,
			OllamaFilterRequest: OllamaFilterRequest{
				Model:                               o.Model,
				OllamaSystemPromptFirstInstructions: OllamaFilterFirstInstructions,
				OllamaUserPrompt:                    o.UserPrompt,
				OllamaSystemPromptFinalInstructions: OllamaFilterFinalInstructions,
				ACARSMessage:                        ms,
			},
			OllamaFilterResponse: r,
		}
		db.Create(&ofr)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(OllamaFilterTimeout)*time.Second)
	defer cancel()
	log.Debug(Content("calling OllamaFilter, model "), Note(o.Model))
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
		retry.OnRetry(func(n uint, err error) {
			log.Error(Attention("OllamaFilter attempt #%d failed: %v", n+1, err))
		}),
	)

	action := map[bool]string{
		true:  Custom(*color.New(color.FgCyan), "allow"),
		false: Custom(*color.New(color.FgYellow), "filter"),
	}

	if err != nil {
		log.Error(Attention("too many failures calling OllamaFilter, giving up and "),
			action[!o.FilterOnFailure],
			Attention("ing: %s", err))
		return o.FilterOnFailure, "", err
	}

	log.Info(Content("Ollama decision: "),
		action[r.MessageMatchesCriteria],
		Content(" for message ending in \""),
		Note(Last20Characters(ms)),
		Content("\", reasoning: "),
		Emphasised(r.Reasoning))
	return !r.MessageMatchesCriteria, "final decision", nil
}

func (f OllamaFilterer) Configured() bool {
	return !reflect.DeepEqual(f, OllamaFilterer{})
}
