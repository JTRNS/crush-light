# File History

**File history** tracks every version of every file modified during a [session](./sessions.md). It serves as an undo log, a diff source for the UI, and a safety mechanism that prevents [tools](./tools.md) from overwriting changes they haven't seen.

## Versioning

Each file modification creates a new version record in the `files` table:

| Column | Purpose |
|--------|---------|
| `id` | Unique record identifier. |
| `session_id` | Owning [session](./sessions.md). |
| `path` | Relative file path. |
| `content` | Full file content at this version. |
| `version` | Auto-incrementing version number (starts at 0). |
| `created_at` / `updated_at` | Unix timestamps in milliseconds. |

A unique constraint on `(path, session_id, version)` prevents duplicate versions. When a version conflict occurs, the service retries with an incremented version number (up to 3 attempts).

Schema: `internal/db/migrations/20250424200609_initial.sql`.

## Service API

`history.Service` (`internal/history/file.go`) provides:

| Method | Behavior |
|--------|----------|
| `Create` | Store the initial version (version 0) of a file. |
| `CreateVersion` | Store a new version, auto-incrementing from the latest. |
| `Get` / `GetByPathAndSession` | Retrieve a specific version. |
| `ListBySession` | All file versions in a session. |
| `ListLatestSessionFiles` | Latest version of each unique path in a session. |
| `Delete` / `DeleteSessionFiles` | Remove versions. Cascade-deletes when a [session](./sessions.md) is deleted. |

The service embeds a [`pubsub.Broker[File]`](./pubsub.md) and publishes created/deleted events.

## File tracker

The **file tracker** (`internal/filetracker/service.go`) is a companion service that records when each file was last read by the [agent](./agents.md). It stores records in the `read_files` table:

| Column | Purpose |
|--------|---------|
| `session_id` | Owning [session](./sessions.md). |
| `path` | Relative file path. |
| `read_at` | Unix timestamp in seconds (note: seconds, not milliseconds). |

The file tracker is used by write-capable [tools](./tools.md) (`edit`, `multiedit`, `write`) to verify that the agent has seen the current file content before making changes. If a file has been modified on disk since the last recorded read, the tool warns the agent or rejects the edit.

The `view` tool records a read each time a file is opened. The `edit` and `write` tools check `LastReadTime()` before applying changes.

API: `RecordRead`, `LastReadTime`, `ListReadFiles`.

Schema: `internal/db/migrations/20260127000000_add_read_files_table.sql`.

## How tools use file history

1. **View** – records a read via the file tracker.
2. **Edit / MultiEdit** – checks `LastReadTime` to detect stale reads, then calls `CreateVersion` to snapshot the new content.
3. **Write** – same stale-read check, then stores the new content as a version.

The TUI uses `ListLatestSessionFiles` to populate the file diff sidebar, showing which files changed during the session and providing before/after comparisons.

## Related

- [Sessions](./sessions.md) – file history is scoped per session.
- [Tools](./tools.md) – file-modifying tools create versions and check read times.
- [Messages](./messages.md) – tool results reference the files that changed.
- [Pub/Sub](./pubsub.md) – file events are published for UI updates.
- [Workspaces](./workspaces.md) – file history is stored in the workspace database.
