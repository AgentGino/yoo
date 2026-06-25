// Package cli contains Yoo's command-line interface and orchestration.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/agentgino/yoo/internal/config"
	"github.com/agentgino/yoo/internal/openrouter"
)

// Runner executes the CLI with injectable streams for tests.
type Runner struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Version string
}

type options struct {
	configPath  string
	model       string
	mode        string
	temperature float64
	listModels  bool
	showConfig  bool
	showVersion bool
}

// Run parses args, loads config, and calls OpenRouter when a prompt is provided.
func (r Runner) Run(ctx context.Context, args []string) error {
	stdout := r.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := r.Stderr
	if stderr == nil {
		stderr = io.Discard
	}

	opts, remaining, err := parseArgs(args, stderr)
	if err != nil {
		return err
	}
	if opts.showVersion {
		fmt.Fprintf(stdout, "yoo %s\n", r.Version)
		return nil
	}

	cfg, created, err := config.Load(opts.configPath)
	if err != nil {
		return err
	}
	if created {
		path := opts.configPath
		if path == "" {
			path, _ = config.ConfigPath()
		}
		fmt.Fprintf(stderr, "created config: %s\n", path)
	}
	if opts.showConfig {
		path := opts.configPath
		if path == "" {
			path, _ = config.ConfigPath()
		}
		fmt.Fprintln(stdout, path)
		return nil
	}
	if opts.listModels {
		for _, model := range cfg.Models {
			fmt.Fprintln(stdout, model)
		}
		return nil
	}

	mode, promptText, err := resolveModeAndPrompt(cfg, opts.mode, remaining)
	if err != nil {
		printUsage(stderr, cfg)
		return err
	}
	prompt := cfg.Prompts[mode]
	model := firstNonEmpty(opts.model, prompt.Model, cfg.Defaults.Model)
	temperature := cfg.Defaults.Temperature
	if opts.temperature >= 0 {
		temperature = opts.temperature
	}

	client, err := openrouter.NewClient(openrouter.Options{
		APIKey:      cfg.APIKey(),
		BaseURL:     cfg.OpenRouter.BaseURL,
		HTTPReferer: cfg.OpenRouter.HTTPReferer,
		XTitle:      cfg.OpenRouter.XTitle,
	})
	if err != nil {
		return err
	}
	content, err := client.Chat(ctx, openrouter.ChatRequest{
		Model:       model,
		Temperature: temperature,
		Messages: []openrouter.Message{
			{Role: "system", Content: prompt.System},
			{Role: "user", Content: promptText},
		},
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, content)
	return nil
}

func parseArgs(args []string, stderr io.Writer) (options, []string, error) {
	opts := options{temperature: -1}
	fs := flag.NewFlagSet("yoo", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&opts.configPath, "config", "", "config path (default: $XDG_CONFIG_HOME/yoo/config.json)")
	fs.StringVar(&opts.model, "model", "", "OpenRouter model id, e.g. anthropic/claude-3.5-sonnet")
	fs.StringVar(&opts.model, "m", "", "short for -model")
	fs.StringVar(&opts.mode, "mode", "", "prompt mode from config")
	fs.StringVar(&opts.mode, "p", "", "short for -mode")
	fs.Float64Var(&opts.temperature, "temperature", -1, "sampling temperature 0..2")
	fs.BoolVar(&opts.listModels, "list-models", false, "list configured model shortcuts")
	fs.BoolVar(&opts.showConfig, "show-config", false, "print config path")
	fs.BoolVar(&opts.showVersion, "version", false, "print version")
	if err := fs.Parse(args); err != nil {
		return options{}, nil, err
	}
	return opts, fs.Args(), nil
}

func resolveModeAndPrompt(cfg config.Config, explicitMode string, args []string) (string, string, error) {
	mode := firstNonEmpty(explicitMode, cfg.Defaults.Mode)
	if strings.TrimSpace(mode) == "" {
		return "", "", errors.New("mode is required")
	}
	if len(args) > 0 {
		if _, ok := cfg.Prompts[args[0]]; ok && explicitMode == "" {
			mode = args[0]
			args = args[1:]
		}
	}
	if _, ok := cfg.Prompts[mode]; !ok {
		return "", "", fmt.Errorf("unknown mode %q", mode)
	}
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if prompt == "" {
		return "", "", errors.New("prompt is required")
	}
	return mode, prompt, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func printUsage(w io.Writer, cfg config.Config) {
	fmt.Fprintln(w, "Usage: yoo [flags] [mode] <prompt>")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  yoo shell list files changed today")
	fmt.Fprintln(w, "  yoo -m anthropic/claude-3.5-sonnet code write a Go HTTP client")
	fmt.Fprintf(w, "Modes: %s\n", strings.Join(cfg.ModeNames(), ", "))
}
