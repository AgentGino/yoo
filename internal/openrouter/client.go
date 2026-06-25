// Package openrouter wraps OpenRouter's OpenAI-compatible chat completion API.
package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const chatCompletionsPath = "/chat/completions"

// Message is one chat-completions message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest contains the model request fields Yo needs.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

// Client calls OpenRouter with explicit timeout and attribution headers.
type Client struct {
	apiKey      string
	baseURL     string
	httpReferer string
	xTitle      string
	httpClient  *http.Client
}

// Options configures a Client.
type Options struct {
	APIKey      string
	BaseURL     string
	HTTPReferer string
	XTitle      string
	HTTPClient  *http.Client
}

// NewClient validates options and returns a reusable OpenRouter client.
func NewClient(opts Options) (*Client, error) {
	if strings.TrimSpace(opts.APIKey) == "" {
		return nil, errors.New("OPENROUTER_API_KEY is not set")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(opts.BaseURL), "/")
	if baseURL == "" {
		return nil, errors.New("OpenRouter base URL is required")
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	return &Client{
		apiKey:      strings.TrimSpace(opts.APIKey),
		baseURL:     baseURL,
		httpReferer: strings.TrimSpace(opts.HTTPReferer),
		xTitle:      strings.TrimSpace(opts.XTitle),
		httpClient:  httpClient,
	}, nil
}

// Chat sends a chat-completions request and returns choices[0].message.content.
func (c *Client) Chat(ctx context.Context, req ChatRequest) (string, error) {
	if strings.TrimSpace(req.Model) == "" {
		return "", errors.New("model is required")
	}
	if len(req.Messages) == 0 {
		return "", errors.New("at least one message is required")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+chatCompletionsPath, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	if c.httpReferer != "" {
		httpReq.Header.Set("HTTP-Referer", c.httpReferer)
	}
	if c.xTitle != "" {
		httpReq.Header.Set("X-Title", c.xTitle)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("call OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read OpenRouter response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("OpenRouter returned %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}

	var decoded chatResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return "", fmt.Errorf("decode OpenRouter response: %w", err)
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", errors.New("OpenRouter response did not include choices[0].message.content")
	}
	return strings.TrimSpace(decoded.Choices[0].Message.Content), nil
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}
