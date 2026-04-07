# Configuration

**Configuration** controls every aspect of Crush's behavior: which [providers](./providers.md) are available, which [models](./models.md) are selected, what [tools](./tools.md) are enabled, how [agents](./agents.md) behave, and where data is stored. It is loaded from JSON files, merged across scopes, and exposed at runtime through a `ConfigStore`.

## Config file

The primary configuration file is `crush.json`. Its top-level shape (`internal/config/config.go`):

| Key | Type | Purpose |
|-----|------|---------|
| `models` | `map[large\|small]SelectedModel` | Active [model](./models.md) selection per slot. |
| `recent_models` | `map[large\|small][]SelectedModel` | History of recently used [models](./models.md). |
| `providers` | `map[string]ProviderConfig` | [Provider](./providers.md) definitions. |
| `mcp` | `map[string]MCPConfig` | MCP server connections. |
| `lsp` | `map[string]LSPConfig` | LSP server definitions. |
| `options` | `Options` | General settings (see below). |
| `permissions` | `Permissions` | [Permission](./permissions.md) rules. |
| `tools` | `Tools` | Per-tool configuration overrides. |
| `agents` | `map[string]Agent` | [Agent](./agents.md) definitions. |

## Scopes and merging

Configuration is loaded from multiple files and merged with increasing priority:

| Priority | Scope | Path | Override env |
|----------|-------|------|--------------|
| 1 (lowest) | Global config | `~/.config/crush/crush.json` | `CRUSH_GLOBAL_CONFIG` |
| 2 | Global data | `~/.local/share/crush/crush.json` | `CRUSH_GLOBAL_DATA` |
| 3 | Project (walking up) | `crush.json` or `.crush.json` in any ancestor directory | — |
| 4 (highest) | [Workspace](./workspaces.md) | `.crush/crush.json` | — |

Files are deep-merged using `jsons.Merge()`. A field set in a higher-priority file overrides the same field from a lower-priority one.

Loading logic: `internal/config/load.go`.

## Options

The `options` block holds general settings:

| Field | Purpose |
|-------|---------|
| `context_paths` | Extra [context file](./context-files.md) paths to load. |
| `skills_paths` | Directories to discover [skills](./skills.md). |
| `data_directory` | Override the `.crush` data directory. |
| `disabled_tools` | [Tools](./tools.md) hidden from all agents globally. |
| `disabled_skills` | [Skills](./skills.md) to exclude. |
| `disable_auto_summarize` | Turn off [auto-summarization](./agents.md#auto-summarization). |
| `disable_provider_auto_update` | Don't fetch updates from [Catwalk](./providers.md#catwalk). |
| `disable_default_providers` | Ignore all embedded [providers](./providers.md). |
| `debug` / `debug_lsp` | Enable debug logging. |
| `attribution` | Git author info for commits. |
| `initialize_as` | Default [context file](./context-files.md) name for project init (default: `AGENTS.md`). |
| `auto_lsp` | Auto-discover and start LSP servers. |
| `progress` | Show indeterminate progress updates. |
| `disable_notifications` | Suppress desktop notifications. |

## ConfigStore

`ConfigStore` (`internal/config/store.go`) is the runtime wrapper around a loaded `Config`. It provides:

- Read access to all config values.
- Mutation methods that write back to the appropriate scope file (`SetConfigField`, `RemoveConfigField`, `UpdatePreferredModel`, `SetProviderAPIKey`).
- Agent setup: `SetupAgents()` creates the built-in [coder and task agents](./agents.md#built-in-agents).
- Provider refresh: triggers [Catwalk](./providers.md#catwalk) updates.

## Environment overrides

Several settings can be overridden by environment variables:

| Variable | Effect |
|----------|--------|
| `CRUSH_DISABLE_PROVIDER_AUTO_UPDATE` | Skip [Catwalk](./providers.md#catwalk) fetch. |
| `CRUSH_GLOBAL_CONFIG` | Custom global config path. |
| `CRUSH_GLOBAL_DATA` | Custom global data path. |
| `CRUSH_SKILLS_DIR` | Custom [skills](./skills.md) directory. |
| `CRUSH_CACHE_DIR` | Custom cache directory. |

[Provider](./providers.md)-specific variables like `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `VERTEXAI_PROJECT`, etc., are checked during provider validation.

## Related

- [Providers](./providers.md) – configured under `providers`.
- [Models](./models.md) – selected under `models`.
- [Agents](./agents.md) – defined under `agents`.
- [Tools](./tools.md) – filtered by `options.disabled_tools`.
- [Skills](./skills.md) – discovered from `options.skills_paths`.
- [Context Files](./context-files.md) – paths listed in `options.context_paths`.
- [Workspaces](./workspaces.md) – each workspace has its own config scope.
