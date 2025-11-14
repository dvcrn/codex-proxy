package server

import (
	"testing"
)

func TestSupportedModelsIncludesBaseAndSuffixVariants(t *testing.T) {
	models := supportedModels()
	seen := make(map[string]bool, len(models))
	for _, m := range models {
		seen[m.ID] = true
	}

	// Base models
	for _, id := range []string{
		modelGPT5,
		modelGPT5Codex,
		modelGPT51,
		modelGPT51Codex,
		modelGPT5CodexMini,
		modelGPT51CodexMini,
	} {
		if !seen[id] {
			t.Fatalf("expected base model %q to be present in supported models", id)
		}
	}

	// Suffix variants derived from allowed efforts
	expectedVariants := []string{
		"gpt-5-high",
		"gpt-5-minimal",
		"gpt-5.1-high",
		"gpt-5.1-low",
		"gpt-5-codex-high",
		"gpt-5-codex-minimal",
		"gpt-5.1-codex-medium",
		"gpt-5-codex-mini-medium",
		"gpt-5-codex-mini-high",
		"gpt-5.1-codex-mini-medium",
		"gpt-5.1-codex-mini-high",
	}

	for _, id := range expectedVariants {
		if !seen[id] {
			t.Fatalf("expected variant model %q to be present in supported models", id)
		}
	}
}
