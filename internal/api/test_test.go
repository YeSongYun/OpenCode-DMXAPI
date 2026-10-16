package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestTesterSupportsAllAPIFormats(t *testing.T) {
	var (
		mu       sync.Mutex
		lastPath string
		lastHead http.Header
		lastBody map[string]any
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make(map[string]any)
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		lastPath = r.URL.Path
		lastHead = r.Header.Clone()
		lastBody = body
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/messages"):
			_, _ = w.Write([]byte("{\"content\":[{\"type\":\"text\",\"text\":\"ok\"}]}"))
		case strings.HasSuffix(r.URL.Path, ":generateContent"):
			_, _ = w.Write([]byte("{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"ok\"}]}}]}"))
		case strings.HasSuffix(r.URL.Path, "/responses"):
			_, _ = w.Write([]byte("{\"output\":[{\"type\":\"message\"}]}"))
		case strings.HasSuffix(r.URL.Path, "/chat/completions"):
			_, _ = w.Write([]byte("{\"choices\":[{\"index\":0,\"message\":{\"role\":\"assistant\",\"content\":\"ok\"}}]}"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tests := []struct {
		name          string
		model         string
		wantPath      string
		wantAPIHeader string
	}{
		{name: "anthropic", model: "claude-sonnet-4", wantPath: "/v1/messages", wantAPIHeader: "x-api-key"},
		{name: "google", model: "gemini-2.5-pro", wantPath: "/v1beta/models/gemini-2.5-pro:generateContent", wantAPIHeader: "x-goog-api-key"},
		{name: "responses", model: "gpt-5-mini", wantPath: "/v1/responses"},
		{name: "chat", model: "deepseek-v4", wantPath: "/v1/chat/completions"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := NewTester(server.URL, "sk-test-key").TestConnection(tc.model)
			if err != nil {
				t.Fatalf("TestConnection() error = %v", err)
			}

			mu.Lock()
			path := lastPath
			headers := lastHead
			body := lastBody
			mu.Unlock()

			if path != tc.wantPath {
				t.Fatalf("request path = %q, want %q", path, tc.wantPath)
			}
			if headers.Get("Authorization") != "Bearer sk-test-key" {
				t.Fatalf("Authorization = %q", headers.Get("Authorization"))
			}
			if tc.wantAPIHeader != "" && headers.Get(tc.wantAPIHeader) != "sk-test-key" {
				t.Fatalf("%s = %q", tc.wantAPIHeader, headers.Get(tc.wantAPIHeader))
			}
			if body["model"] != tc.model && tc.name != "google" {
				t.Fatalf("request model = %v, want %q", body["model"], tc.model)
			}
			if tc.name == "google" && body["contents"] == nil {
				t.Fatalf("Google request missing contents: %#v", body)
			}
		})
	}
}

func TestTesterUsesCustomVersionedBaseURL(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte("{\"choices\":[{\"index\":0,\"message\":{\"role\":\"assistant\",\"content\":\"ok\"}}]}"))
	}))
	defer server.Close()

	if err := NewTester(server.URL+"/v1/chat/completions", "sk-test-key").TestConnection("qwen3-max"); err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("request path = %q, want /v1/chat/completions", gotPath)
	}
}
