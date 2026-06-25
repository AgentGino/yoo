// Package config owns Yoo's on-disk configuration schema and validation.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// DefaultBaseURL is OpenRouter's OpenAI-compatible chat completions endpoint.
	DefaultBaseURL = "https://openrouter.ai/api/v1"
	// DefaultModel is cheap, fast, and broadly available on OpenRouter.
	DefaultModel = "openai/gpt-4o-mini"
)

// Config is the stable user-editable configuration stored as JSON.
type Config struct {
	OpenRouter OpenRouterConfig  `json:"openrouter"`
	Defaults   Defaults          `json:"defaults"`
	Prompts    map[string]Prompt `json:"prompts"`
	Models     []string          `json:"models"`
}

// OpenRouterConfig contains provider-level OpenRouter settings.
type OpenRouterConfig struct {
	APIKeyEnv   string `json:"api_key_env"`
	BaseURL     string `json:"base_url"`
	HTTPReferer string `json:"http_referer"`
	XTitle      string `json:"x_title"`
}

// Defaults contains command defaults that can be overridden with flags.
type Defaults struct {
	Model       string  `json:"model"`
	Mode        string  `json:"mode"`
	Temperature float64 `json:"temperature"`
}

// Prompt defines a named system prompt and optional model override.
type Prompt struct {
	System string `json:"system"`
	Model  string `json:"model,omitempty"`
}

// APIKey resolves the configured OpenRouter API key environment variable.
func (c Config) APIKey() string {
	name := strings.TrimSpace(c.OpenRouter.APIKeyEnv)
	if name == "" {
		name = "OPENROUTER_API_KEY"
	}
	return strings.TrimSpace(os.Getenv(name))
}

// ModeNames returns prompt mode names in stable order for usage text.
func (c Config) ModeNames() []string {
	names := make([]string, 0, len(c.Prompts))
	for name := range c.Prompts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ConfigPath returns the XDG-style Yoo config path.
func ConfigPath() (string, error) {
	if override := strings.TrimSpace(os.Getenv("YOO_CONFIG")); override != "" {
		return override, nil
	}
	base := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "yoo", "config.json"), nil
}

// Default returns a complete config suitable for first run.
func Default() Config {
	return Config{
		OpenRouter: OpenRouterConfig{
			APIKeyEnv:   "OPENROUTER_API_KEY",
			BaseURL:     DefaultBaseURL,
			HTTPReferer: "https://github.com/AgentGino/yoo",
			XTitle:      "yoo",
		},
		Defaults: Defaults{
			Model:       DefaultModel,
			Mode:        "chat",
			Temperature: 0.2,
		},
		Prompts: map[string]Prompt{
			"chat": {
				System: "You are Yoo, a direct command-line assistant. Be concise, useful, and avoid filler.",
			},
			"shell": {
				System: "Return only the safest POSIX shell command that satisfies the request. No markdown, no explanation.",
			},
			"code": {
				System: "You are a senior coding assistant. Return concise, correct code or focused implementation guidance.",
			},
		},
		Models: []string{
			DefaultModel,
			"anthropic/claude-3.5-sonnet",
			"google/gemini-2.0-flash-001",
			"meta-llama/llama-3.1-70b-instruct",
		},
	}
}

// Load reads config from disk, creating a default file if missing.
func Load(path string) (Config, bool, error) {
	if strings.TrimSpace(path) == "" {
		resolved, err := ConfigPath()
		if err != nil {
			return Config{}, false, err
		}
		path = resolved
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		cfg := Default()
		if err := Save(path, cfg); err != nil {
			return Config{}, false, err
		}
		return cfg, true, nil
	} else if err != nil {
		return Config{}, false, fmt.Errorf("stat config: %w", err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		return Config{}, false, fmt.Errorf("read config: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(body, &cfg); err != nil {
		return Config{}, false, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, false, err
	}
	return cfg, false, nil
}

// Save writes a formatted JSON config with owner-only permissions.
func Save(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	body, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	body = append(body, '\n')
	if err := os.WriteFile(path, body, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// Validate rejects incomplete config at the boundary instead of failing mid-request.
func (c Config) Validate() error {
	if strings.TrimSpace(c.OpenRouter.BaseURL) == "" {
		return errors.New("config openrouter.base_url is required")
	}
	if strings.TrimSpace(c.OpenRouter.APIKeyEnv) == "" {
		return errors.New("config openrouter.api_key_env is required")
	}
	if strings.TrimSpace(c.Defaults.Model) == "" {
		return errors.New("config defaults.model is required")
	}
	if strings.TrimSpace(c.Defaults.Mode) == "" {
		return errors.New("config defaults.mode is required")
	}
	if c.Defaults.Temperature < 0 || c.Defaults.Temperature > 2 {
		return errors.New("config defaults.temperature must be between 0 and 2")
	}
	if len(c.Prompts) == 0 {
		return errors.New("config prompts must define at least one mode")
	}
	for name, prompt := range c.Prompts {
		if strings.TrimSpace(name) == "" {
			return errors.New("config prompts cannot contain an empty mode name")
		}
		if strings.TrimSpace(prompt.System) == "" {
			return fmt.Errorf("config prompt %q system is required", name)
		}
	}
	if _, ok := c.Prompts[c.Defaults.Mode]; !ok {
		return fmt.Errorf("config defaults.mode %q is not defined in prompts", c.Defaults.Mode)
	}
	return nil
}
