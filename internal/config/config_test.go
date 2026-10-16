package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMatchModel(t *testing.T) {
	tests := []struct {
		name   string
		model  string
		format APIFormat
	}{
		{name: "anthropic prefix", model: "claude-3-5-sonnet", format: APIFormatAnthropic},
		{name: "anthropic namespace", model: "anthropic/claude-3-5-sonnet", format: APIFormatAnthropic},
		{name: "anthropic cc suffix", model: "custom-model-cc", format: APIFormatAnthropic},
		{name: "google prefix", model: "gemini-2.5-pro", format: APIFormatGoogle},
		{name: "google namespace", model: "google/gemini-2.5-pro", format: APIFormatGoogle},
		{name: "responses gpt 5", model: "gpt-5", format: APIFormatOpenAIResponses},
		{name: "chat gpt 4o", model: "gpt-4o", format: APIFormatOpenAIChat},
		{name: "responses namespace", model: "openai/gpt-4.1", format: APIFormatOpenAIResponses},
		{name: "chat namespace overrides model prefix", model: "chat/gpt-5", format: APIFormatOpenAIChat},
		{name: "chat compatible namespace", model: "openai-chat/qwen-plus", format: APIFormatOpenAIChat},
		{name: "chat common model", model: "deepseek-chat", format: APIFormatOpenAIChat},
		{name: "unknown model fallback", model: "vendor/custom-model", format: APIFormatOpenAIChat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchModel(tt.model)
			if got.Format != tt.format {
				t.Fatalf("MatchModel(%q).Format = %q, want %q (reason: %s)", tt.model, got.Format, tt.format, got.Reason)
			}
			if got.Model != tt.model {
				t.Fatalf("MatchModel(%q).Model = %q, want original model ID", tt.model, got.Model)
			}
			if got.Reason == "" {
				t.Fatalf("MatchModel(%q) returned an empty reason", tt.model)
			}
		})
	}

	trimmed := MatchModel("  claude-3-haiku  ")
	if trimmed.Model != "claude-3-haiku" || trimmed.Format != APIFormatAnthropic {
		t.Fatalf("MatchModel should trim input while preserving the model ID: %#v", trimmed)
	}
}

func TestBuildProviderBaseURL(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		format APIFormat
		want   string
	}{
		{name: "chat root", rawURL: "https://api.example.test", format: APIFormatOpenAIChat, want: "https://api.example.test/v1"},
		{name: "chat keeps custom base path", rawURL: "https://api.example.test/api/dmx", format: APIFormatOpenAIChat, want: "https://api.example.test/api/dmx/v1"},
		{name: "chat strips complete endpoint", rawURL: "https://api.example.test/v1/chat/completions", format: APIFormatOpenAIChat, want: "https://api.example.test/v1"},
		{name: "responses strips complete endpoint", rawURL: "https://api.example.test/v1/responses", format: APIFormatOpenAIResponses, want: "https://api.example.test/v1"},
		{name: "anthropic strips complete endpoint", rawURL: "https://api.example.test/v1/messages", format: APIFormatAnthropic, want: "https://api.example.test/v1"},
		{name: "anthropic custom path is not rewritten", rawURL: "https://api.example.test/api/anthropic", format: APIFormatAnthropic, want: "https://api.example.test/api/anthropic"},
		{name: "google root", rawURL: "https://generativelanguage.example.test", format: APIFormatGoogle, want: "https://generativelanguage.example.test/v1beta"},
		{name: "google strips generate content endpoint", rawURL: "https://generativelanguage.example.test/v1beta/models/gemini-2.5-pro:generateContent", format: APIFormatGoogle, want: "https://generativelanguage.example.test/v1beta"},
		{name: "query is retained", rawURL: "https://api.example.test/v1/?tenant=demo", format: APIFormatOpenAIChat, want: "https://api.example.test/v1?tenant=demo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildProviderBaseURL(tt.rawURL, tt.format); got != tt.want {
				t.Fatalf("BuildProviderBaseURL(%q, %q) = %q, want %q", tt.rawURL, tt.format, got, tt.want)
			}
		})
	}
}

func TestBuildAPIEndpoint(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		format APIFormat
		model  string
		want   string
	}{
		{name: "chat root", rawURL: "https://api.example.test", format: APIFormatOpenAIChat, model: "deepseek-chat", want: "https://api.example.test/v1/chat/completions"},
		{name: "chat existing endpoint", rawURL: "https://api.example.test/v1/chat/completions", format: APIFormatOpenAIChat, model: "deepseek-chat", want: "https://api.example.test/v1/chat/completions"},
		{name: "responses existing endpoint", rawURL: "https://api.example.test/v1/responses", format: APIFormatOpenAIResponses, model: "gpt-5", want: "https://api.example.test/v1/responses"},
		{name: "anthropic existing endpoint", rawURL: "https://api.example.test/v1/messages", format: APIFormatAnthropic, model: "claude-3-5-sonnet", want: "https://api.example.test/v1/messages"},
		{name: "google generate content", rawURL: "https://generativelanguage.example.test/v1beta", format: APIFormatGoogle, model: "gemini-2.5-pro", want: "https://generativelanguage.example.test/v1beta/models/gemini-2.5-pro:generateContent"},
		{name: "google escapes model path", rawURL: "https://generativelanguage.example.test", format: APIFormatGoogle, model: "gemini/foo", want: "https://generativelanguage.example.test/v1beta/models/gemini%2Ffoo:generateContent"},
		{name: "query stays after endpoint", rawURL: "https://api.example.test/v1?tenant=demo", format: APIFormatOpenAIChat, model: "deepseek-chat", want: "https://api.example.test/v1/chat/completions?tenant=demo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildAPIEndpoint(tt.rawURL, tt.format, tt.model); got != tt.want {
				t.Fatalf("BuildAPIEndpoint(%q, %q, %q) = %q, want %q", tt.rawURL, tt.format, tt.model, got, tt.want)
			}
		})
	}
}

func TestNewDMXAPIConfigGroupsMixedModels(t *testing.T) {
	models := []string{
		"claude-3-5-sonnet",
		"gemini-2.5-pro",
		"gpt-5",
		"gpt-4o",
		"deepseek-chat",
		"",
		"  ",
	}
	got := NewDMXAPIConfig("https://api.example.test/api", "secret", models)
	if got == nil {
		t.Fatal("NewDMXAPIConfig returned nil")
	}

	wantProviders := map[string]struct {
		format  APIFormat
		models  []string
		baseURL string
	}{
		"dmxapi-anthropic":        {format: APIFormatAnthropic, models: []string{"claude-3-5-sonnet"}, baseURL: "https://api.example.test/api"},
		"dmxapi-google":           {format: APIFormatGoogle, models: []string{"gemini-2.5-pro"}, baseURL: "https://api.example.test/api/v1beta"},
		"dmxapi-openai-responses": {format: APIFormatOpenAIResponses, models: []string{"gpt-5"}, baseURL: "https://api.example.test/api/v1"},
		"dmxapi-openai":           {format: APIFormatOpenAIChat, models: []string{"gpt-4o", "deepseek-chat"}, baseURL: "https://api.example.test/api/v1"},
	}
	if len(got.Provider) != len(wantProviders) {
		t.Fatalf("NewDMXAPIConfig created %d providers, want %d: %#v", len(got.Provider), len(wantProviders), got.Provider)
	}
	for id, want := range wantProviders {
		provider, ok := got.Provider[id]
		if !ok {
			t.Fatalf("NewDMXAPIConfig missing provider %q", id)
		}
		info := GetProviderInfo(want.format)
		if provider.NPM != info.NPM || provider.Name != info.Name {
			t.Fatalf("provider %q metadata = (%q, %q), want (%q, %q)", id, provider.NPM, provider.Name, info.NPM, info.Name)
		}
		if provider.Options.BaseURL != want.baseURL || provider.Options.APIKey != "secret" {
			t.Fatalf("provider %q options = %#v, want baseURL %q and API key", id, provider.Options, want.baseURL)
		}
		if len(provider.Models) != len(want.models) {
			t.Fatalf("provider %q has %d models, want %d: %#v", id, len(provider.Models), len(want.models), provider.Models)
		}
		for _, model := range want.models {
			entry, ok := provider.Models[model]
			if !ok || entry.Name != model {
				t.Fatalf("provider %q model %q = %#v, want model name preserved", id, model, entry)
			}
		}
	}
}

func TestGetRepresentativeModelsGroupsByFormat(t *testing.T) {
	models := []string{
		"deepseek-chat",
		"claude-3-haiku",
		"claude-3-5-sonnet",
		"gemini-2.0-flash",
		"gpt-5",
		"gpt-4o",
		"qwen-plus",
	}
	want := []string{"claude-3-haiku", "gemini-2.0-flash", "gpt-5", "deepseek-chat"}
	got := GetRepresentativeModels(models)
	if len(got) != len(want) {
		t.Fatalf("GetRepresentativeModels returned %d models, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("GetRepresentativeModels[%d] = %q, want %q (all = %#v)", i, got[i], want[i], got)
		}
	}
}

func TestReaderReadsLegacyDMXAPIProvider(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	configDir := filepath.Join(configHome, "opencode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("create config directory: %v", err)
	}
	data := []byte(`{
  "provider": {
    "dmxapi": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "DMXAPI",
      "options": {
        "baseURL": "https://api.example.test/v1/chat/completions",
        "apiKey": "legacy-key"
      },
      "models": {
        "gpt-4o": {"name": "gpt-4o"},
        "deepseek-chat": {"name": "deepseek-chat"}
      }
    },
    "openai": {
      "options": {"baseURL": "https://other.example.test", "apiKey": "ignored"},
      "models": {"other-model": {"name": "other-model"}}
    }
  }
}`)
	if err := os.WriteFile(filepath.Join(configDir, "opencode.json"), data, 0600); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	got := NewReader().ReadExistingConfig()
	if got == nil {
		t.Fatal("ReadExistingConfig returned nil for a legacy dmxapi provider")
	}
	if got.URL != "https://api.example.test" {
		t.Fatalf("ReadExistingConfig URL = %q, want normalized root URL", got.URL)
	}
	if got.APIKey != "legacy-key" {
		t.Fatalf("ReadExistingConfig APIKey = %q, want legacy-key", got.APIKey)
	}
	wantModels := []string{"deepseek-chat", "gpt-4o"}
	if strings.Join(got.Models, "\x00") != strings.Join(wantModels, "\x00") {
		t.Fatalf("ReadExistingConfig Models = %#v, want %#v", got.Models, wantModels)
	}
}

func TestMergeManagedProviderPreservesCustomFields(t *testing.T) {
	oldProvider := map[string]interface{}{
		"name":   "old",
		"custom": map[string]interface{}{"keep": true},
		"models": map[string]interface{}{
			"keep-model": map[string]interface{}{"name": "keep-model", "temperature": 0.2},
			"old-model":  map[string]interface{}{"name": "old-model"},
		},
	}
	newProvider := map[string]interface{}{
		"name": "new",
		"models": map[string]interface{}{
			"keep-model": map[string]interface{}{"name": "keep-model"},
			"new-model":  map[string]interface{}{"name": "new-model"},
		},
	}

	got := mergeManagedProvider(oldProvider, newProvider)
	if got["name"] != "new" {
		t.Fatalf("provider name = %v, want new", got["name"])
	}
	if got["custom"].(map[string]interface{})["keep"] != true {
		t.Fatalf("provider custom field was not preserved: %#v", got["custom"])
	}
	models := got["models"].(map[string]interface{})
	if _, ok := models["old-model"]; ok {
		t.Fatalf("stale model should be removed: %#v", models)
	}
	if models["keep-model"].(map[string]interface{})["temperature"] != 0.2 {
		t.Fatalf("custom model field was not preserved: %#v", models["keep-model"])
	}
}

func TestManagedProviderKeyScope(t *testing.T) {
	for _, key := range []string{"dmxapi", "dmxapi-anthropic", "dmxapi-google", "dmxapi-openai", "dmxapi-openai-responses"} {
		if !isManagedProviderKey(key) {
			t.Fatalf("expected %q to be managed", key)
		}
	}
	if isManagedProviderKey("dmxapi-custom") {
		t.Fatal("custom dmxapi provider should not be removed as managed")
	}
}
