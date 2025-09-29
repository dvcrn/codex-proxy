package server

import "strings"

func transformResponsesRequestBody(body map[string]interface{}, requestedModel string, requestedEffort string) (string, string) {
	normalizedModel := normalizeModel(requestedModel)
	body["model"] = normalizedModel

	// Responses must always disable server-side store per upstream requirements
	body["store"] = false

	instructions := codexInstructionsPrefix()
	if existingInstr, ok := body["instructions"].(string); ok {
		trimmed := strings.TrimSpace(existingInstr)
		if trimmed != "" {
			instructions = instructions + "\n\n" + replaceNames(trimmed)
		}
	}
	body["instructions"] = instructions

	sanitizeResponsesInput(body)

	// Always request reasoning encrypted content to match Codex expectations
	body["include"] = []interface{}{"reasoning.encrypted_content"}

	// Ensure tool choice and parallel tool calls defaults
	if _, ok := body["tool_choice"]; !ok {
		body["tool_choice"] = "auto"
	}
	if _, ok := body["parallel_tool_calls"]; !ok {
		body["parallel_tool_calls"] = false
	}

	// Re-map max_output_tokens to max_tokens for upstream compatibility
	if _, ok := body["max_output_tokens"]; ok {
		delete(body, "max_output_tokens")
	}
	if _, ok := body["max_tokens"]; ok {
		delete(body, "max_tokens")
	}

	normalizedEffort := normalizeReasoningEffort(requestedEffort)
	summary := resolveReasoningSummary(body)
	reasoningSettings := map[string]interface{}{}
	if summary != nil {
		reasoningSettings["summary"] = summary
	}
	if normalizedEffort != "" {
		reasoningSettings["effort"] = normalizedEffort
	}
	if len(reasoningSettings) > 0 {
		body["reasoning"] = reasoningSettings
	} else {
		delete(body, "reasoning")
	}

	delete(body, "reasoning_effort")

	if _, ok := body["prompt_cache_key"].(string); !ok {
		instructions, _ := body["instructions"].(string)
		firstText := extractFirstUserText(body)
		if key := derivePromptCacheKey(normalizedModel, instructions, firstText); key != "" {
			body["prompt_cache_key"] = key
		}
	}

	return normalizedModel, normalizedEffort
}

func sanitizeResponsesInput(body map[string]interface{}) {
	input, ok := body["input"].([]interface{})
	if !ok {
		return
	}
	for _, msg := range input {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}
		contents, ok := msgMap["content"].([]interface{})
		if !ok {
			continue
		}
		for _, item := range contents {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if text, ok := itemMap["text"].(string); ok && text != "" {
				itemMap["text"] = replaceNames(text)
			}
		}
	}
}
