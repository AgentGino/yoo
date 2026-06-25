package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChatSendsOpenRouterHeadersAndReturnsContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != chatCompletionsPath {
			t.Fatalf("expected path %s, got %s", chatCompletionsPath, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header %q", got)
		}
		if got := r.Header.Get("HTTP-Referer"); got != "https://example.test" {
			t.Fatalf("unexpected referer %q", got)
		}
		var body ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body.Model != "openai/gpt-4o-mini" {
			t.Fatalf("unexpected model %q", body.Model)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"hello"}}]}`))
	}))
	defer server.Close()

	client, err := NewClient(Options{
		APIKey:      "test-key",
		BaseURL:     server.URL,
		HTTPReferer: "https://example.test",
		XTitle:      "yoo-test",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	got, err := client.Chat(context.Background(), ChatRequest{
		Model:    "openai/gpt-4o-mini",
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("expected hello, got %q", got)
	}
}

func TestChatReturnsHTTPErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad model", http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := NewClient(Options{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.Chat(context.Background(), ChatRequest{
		Model:    "bad/model",
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected HTTP error")
	}
}
