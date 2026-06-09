package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mcuadros/go-defaults"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// TestOpenAIFiltererSendsSingleSystemMessage verifies that OpenAIFilterer.Filter
// sends exactly one leading system message followed by a single user message.
//
// Regression test for https://github.com/tyzbit/acars-processor/issues/45:
// Strict chat templates (e.g. Qwen3 served by llama.cpp --jinja) reject
// requests with multiple consecutive system messages and return HTTP 400,
// which (combined with the default FilterOnFailure=true) causes every
// message to be silently dropped.
func TestOpenAIFiltererSendsSingleSystemMessage(t *testing.T) {
	var (
		gotPath             string
		gotMethod           string
		gotMessages         []map[string]any
		gotContentType      string
		gotAuthorizationHdr string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotAuthorizationHdr = r.Header.Get("Authorization")

		buf, _ := io.ReadAll(r.Body)
		var body struct {
			Messages []map[string]any `json:"messages"`
			Model    string           `json:"model"`
		}
		if err := json.Unmarshal(buf, &body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		gotMessages = body.Messages

		// Reply with a valid chat.completions response so Filter() can parse it.
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"created": 0,
			"model":   body.Model,
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": `{"message_matches_criteria":"false","reasoning":"test"}`,
					},
					"finish_reason": "stop",
				},
			},
		})
	}))
	defer srv.Close()

	f := OpenAIFilterer{
		FilterOnFailure: false,
		APIKey:          "test-key",
		Model:           "gpt-4o",
		UserPrompt:      "USER-PROMPT-MARKER",
		SystemPrompt:    "SYS-MARKER",
		URL:             srv.URL + "/v1/",
	}

	filterThis, _, err := f.Filter(APMessage{
		ACARSProcessorPrefix + "MessageText": "hello world",
	})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}
	// Mock returns message_matches_criteria="false" (no match) -> the filter
	// should drop the message, i.e. filterThisMessage == true.
	if !filterThis {
		t.Fatalf("expected filterThis=true (model reported no match), got false")
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected POST request, got %s", gotMethod)
	}
	if gotPath != "/v1/chat/completions" {
		t.Errorf("expected path /v1/chat/completions, got %s", gotPath)
	}
	if !strings.HasPrefix(gotContentType, "application/json") {
		t.Errorf("expected JSON content type, got %q", gotContentType)
	}
	if gotAuthorizationHdr != "Bearer test-key" {
		t.Errorf("expected Bearer auth header, got %q", gotAuthorizationHdr)
	}

	// Core regression check: exactly one system message and one user message.
	if len(gotMessages) != 2 {
		t.Fatalf("expected exactly 2 messages (1 system + 1 user), got %d: %#v", len(gotMessages), gotMessages)
	}
	if role, _ := gotMessages[0]["role"].(string); role != "system" {
		t.Errorf("expected first message role=system, got %q", role)
	}
	if role, _ := gotMessages[1]["role"].(string); role != "user" {
		t.Errorf("expected second message role=user, got %q", role)
	}

	// The single system message must contain the framing prompt, the user
	// prompt, and the closing instructions, in that order. The openai-go
	// library serializes content as either a plain string or an array of
	// {type,text} parts; we accept either form and concatenate any parts.
	sysContent := extractMessageContent(t, gotMessages[0])
	idxMarker := strings.Index(sysContent, "USER-PROMPT-MARKER")
	idxSys := strings.Index(sysContent, "SYS-MARKER")
	idxFinal := strings.Index(sysContent, "message_matches_criteria")
	if idxMarker < 0 {
		t.Errorf("user prompt not found in system message; got: %q", sysContent)
	}
	if idxSys < 0 {
		t.Errorf("default system prompt not found in system message; got: %q", sysContent)
	}
	if idxFinal < 0 {
		t.Errorf("final instructions not found in system message; got: %q", sysContent)
	}
	if !(idxSys < idxMarker && idxMarker < idxFinal) {
		t.Errorf("expected order: framing < user-prompt < final; got sys=%d marker=%d final=%d in %q",
			idxSys, idxMarker, idxFinal, sysContent)
	}

	// Ensure the user message contains the original message text, unchanged.
	userContent := extractMessageContent(t, gotMessages[1])
	if userContent != "hello world" {
		t.Errorf("expected user message content %q, got %q", "hello world", userContent)
	}
}

// TestOpenAIFiltererFilterOnFailureDefaultsTrue documents the existing default
// behavior called out in the issue: when the backend returns an error,
// Filter() returns FilterOnFailure, which defaults to true, causing the
// message to be dropped. The fix in this branch only changes the request
// shape, not the failure-handling policy.
func TestOpenAIFiltererFilterOnFailureDefaultsTrue(t *testing.T) {
	// Struct-tag defaults are applied at config-load time via the
	// mcuadros/go-defaults library. Verify the loaded default matches the
	// tag, so a user who simply instantiates OpenAIFilterer{} at runtime
	// (without loading config) gets the documented default-on-failure
	// behavior.
	f := OpenAIFilterer{}
	defaults.SetDefaults(&f)
	if !f.FilterOnFailure {
		t.Fatalf("expected FilterOnFailure default to be true (per struct tag); got false")
	}
}

// TestOpenAIFiltererDropsMessageOnBackendError captures the silent-drop
// failure mode described in the issue: a backend error combined with
// FilterOnFailure=true causes the message to be filtered out. The
// implementation does surface the error to the caller (so the step
// framework can log it), but the message is dropped from the pipeline
// because FilterOnFailure=true. This test pins that behavior so we
// notice if it ever changes unintentionally.
func TestOpenAIFiltererDropsMessageOnBackendError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate the exact failure mode in the issue: a strict-template
		// rejection (400 "System message must be at the beginning").
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":{"code":400,"message":"System message must be at the beginning.","type":"invalid_request_error"}}`)
	}))
	defer srv.Close()

	f := OpenAIFilterer{
		FilterOnFailure: true,
		APIKey:          "k",
		Model:           "gpt-4o",
		UserPrompt:      "x",
		URL:             srv.URL + "/v1/",
	}
	filterThis, _, err := f.Filter(APMessage{ACARSProcessorPrefix + "MessageText": "hi"})
	if err == nil {
		t.Fatalf("expected Filter to surface the backend error to the caller, got nil")
	}
	if !filterThis {
		t.Fatalf("expected message to be filtered out on backend error with FilterOnFailure=true, but it passed through")
	}
}

// Sanity check: ensure the openai-go types we depend on still expose the
// constructors we use. This catches accidental import or version drift.
func TestOpenAIFiltererOpenAIHelpersAvailable(t *testing.T) {
	_ = openai.SystemMessage("x")
	_ = openai.UserMessage("y")
	_ = openai.NewClient(option.WithAPIKey("k"))
	_ = context.TODO()
}

// extractMessageContent returns the textual content of a message as decoded
// from the JSON body. openai-go may serialize content either as a plain
// string or as an array of typed parts (e.g. {"type":"text","text":"..."}).
func extractMessageContent(t *testing.T, msg map[string]any) string {
	t.Helper()
	raw, ok := msg["content"]
	if !ok {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return v
	case []any:
		var sb strings.Builder
		for _, part := range v {
			m, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := m["text"].(string); ok {
				sb.WriteString(text)
			}
		}
		return sb.String()
	default:
		return ""
	}
}
