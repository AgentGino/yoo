package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCreatesDefaultConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg, created, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !created {
		t.Fatal("expected config to be created")
	}
	if cfg.Defaults.Model != DefaultModel {
		t.Fatalf("expected default model %q, got %q", DefaultModel, cfg.Defaults.Model)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("expected owner-only config permissions 0600, got %o", got)
	}
}

func TestValidateRejectsBadDefaultMode(t *testing.T) {
	cfg := Default()
	cfg.Defaults.Mode = "missing"

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for missing default mode")
	}
}

func TestAPIKeyUsesConfiguredEnvironmentVariable(t *testing.T) {
	t.Setenv("YO_TEST_OPENROUTER_KEY", " sk-test ")
	cfg := Default()
	cfg.OpenRouter.APIKeyEnv = "YO_TEST_OPENROUTER_KEY"

	if got := cfg.APIKey(); got != "sk-test" {
		t.Fatalf("expected trimmed API key, got %q", got)
	}
}
