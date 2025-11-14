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
	if instr == "" || containsSubstring(instr, "Please greet Codex.") {
		t.Fatalf("instructions should be canonical prefix and not include user text, got %q", instr)
	}

	input := body["input"].([]interface{})
	found := false
	for _, it := range input {
		msg, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		content, ok := msg["content"].([]interface{})
		if !ok || len(content) == 0 {
			continue
		}
		if item, ok := content[0].(map[string]interface{}); ok {
			if txt, _ := item["text"].(string); txt == "Hello from Codex" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("expected to find replaced user text 'Hello from Codex' in input messages; got %v", body["input"])
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

	inputMessages := body["input"].([]interface{})
	for idx, item := range inputMessages {
		if _, ok := item.([]interface{}); ok {
			t.Fatalf("input[%d] should be a message map, found nested array", idx)
		}
	}
}

func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestTransformResponsesRequestBody_ModelSpecificReasoningClamp(t *testing.T) {
	// Mini codex should clamp low effort to medium and default to medium when
	// no explicit effort is provided.
	baseBody := func() map[string]interface{} {
		return map[string]interface{}{
			"instructions": "Do something.",
			"input": []interface{}{
				map[string]interface{}{
					"role": "user",
					"content": []interface{}{
						map[string]interface{}{
							"type": "input_text",
							"text": "Hello",
						},
					},
				},
			},
		}
	}

	// Case 1: explicit low effort gets clamped to medium
	body1 := baseBody()
	requestedEffort1 := "low"
	nModel1, nEffort1 := transformResponsesRequestBody(body1, "gpt-5.1-codex-mini", requestedEffort1)
	if nModel1 != "gpt-5.1-codex-mini" {
		t.Fatalf("expected normalized model gpt-5.1-codex-mini, got %q", nModel1)
	}
	if nEffort1 != "medium" {
		t.Fatalf("expected normalized effort medium, got %q", nEffort1)
	}
	reasoning1, ok := body1["reasoning"].(map[string]interface{})
	if !ok || reasoning1["effort"] != "medium" {
		t.Fatalf("expected reasoning effort medium in body, got %v", body1["reasoning"])
	}

	// Case 2: no effort provided defaults to model-specific default (medium)
	body2 := baseBody()
	nModel2, nEffort2 := transformResponsesRequestBody(body2, "gpt-5.1-codex-mini", "")
	if nModel2 != "gpt-5.1-codex-mini" {
		t.Fatalf("expected normalized model gpt-5.1-codex-mini, got %q", nModel2)
	}
	if nEffort2 != "medium" {
		t.Fatalf("expected normalized effort medium, got %q", nEffort2)
	}
	reasoning2, ok := body2["reasoning"].(map[string]interface{})
	if !ok || reasoning2["effort"] != "medium" {
		t.Fatalf("expected reasoning effort medium in body, got %v", body2["reasoning"])
	}
}
