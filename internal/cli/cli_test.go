package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentgino/yoo/internal/config"
)

func TestListModelsDoesNotRequireAPIKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Default()
	cfg.Models = []string{"model/a", "model/b"}
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	var stdout bytes.Buffer
	runner := Runner{Stdout: &stdout, Stderr: &bytes.Buffer{}, Version: "test"}
	if err := runner.Run(context.Background(), []string{"-config", path, "-list-models"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if got := stdout.String(); got != "model/a\nmodel/b\n" {
		t.Fatalf("unexpected stdout %q", got)
	}
}

func TestResolveModeAndPromptSupportsLegacyModePosition(t *testing.T) {
	cfg := config.Default()
	mode, prompt, err := resolveModeAndPrompt(cfg, "", []string{"shell", "list", "files"})
	if err != nil {
		t.Fatalf("resolveModeAndPrompt: %v", err)
	}
	if mode != "shell" || prompt != "list files" {
		t.Fatalf("unexpected mode/prompt %q/%q", mode, prompt)
	}
}

func TestMissingPromptPrintsUsage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := config.Save(path, config.Default()); err != nil {
		t.Fatalf("save config: %v", err)
	}

	var stderr bytes.Buffer
	runner := Runner{Stdout: &bytes.Buffer{}, Stderr: &stderr, Version: "test"}
	err := runner.Run(context.Background(), []string{"-config", path})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr.String(), "Usage: yoo") {
		t.Fatalf("expected usage, got %q", stderr.String())
	}
}
