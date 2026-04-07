# Tools

A **tool** is a callable action that an [agent](./agents.md) can invoke during a conversation to interact with the file system, run commands, search code, or access external services. Each tool has a name, a markdown description, and a Go implementation that satisfies the `fantasy.AgentTool` interface.

## Tool catalogue

### File operations

| Tool | File | Purpose |
|------|------|---------|
| `view` | `view.go` | Read file or directory contents with line numbers. Supports images, offset/limit pagination, and suggests similar filenames on miss. Max 1 MB / 2 000 lines. |
| `edit` | `edit.go` | Single find-and-replace in an existing file. Requires exact whitespace matching and verifies the file has not changed since last read. |
| `multiedit` | `multiedit.go` | Apply multiple sequential find-and-replace operations to one file. Supports partial success. |
| `write` | `write.go` | Create or overwrite a file with full content. Auto-creates parent directories. |

### Search and discovery

| Tool | File | Purpose |
|------|------|---------|
| `glob` | `glob.go` | Find files by glob pattern. Uses ripgrep when available, respects `.gitignore` and `.crushignore`. Max 100 results. |
| `grep` | `grep.go` | Search file contents by regex or literal text. Respects ignore files, returns results grouped by file with line numbers. Max 100 matches. |
| `ls` | `ls.go` | List directory tree structure. Respects ignore files. Max 1 000 entries. |

### Shell execution

| Tool | File | Purpose |
|------|------|---------|
| `bash` | `bash.go` | Execute a [shell](./shell.md) command. Supports background jobs, auto-backgrounding after a timeout, and enforces a banned-command list for security. |
| `job_output` | `job_output.go` | Retrieve output from a background [shell](./shell.md) job by `shell_id`. |
| `job_kill` | `job_kill.go` | Terminate a background [shell](./shell.md) job. |
| `tmux` | `tmux.go` | Manage persistent tmux sessions for long-running processes (servers, watchers). Operations: create, send_keys, capture, list, kill. |

### Web and external

| Tool | File | Purpose |
|------|------|---------|
| `fetch` | `fetch.go` | Retrieve raw content from a URL. Supports text, markdown, and HTML formats. Max 1 MB response. No cookies or authentication. |
| `sourcegraph` | `sourcegraph.go` | Cross-repository code search via the Sourcegraph GraphQL API. Max 20 results. |

### Code intelligence

| Tool | File | Purpose |
|------|------|---------|
| `lsp_diagnostics` | `diagnostics.go` | Collect lint and type-check diagnostics from active LSP servers. Groups results by severity. |
| `lsp_references` | `references.go` | Find all references to a symbol via LSP. Semantic-aware, not plain text search. |
| `lsp_restart` | `lsp_restart.go` | Restart all LSP server instances. |

### Sub-agent

| Tool | File | Purpose |
|------|------|---------|
| `agent` | `agent_tool.go` | Launch a read-only [task agent](./agents.md#task) as a sub-agent for parallel research. The sub-agent gets glob, grep, ls, and view only. |

All tool implementations live under `internal/agent/tools/`. Each tool also has a companion `.md` file in the same directory that provides the description injected into the [system prompt](./system-prompts.md).

## How tools execute

1. The [agent](./agents.md) loop receives a `tool_call` [content part](./messages.md#content-parts) from the LLM.
2. The matching tool's `Call()` method runs with the parsed parameters.
3. The result (string + metadata map) is wrapped in a `tool_result` [content part](./messages.md#content-parts) and appended to the conversation.
4. The [agent](./agents.md) loop feeds the result back to the LLM for the next turn.

Tools that modify files (`edit`, `multiedit`, `write`) record each change in [file history](./file-history.md) and verify that the file has not been modified since it was last read (via the [file tracker](./file-history.md#file-tracker)).

## Tool filtering

Not every [agent](./agents.md) gets every tool. The [configuration](./configuration.md) controls two filtering mechanisms:

- **`allowed_tools`** on each agent definition determines its tool set. The task agent is restricted to read-only tools.
- **`options.disabled_tools`** globally hides specific tools from all agents.

Filtering logic: `resolveAllowedTools()` and `resolveReadOnlyTools()` in `internal/config/config.go`.

## Permissions

Before executing a tool that may have side effects, the [permission](./permissions.md) service is consulted. In the current implementation all requests are auto-approved, but the infrastructure supports interactive approval via the UI.

## Ignore files

`glob` and `grep` respect two ignore files:

- `.gitignore` – standard Git ignore rules.
- `.crushignore` – additional patterns specific to Crush, placed alongside `.gitignore`.

## Context keys

Tools receive session-scoped metadata via Go context values:

| Key | Purpose |
|-----|---------|
| `SessionIDContextKey` | Current [session](./sessions.md) ID for file tracking and history. |
| `MessageIDContextKey` | Current [message](./messages.md) ID. |
| `SupportsImagesContextKey` | Whether the [model](./models.md) supports image input. |
| `ModelNameContextKey` | Active model name string. |

## Related

- [Agents](./agents.md) – consumers of tools.
- [Shell](./shell.md) – execution environment for bash/job tools.
- [Permissions](./permissions.md) – authorization for tool execution.
- [File History](./file-history.md) – version tracking for file-modifying tools.
- [System Prompts](./system-prompts.md) – tool descriptions are injected here.
