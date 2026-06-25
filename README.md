# yo

`yo` is a small Go CLI for asking OpenRouter models from the terminal. It is built for quick answers, shell command generation, and model switching without touching code.

## Install

```bash
brew install AgentGino/tools/yo
```

Or install directly with Go:

```bash
go install github.com/agentgino/yo/cmd/yo@latest
```

## Configure

Set your OpenRouter key:

```bash
export OPENROUTER_API_KEY="sk-or-..."
```

First run creates `~/.config/yo/config.json` with default modes and model shortcuts. You can override the path with `YO_CONFIG` or `-config`.

```bash
yo -show-config
yo -list-models
```

Default config shape:

```json
{
  "openrouter": {
    "api_key_env": "OPENROUTER_API_KEY",
    "base_url": "https://openrouter.ai/api/v1",
    "http_referer": "https://github.com/AgentGino/yo",
    "x_title": "yo"
  },
  "defaults": {
    "model": "openai/gpt-4o-mini",
    "mode": "chat",
    "temperature": 0.2
  },
  "prompts": {
    "chat": {
      "system": "You are Yo, a direct command-line assistant. Be concise, useful, and avoid filler."
    },
    "shell": {
      "system": "Return only the safest POSIX shell command that satisfies the request. No markdown, no explanation."
    },
    "code": {
      "system": "You are a senior coding assistant. Return concise, correct code or focused implementation guidance."
    }
  },
  "models": [
    "openai/gpt-4o-mini",
    "anthropic/claude-3.5-sonnet",
    "google/gemini-2.0-flash-001",
    "meta-llama/llama-3.1-70b-instruct"
  ]
}
```

## Usage

```bash
yo what is the fastest way to gzip a folder?
yo shell list files changed today
yo code write a Go function that retries HTTP 429
yo -m anthropic/claude-3.5-sonnet explain this kubectl error
yo -mode shell -temperature 0 find large log files under /var/log
```

Flags:

| Flag | Purpose |
|---|---|
| `-m`, `-model` | OpenRouter model id for this request |
| `-p`, `-mode` | Prompt mode from config |
| `-temperature` | Sampling temperature, 0 to 2 |
| `-list-models` | Print configured model shortcuts |
| `-show-config` | Print active config path |
| `-config` | Use a specific config JSON file |
| `-version` | Print build version |

## Development

```bash
go test ./...
go build -ldflags "-s -w -X main.Version=dev" -o /tmp/yo ./cmd/yo
/tmp/yo -version
```

## Notes

- API calls go to OpenRouter's OpenAI-compatible `/chat/completions` endpoint.
- `HTTP-Referer` and `X-Title` are sent from config for OpenRouter app attribution.
- Config files are created with `0600` permissions because they identify the key environment variable and may later hold sensitive metadata.

## License

MIT
