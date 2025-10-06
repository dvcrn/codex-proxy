package server

const (
	modelGPT5      = "gpt-5"
	modelGPT5Codex = "gpt-5-codex"
)

// modelMetadata mirrors the JSON schema required by the OpenAI-compatible
// /v1/models endpoint. The structure intentionally keeps nested fields as
// generic maps to simplify aligning with the upstream payload shape and to
// avoid bespoke types for every nested object.
type modelMetadata struct {
	Capabilities        map[string]interface{} `json:"capabilities"`
	ID                  string                 `json:"id"`
	ModelPickerCategory string                 `json:"model_picker_category,omitempty"`
	ModelPickerEnabled  bool                   `json:"model_picker_enabled"`
	Name                string                 `json:"name"`
	Object              string                 `json:"object"`
	Policy              *modelPolicy           `json:"policy,omitempty"`
	Preview             bool                   `json:"preview"`
	SupportedEndpoints  []string               `json:"supported_endpoints,omitempty"`
	Vendor              string                 `json:"vendor"`
	Version             string                 `json:"version"`
}

type modelPolicy struct {
	State string `json:"state"`
	Terms string `json:"terms"`
}

type modelsResponse struct {
	Object string          `json:"object"`
	Data   []modelMetadata `json:"data"`
}

var modelMetadataByID = map[string]modelMetadata{
	modelGPT5: {
		Capabilities: map[string]interface{}{
			"family": "gpt-5",
			"limits": map[string]interface{}{
				"max_context_window_tokens": 264000,
				"max_output_tokens":         64000,
				"max_prompt_tokens":         128000,
				"vision": map[string]interface{}{
					"max_prompt_image_size": 3145728,
					"max_prompt_images":     1,
					"supported_media_types": []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
				},
			},
			"object":    "model_capabilities",
			"supports":  map[string]interface{}{"parallel_tool_calls": true, "streaming": true, "structured_outputs": true, "tool_calls": true, "vision": true},
			"tokenizer": "o200k_base",
			"type":      "chat",
		},
		ID:                  modelGPT5,
		ModelPickerCategory: "versatile",
		ModelPickerEnabled:  true,
		Name:                "GPT-5",
		Object:              "model",
		Policy: &modelPolicy{
			State: "enabled",
			Terms: "Enable access to the latest GPT-5 model from OpenAI. [Learn more about how GitHub Copilot serves GPT-5](https://gh.io/copilot-openai).",
		},
		Preview: false,
		Vendor:  "Azure OpenAI",
		Version: "gpt-5",
	},
	modelGPT5Codex: {
		Capabilities: map[string]interface{}{
			"family": "gpt-5-codex",
			"limits": map[string]interface{}{
				"max_context_window_tokens": 200000,
				"max_output_tokens":         64000,
				"max_prompt_tokens":         128000,
				"vision": map[string]interface{}{
					"max_prompt_image_size": 3145728,
					"max_prompt_images":     1,
					"supported_media_types": []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
				},
			},
			"object":    "model_capabilities",
			"supports":  map[string]interface{}{"parallel_tool_calls": true, "streaming": true, "structured_outputs": true, "tool_calls": true, "vision": true},
			"tokenizer": "o200k_base",
			"type":      "chat",
		},
		ID:                  modelGPT5Codex,
		ModelPickerCategory: "powerful",
		ModelPickerEnabled:  true,
		Name:                "GPT-5-Codex (Preview)",
		Object:              "model",
		Policy: &modelPolicy{
			State: "enabled",
			Terms: "Enable access to the latest GPT-5-Codex model from OpenAI. [Learn more about how GitHub Copilot serves GPT-5-Codex](https://gh.io/copilot-openai).",
		},
		Preview:            true,
		SupportedEndpoints: []string{"/responses"},
		Vendor:             "OpenAI",
		Version:            "gpt-5-codex",
	},
}

var supportedModelIDs = []string{modelGPT5, modelGPT5Codex}

func supportedModels() []modelMetadata {
	models := make([]modelMetadata, 0, len(supportedModelIDs))
	for _, id := range supportedModelIDs {
		if m, ok := modelMetadataByID[id]; ok {
			models = append(models, m)
		}
	}
	return models
}
