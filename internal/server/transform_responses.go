package server

import "strings"

func transformResponsesRequestBody(body map[string]interface{}, requestedModel string, requestedEffort string) (string, string) {
	normalizedModel := normalizeModel(requestedModel)
	body["model"] = normalizedModel

	// Responses must always disable server-side store per upstream requirements
	body["store"] = false

	// Preserve any user-provided instructions separately and do NOT merge them
	// into the core Codex system prompt. Instead, set the canonical instructions
	// to the codex prefix and append the user's instructions as an additional
	// `input` message so the upstream system prompt remains unchanged.
	var userInstr string
	if existingInstr, ok := body["instructions"].(string); ok {
		userInstr = strings.TrimSpace(existingInstr)
		// remove original to avoid accidental merging downstream
		delete(body, "instructions")
	}

	instructions := codexInstructionsPrefix()
	body["instructions"] = instructions

	overrideInstructions := map[string]interface{}{
		"type": "message",
		"id":   nil,
		"role": "user",
		"content": []interface{}{
			map[string]interface{}{
				"type": "input_text",
				"text": inversePrompt,
			},
		},
	}

	allInstructions := []interface{}{overrideInstructions}

	if existingInput, ok := body["input"]; ok {
		if inSlice, ok := existingInput.([]interface{}); ok {
			allInstructions = append(allInstructions, inSlice...)
		}
	}

	if userInstr != "" {
		repl := replaceNames(userInstr)
		userMsg := map[string]interface{}{
			"type": "message",
			"id":   nil,
			"role": "user",
			"content": []interface{}{
				map[string]interface{}{
					"type": "input_text",
					"text": repl,
				},
			},
		}
		allInstructions = append(allInstructions, userMsg)
	}

	body["input"] = allInstructions

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
	clampedEffort := clampReasoningEffortForModel(normalizedEffort, normalizedModel)
	summary := resolveReasoningSummary(body)
	reasoningSettings := map[string]interface{}{}
	if summary != nil {
		reasoningSettings["summary"] = summary
	}
	if clampedEffort != "" {
		reasoningSettings["effort"] = clampedEffort
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

	return normalizedModel, clampedEffort
}

func sanitizeResponsesInput(body map[string]interface{}) {
	input, ok := body["input"].([]interface{})
	if !ok {
		return
	}
	filtered := input[:0:0]
	for _, msg := range input {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			filtered = append(filtered, msg)
			continue
		}
		if role, _ := msgMap["role"].(string); role == "system" {
			continue
		}
		contents, ok := msgMap["content"].([]interface{})
		if !ok {
			filtered = append(filtered, msg)
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
		filtered = append(filtered, msg)
	}
	body["input"] = filtered
}
