# Context Files

**Context files** are project-specific markdown documents that Crush reads at startup and injects into the [system prompt](./system-prompts.md). They let project maintainers provide instructions, conventions, and domain knowledge that every [agent](./agents.md) session inherits automatically.

## Default discovery paths

Crush looks for the following files in the working directory (`internal/config/config.go`):

| Path | Origin |
|------|--------|
| `.github/copilot-instructions.md` | GitHub Copilot |
| `.cursorrules` | Cursor |
| `.cursor/rules/` | Cursor (directory of rule files) |
| `CLAUDE.md` / `CLAUDE.local.md` | Claude Code |
| `GEMINI.md` / `gemini.md` | Gemini |
| `crush.md` / `Crush.md` / `CRUSH.md` | Crush-specific |
| `crush.local.md` / `Crush.local.md` / `CRUSH.local.md` | Crush-specific (local overrides) |
| `AGENTS.md` / `agents.md` / `Agents.md` | Agent Skills open standard |

Additional paths can be added through `options.context_paths` in [configuration](./configuration.md). Per-[agent](./agents.md) context paths are also supported via the agent's `context_paths` field.

## Local variants

Files ending in `.local.md` (e.g., `CRUSH.local.md`, `CLAUDE.local.md`) are intended for personal overrides that are gitignored. They are loaded alongside their non-local counterparts, giving individual contributors a way to customize agent behavior without affecting the shared repository.

## How context files are used

1. During [configuration](./configuration.md) loading, all matching context file paths are collected.
2. When the [agent](./agents.md) builds its [system prompt](./system-prompts.md), each context file's content is read and injected into the template at the `{{.ContextFiles}}` placeholder.
3. The content appears as part of the system message, giving the LLM access to project-specific instructions on every turn.

This means context files influence all [agent](./agents.md) behavior — coding style, testing patterns, build commands, architecture decisions — without requiring any explicit user prompt.

## What to put in a context file

The [initialization prompt](./system-prompts.md#initialization-prompt) (`internal/agent/templates/initialize.md.tpl`) suggests including:

- Essential build, test, and lint commands.
- Code organization and architecture overview.
- Naming conventions and coding style.
- Testing patterns and frameworks.
- Non-obvious gotchas and implicit conventions.

The goal is agent-friendly documentation: information that helps an LLM operate effectively, not a human onboarding guide.

## Project initialization

When Crush opens a [workspace](./workspaces.md) in a directory for the first time and no context files exist, it can auto-generate one. The default target file is `AGENTS.md` (configurable via `options.initialize_as`). The [agent](./agents.md) analyzes the repository and writes a context file describing the project. A `.crush/init` flag prevents re-initialization on subsequent runs.

Implementation: `internal/config/init.go`.

## Related

- [System Prompts](./system-prompts.md) – context files are injected into the system prompt template.
- [Configuration](./configuration.md) – discovery paths are configured here.
- [Agents](./agents.md) – consume context files as part of their instructions.
- [Skills](./skills.md) – a complementary mechanism for reusable task-specific instructions.
- [Workspaces](./workspaces.md) – context files are resolved relative to the workspace root.
