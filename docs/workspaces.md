# Workspaces

A **workspace** is an isolated runtime instance that binds a project directory to its own database, [configuration](./configuration.md), [agent](./agents.md), and service layer. Each workspace operates independently, so multiple projects can run side-by-side on the same Crush server without interfering with each other.

## What a workspace contains

A workspace wraps an `app.App` instance (`internal/backend/workspace.go`) with:

| Field | Purpose |
|-------|---------|
| `ID` | UUID assigned at creation. |
| `Path` | Absolute path to the project directory. |
| `Cfg` | A [`ConfigStore`](./configuration.md) scoped to this project. |
| `Env` | Environment variables inherited from the creating process. |
| `*app.App` | Embedded application core (see below). |

## The App core

`app.App` (`internal/app/app.go`) is the dependency root that wires together every service a workspace needs:

- [`session.Service`](./sessions.md) – session CRUD.
- [`message.Service`](./messages.md) – message storage and streaming.
- [`history.Service`](./file-history.md) – file version tracking.
- [`permission.Service`](./permissions.md) – tool authorization.
- [`filetracker.Service`](./file-history.md#file-tracker) – read-time tracking.
- [`agent.Coordinator`](./agents.md#coordinator) – LLM conversation orchestration.
- LSP manager – language server lifecycle.
- [Pub/sub](./pubsub.md) brokers – event routing.

All services share a single SQLite connection scoped to the workspace's data directory.

## Data directory

By default, workspace data lives in `.crush/` relative to the project root:

```
.crush/
├── crush.json        # workspace-scoped configuration
├── crush.db          # SQLite database (sessions, messages, file history)
├── skills/           # project-specific skills
└── init              # flag file: project has been initialized
```

The path can be overridden with `options.data_directory` in [configuration](./configuration.md).

## Lifecycle

### Creation

A workspace is created when:

- The TUI starts for a project directory.
- A non-interactive `crush run` command targets a directory.
- The HTTP API receives `POST /v1/workspaces` with a path.

Creation opens the database, runs migrations, instantiates all services, subscribes to [pub/sub](./pubsub.md) events, validates [configuration](./configuration.md), and initializes the [coder agent](./agents.md#coder).

### Deletion

Deleting a workspace shuts down its services, closes the database, and removes it from the backend workspace map. When the last workspace is removed, the server's shutdown callback fires.

## Backend management

The `Backend` struct (`internal/backend/backend.go`) manages a concurrent map of active workspaces keyed by UUID. It exposes:

- `CreateWorkspace(path, env)` – spin up a new workspace.
- `GetWorkspace(id)` / `ListWorkspaces()` – retrieve running workspaces.
- `DeleteWorkspace(id)` – tear down and remove.

All [session](./sessions.md), [agent](./agents.md), [configuration](./configuration.md), and event operations are scoped through a workspace reference.

## Local vs. remote

Crush can run in two modes:

| Mode | How workspace is accessed |
|------|--------------------------|
| **Local** | `AppWorkspace` – services run in-process. Used by the TUI and `crush run`. |
| **Remote** | `ClientWorkspace` – a thin HTTP client talks to a Crush server that hosts the workspace. Used when a server is already running. |

The CLI's `setupWorkspace()` function decides which mode to use. Both implement the same `workspace.Workspace` interface so consumers are unaware of the difference.

## Project initialization

On first use in a new directory, the workspace checks whether the project needs initialization (`config.ProjectNeedsInitialization`). If so, the [agent](./agents.md) generates a default [context file](./context-files.md) (typically `AGENTS.md`) describing the project, and a `.crush/init` flag is written to prevent re-initialization.

## Related

- [Configuration](./configuration.md) – scoped per workspace.
- [Sessions](./sessions.md) – live inside a workspace.
- [Agents](./agents.md) – one coordinator per workspace.
- [Pub/Sub](./pubsub.md) – event routing within a workspace.
