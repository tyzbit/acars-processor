package filter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/avast/retry-go"
	"github.com/fatih/color"
	api "github.com/ollama/ollama/api"
	log "github.com/sirupsen/logrus"
	. "github.com/tyzbit/acars-processor/config"
	. "github.com/tyzbit/acars-processor/database"
	. "github.com/tyzbit/acars-processor/decorate"
	"github.com/tyzbit/acars-processor/util"
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
func OllamaFilter(m string) bool {
	log.Debug(Content("submitting message ending in \""),
		Note(util.Last20Characters(m)),
		Content("\" for filtering with Ollama"))
	// If message is blank, return
	if regexp.MustCompile(`^\s*$`).MatchString(m) {
		log.Debug(Content("message was blank, filtering without calling Ollama"))
		return false
	}
	if Config.Filters.Ollama.Model == "" || Config.Filters.Ollama.UserPrompt == "" {
		log.Warn(Attention("OllamaFilter model and prompt are required to use the Ollama filter"))
		return true
	}
	url, err := url.Parse(Config.Filters.Ollama.URL)
	if err != nil {
		log.Error(Attention("OllamaFilter url could not be parsed: %s", err))
		return true
	}
	client := api.NewClient(url, &http.Client{})
	if err != nil {
		log.Error(Attention("error initializing OllamaFilter: %s", err))
		return true
	}

	if Config.Filters.Ollama.SystemPrompt != "" {
		OllamaFilterFirstInstructions = Config.Filters.Ollama.SystemPrompt
	}
	if Config.Filters.Ollama.Timeout != 0 {
		OllamaFilterTimeout = Config.Filters.Ollama.Timeout
	}
	if Config.Filters.Ollama.MaxRetryAttempts != 0 {
		OllamaFilterMaxRetryAttempts = Config.Filters.Ollama.MaxRetryAttempts
	}
	if Config.Filters.Ollama.MaxRetryDelaySeconds != 0 {
		OllamaFilterRetryDelaySeconds = Config.Filters.Ollama.MaxRetryDelaySeconds
	}

	stream := false
	requestedFormatJson, err := json.Marshal(OllamaFilterResponseRequestedFormat)
	if err != nil {
		log.Error(Attention("error setting Ollama response format: %s", err))
		return true
	}

	opts := map[string]any{}
	for _, opt := range Config.Filters.Ollama.Options {
		opts[opt.Name] = opt.Value
	}

	systemPrompt := OllamaFilterFirstInstructions + Config.Filters.Ollama.UserPrompt +
		OllamaFilterFinalInstructions
	req := &api.GenerateRequest{
		Model:   Config.Filters.Ollama.Model,
		Format:  requestedFormatJson,
		System:  systemPrompt,
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

		content = util.SanitizeJSONString(content)
		err = json.Unmarshal([]byte(content), &r)
		if err != nil {
			err = fmt.Errorf("%s, Ollama full response: %s", err, resp.Response)
			return err
		}
		ofr := OllamaFilterResult{
			ACARSMessage: m,
			OllamaFilterRequest: OllamaFilterRequest{
				Model:                               Config.Filters.Ollama.Model,
				OllamaSystemPromptFirstInstructions: OllamaFilterFirstInstructions,
				OllamaUserPrompt:                    Config.Filters.Ollama.UserPrompt,
				OllamaSystemPromptFinalInstructions: OllamaFilterFinalInstructions,
				ACARSMessage:                        m,
			},
			OllamaFilterResponse: r,
		}
		DB.Create(&ofr)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(OllamaFilterTimeout)*time.Second)
	defer cancel()
	log.Debug(Content("calling OllamaFilter, model "), Note(Config.Filters.Ollama.Model))
	err = retry.Do(func() error {
		err = client.Generate(ctx, req, respFunc)
		if err != nil {
			return &util.RetriableError{
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
			action[!Config.Filters.Ollama.FilterOnFailure],
			Attention("ing: %s", err))
		return !Config.Filters.Ollama.FilterOnFailure
	}

	log.Info(Content("Ollama decision: "),
		action[r.MessageMatchesCriteria],
		Content(" for message ending in \""),
		Note(util.Last20Characters(m)),
		Content("\", reasoning: "),
		Emphasised(r.Reasoning))
	return r.MessageMatchesCriteria
}
