package server

import (
	"strings"
	"testing"
)

func TestTransformResponsesRequestBody(t *testing.T) {
	body := map[string]interface{}{
		"instructions": "Please greet Zed.",
		"input": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "input_text",
						"text": "Hello from Zed",
					},
				},
			},
		},
		"reasoning_effort": "none",
	}

	normalizedModel, normalizedEffort := transformResponsesRequestBody(body, "gpt-5-codex-preview", "none")

	if normalizedModel != "gpt-5-codex" {
		t.Fatalf("expected normalized model gpt-5-codex, got %q", normalizedModel)
	}
	if normalizedEffort != "low" {
		t.Fatalf("expected normalized effort low, got %q", normalizedEffort)
	}

	instr, _ := body["instructions"].(string)
	if !containsSubstring(instr, "Please greet Codex.") {
		t.Fatalf("instructions should include replaced text, got %q", instr)
	}

	input := body["input"].([]interface{})
	msg := input[0].(map[string]interface{})
	content := msg["content"].([]interface{})
	item := content[0].(map[string]interface{})
	if item["text"] != "Hello from Codex" {
		t.Fatalf("expected text replacement, got %q", item["text"])
	}

	reasoning, ok := body["reasoning"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected reasoning map to be present")
	}
	if reasoning["effort"] != "low" {
		t.Fatalf("expected reasoning effort low, got %v", reasoning["effort"])
	}

	if store, ok := body["store"].(bool); !ok || store {
		t.Fatalf("expected store to be false, got %v", body["store"])
	}

	include, ok := body["include"].([]interface{})
	if !ok || len(include) == 0 || include[0] != "reasoning.encrypted_content" {
		t.Fatalf("expected include to contain reasoning.encrypted_content, got %v", body["include"])
	}

	if _, exists := body["max_output_tokens"]; exists {
		t.Fatalf("expected max_output_tokens to be removed")
	}
	if _, exists := body["max_tokens"]; exists {
		t.Fatalf("expected max_tokens to be removed")
	}

	if _, exists := body["reasoning_effort"]; exists {
		t.Fatalf("expected reasoning_effort to be removed")
	}

	if body["tool_choice"] != "auto" {
		t.Fatalf("expected tool_choice to default to auto, got %v", body["tool_choice"])
	}
	if body["parallel_tool_calls"] != false {
		t.Fatalf("expected parallel_tool_calls to default to false, got %v", body["parallel_tool_calls"])
	}

	input := body["input"].([]interface{})
	for idx, item := range input {
		if _, ok := item.([]interface{}); ok {
			t.Fatalf("input[%d] should be a message map, found nested array", idx)
		}
	}
}

func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}
