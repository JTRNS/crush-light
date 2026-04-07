# Models

A **model** is a specific LLM identified by a [provider](./providers.md) and a model ID (e.g., `openai/gpt-4o`). Crush uses two model slots — **large** and **small** — to balance capability against cost across different tasks.

## Model slots

| Slot | Default use | Typical assignment |
|------|-------------|-------------------|
| `large` | Main conversation, code generation, [tool](./tools.md) use | GPT-4o, Claude Sonnet, Gemini Pro |
| `small` | Title generation, [auto-summarization](./agents.md#auto-summarization) | GPT-4o-mini, Claude Haiku, Gemini Flash |

Both [built-in agents](./agents.md#built-in-agents) default to the large slot. The small model is used internally for lightweight auxiliary tasks.

Slots are defined as `SelectedModelType` constants (`large`, `small`) in `internal/config/config.go`.

## Model selection

A selected model combines a [provider](./providers.md) ID with a model ID and optional generation parameters:

| Field | Purpose |
|-------|---------|
| `provider` | Which [provider](./providers.md) serves this model. |
| `model` | Model identifier within the provider (e.g., `gpt-4o`). |
| `temperature` | Sampling temperature (0.0–2.0). |
| `top_p` / `top_k` | Nucleus / top-k sampling. |
| `max_tokens` | Maximum output tokens. |
| `reasoning_effort` | For reasoning models: `low`, `medium`, or `high`. |
| `think` | Enable extended thinking (Anthropic). |
| `provider_options` | Arbitrary provider-specific overrides. |

These fields are stored in [configuration](./configuration.md) under `models.large` and `models.small`.

## Default resolution

When no explicit model is configured, Crush picks defaults automatically:

1. Find the first enabled known [provider](./providers.md) that has valid credentials.
2. Use that provider's `DefaultLargeModelID` and `DefaultSmallModelID` from the [Catwalk](./providers.md#catwalk) registry.
3. If no known provider qualifies, fall back to the first enabled custom provider's first model.

This logic lives in `defaultModelSelection()` in `internal/config/load.go`.

## Recent models

Crush tracks recently used models per slot in `recent_models` within [configuration](./configuration.md). This powers the model picker in the TUI, letting users quickly switch back to a previously used model.

## Context window

Each model definition from [Catwalk](./providers.md#catwalk) includes a `context_window` size (in tokens). The [agent](./agents.md) uses this value to decide when [auto-summarization](./agents.md#auto-summarization) is needed and how many [messages](./messages.md) can fit in a single request.

## Runtime updates

Models can be changed at runtime through:

- The TUI model picker dialog.
- CLI flags: `--large-model provider/model`, `--small-model provider/model`.
- The HTTP API: `POST /v1/workspaces/{id}/config/model`.

All paths converge on `config.UpdatePreferredModel()`, which persists the choice and notifies the [coordinator](./agents.md#coordinator) via `UpdateModels()`.

## Related

- [Providers](./providers.md) – models belong to providers.
- [Configuration](./configuration.md) – model selection is persisted here.
- [Agents](./agents.md) – consume models for generation.
- [Messages](./messages.md) – record which model produced each response.
