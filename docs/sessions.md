# Sessions

A **session** is Crush’s canonical unit of conversational state inside a workspace.
It is the identifier that groups everything needed to resume, inspect, and manage one conversation over time.

## What a session contains

At the data-model layer (`internal/session/session.go`), a session stores:

- `id`
- `parent_session_id`
- `title`
- `message_count`
- `prompt_tokens`
- `completion_tokens`
- `summary_message_id`
- `cost`
- `created_at`
- `updated_at`

Schema/query references:

- Table and triggers: `internal/db/migrations/20250424200609_initial.sql`
- Session queries: `internal/db/sql/sessions.sql`
- Added summary link field: `internal/db/migrations/20250515105448_add_summary_message_id.sql`

## What sessions are used for

Sessions are used as the primary grouping key for:

- **Messages** (`messages.session_id`) for conversation history.
- **File history versions** (`files.session_id`) for modified-file timelines.
- **Read-file tracking** (`read_files.session_id`) to remember files viewed in the session.
- **Agent runtime state**, including busy state and queued prompts keyed by session ID.

This makes sessions both a user-visible conversation abstraction and an internal boundary for persistence and agent orchestration.

## Lifecycle

High-level lifecycle operations are implemented in `internal/session/session.go` and exposed through workspace/backend/server layers:

- Create: start a new conversation container.
- Get/List: retrieve one or many sessions (top-level sessions are listed for UX).
- Save/Rename: update metadata such as title/usage.
- Delete: remove session and associated data.

Important deletion behavior: session delete explicitly removes session messages and session files before deleting the session record (`internal/session/session.go`). Read-file rows are also removed by foreign-key cascade (`internal/db/migrations/20260127000000_add_read_files_table.sql`).

## User-facing behavior

In the TUI, sessions drive:

- Landing/session picker flows.
- Resume/continue-last behavior.
- Session-level message loading.
- Session-level file diff sidebar data.

Representative call sites include:

- `internal/ui/model/ui.go`
- `internal/ui/model/session.go`
- `internal/ui/dialog/sessions.go`

## API surface

The HTTP server exposes session CRUD and related endpoints under workspace routes in `internal/server/proto.go`, including:

- list/create/get/update/delete session
- list session history
- list session messages
- list session user messages

These delegate through `internal/backend/session.go` to the session/message/history services for the selected workspace.

## Special session forms

The system also creates internal/non-primary session forms:

- **Title generation sessions**: `title-<parentSessionID>`
- **Agent tool sessions**: `<messageID>$$<toolCallID>`
- **Task/sub sessions** with `parent_session_id` set

Child sessions are still active in the codebase, but they are not used for removed todo functionality. They are used for sub-agent orchestration from the `agent` tool (`internal/agent/agent_tool.go`, `internal/agent/coordinator.go`) and for rendering nested tool activity in the UI (`internal/ui/model/ui.go`).

Top-level session listings and usage stats filter to rows where `parent_session_id IS NULL` (`internal/db/sql/sessions.sql`, `internal/db/sql/stats.sql`), so these helper sessions are not treated as regular user conversations.

Non-interactive continue flows explicitly reject child and agent-tool sessions (`internal/app/app.go`, `internal/cmd/run.go`).
