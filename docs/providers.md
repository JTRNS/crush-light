# Providers

A **provider** is a configured connection to an LLM API service. Each provider has an ID, a protocol type, an API endpoint, authentication credentials, and a list of available [models](./models.md). Crush supports multiple providers simultaneously and resolves which one to use based on the active [model](./models.md) selection.

## Provider types

Provider types determine the API protocol used to communicate with the service. The protocol translation is handled by the `fantasy` library (`charm.land/fantasy`), which the [agent](./agents.md) runtime uses to stream completions:

| Type | Service | Notes |
|------|---------|-------|
| `openai` | OpenAI | Standard OpenAI chat completions API. |
| `openai-compatible` | Ollama, vLLM, LiteLLM, etc. | Any service that implements the OpenAI API contract. |
| `anthropic` | Anthropic | Claude models with native tool use and extended thinking. |
| `google` | Google Gemini | Gemini models via the Google AI API. |
| `azure` | Azure OpenAI | OpenAI models hosted on Azure. Requires `AZURE_OPENAI_API_VERSION`. |
| `bedrock` | AWS Bedrock | Anthropic models via AWS. Requires AWS credentials. |
| `vertexai` | Google Vertex AI | Gemini models via GCP. Requires `VERTEXAI_PROJECT` and `VERTEXAI_LOCATION`. |
| `openrouter` | OpenRouter | Proxy service routing to multiple upstream providers. |

## Configuration

A provider is configured in `crush.json` under the `providers` key:

```json
{
  "providers": {
    "my-provider": {
      "type": "openai-compatible",
      "base_url": "http://localhost:11434/v1",
      "api_key": "$OLLAMA_API_KEY",
      "models": [...]
    }
  }
}
```

Key fields (`internal/config/config.go`):

| Field | Purpose |
|-------|---------|
| `type` | Protocol type (see table above). |
| `base_url` | API endpoint URL. |
| `api_key` | API key, may use [variable resolution](#variable-resolution). |
| `oauth` | OAuth2 token (used by Copilot integration). |
| `disable` | Exclude this provider entirely. |
| `system_prompt_prefix` | Text prepended to [system prompts](./system-prompts.md) for this provider. |
| `extra_headers` | Additional HTTP headers sent with every request. |
| `extra_body` | Additional fields merged into the request body. |
| `extra_params` | Provider-specific params (e.g., AWS region). |
| `models` | List of [model](./models.md) definitions available from this provider. |

## Known vs. custom providers

**Known providers** ship as embedded defaults and are updated from the [Catwalk](#catwalk) registry. They include major services like OpenAI, Anthropic, Google, and others. Known providers are automatically configured when the matching API key environment variable is set.

**Custom providers** are user-defined entries in [configuration](./configuration.md). They require at minimum a `type`, `base_url`, and at least one [model](./models.md) definition.

When a known and custom provider share the same ID, user [configuration](./configuration.md) is merged on top of the known defaults.

## Variable resolution

API keys and other string fields support three substitution forms:

| Form | Example | Behavior |
|------|---------|----------|
| `$(command)` | `$(cat ~/.secret)` | Command substitution — executes in a shell. |
| `$VAR` | `$OPENAI_API_KEY` | Environment variable expansion. |
| `${VAR}` | `${OPENAI_API_KEY}` | Braced environment variable expansion. |

Resolution is performed by `ShellVariableResolver` in `internal/config/resolve.go`.

## Catwalk

**Catwalk** is Charm's provider and model registry service hosted at `https://catwalk.charm.sh`. At startup, Crush fetches the latest provider and [model](./models.md) catalogue from Catwalk and caches it locally.

Fallback chain:

1. Fresh fetch from Catwalk → cache to `~/.local/share/crush/providers.json`.
2. If fetch fails → use cached copy.
3. If cache is missing → use embedded defaults compiled into the binary.

Auto-update can be disabled with `CRUSH_DISABLE_PROVIDER_AUTO_UPDATE=1` or `options.disable_provider_auto_update` in [configuration](./configuration.md).

## OAuth / Copilot

GitHub Copilot uses OAuth2 tokens instead of static API keys. Crush can import an existing Copilot authentication token or perform its own OAuth device flow via `crush login`. Token refresh is handled automatically by the [coordinator](./agents.md#coordinator) when a 401 response is received.

## Provider validation

During [configuration](./configuration.md) loading, each provider is validated against type-specific requirements:

- **Bedrock** – AWS credentials must be resolvable.
- **Vertex AI** – `VERTEXAI_PROJECT` and `VERTEXAI_LOCATION` environment variables required.
- **Azure** – `AZURE_OPENAI_API_VERSION` environment variable required.
- **Others** – a non-empty API key is required.

Providers that fail validation are silently skipped.

## Related

- [Models](./models.md) – model definitions live within provider config.
- [Configuration](./configuration.md) – where providers are declared.
- [Agents](./agents.md) – consume providers via the coordinator.
- [System Prompts](./system-prompts.md) – `system_prompt_prefix` is injected here.
