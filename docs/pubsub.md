# Pub/Sub

**Pub/sub** (publish–subscribe) is Crush's internal event system. It decouples services that produce state changes from consumers that react to them — primarily the TUI and the HTTP event stream.

## Broker

The core primitive is `pubsub.Broker[T]` (`internal/pubsub/broker.go`), a generic, typed event broker:

```go
type Broker[T any] struct { ... }
```

A broker manages a set of subscriber channels. When `Publish(eventType, payload)` is called, every active subscriber receives a copy of the event. If a subscriber's channel is full, the event is dropped (non-blocking publish) to prevent slow consumers from stalling producers.

### Configuration

| Parameter | Default | Purpose |
|-----------|---------|---------|
| Buffer size | 64 | Channel capacity per subscriber. |
| Max events | 1 000 | Informational limit (not enforced). |

### Lifecycle

- `Subscribe(ctx)` returns a channel. The subscription is active until the context is cancelled, at which point the channel is closed and removed.
- `Shutdown()` closes all subscriber channels and prevents new subscriptions.
- Thread-safe via `sync.RWMutex`.

## Event types

Every event carries a type tag:

| `EventType` | Meaning |
|-------------|---------|
| `created` | A new entity was created. |
| `updated` | An existing entity was modified. |
| `deleted` | An entity was removed. |

## Who publishes

Each service that mutates state embeds its own `Broker[T]`:

| Service | Broker type | Events |
|---------|-------------|--------|
| [`session.Service`](./sessions.md) | `Broker[Session]` | Session created, updated, deleted. |
| [`message.Service`](./messages.md) | `Broker[Message]` | Message created, updated, deleted. |
| [`history.Service`](./file-history.md) | `Broker[File]` | File version created, deleted. |
| [`permission.Service`](./permissions.md) | `Broker[PermissionRequest]` | Permission requested. |
| [`permission.Service`](./permissions.md) | `Broker[PermissionNotification]` | Permission granted or denied. |
| Agent notifications | `Broker[pubsub.Payload]` | Agent lifecycle events. |

## Who subscribes

### TUI

`app.setupEvents()` (`internal/app/app.go`) subscribes to every broker and funnels all events into a single `chan tea.Msg` that the Bubble Tea runtime consumes. This drives real-time UI updates — streaming [message](./messages.md) content, session list changes, permission dialogs, and file diff refreshes.

### HTTP event stream

The server exposes `GET /v1/workspaces/{id}/events` as a Server-Sent Events (SSE) endpoint. The [backend](./workspaces.md#backend-management) subscribes to workspace events and serializes them as `pubsub.Payload` JSON envelopes with a `type` discriminator:

| `PayloadType` | Source |
|---------------|--------|
| `message` | [Message](./messages.md) events. |
| `session` | [Session](./sessions.md) events. |
| `file` | [File history](./file-history.md) events. |
| `permission_request` | [Permission](./permissions.md) requests. |
| `permission_notification` | [Permission](./permissions.md) decisions. |
| `agent_event` | [Agent](./agents.md) lifecycle events. |
| `lsp_event` | LSP state changes. |
| `mcp_event` | MCP state changes. |

## Design rationale

Pub/sub keeps services loosely coupled. The [session](./sessions.md) service does not know about the TUI; it simply publishes events. This makes it straightforward to add new consumers (HTTP SSE, test harnesses, future integrations) without modifying producers.

The non-blocking publish policy ensures that a slow or crashed subscriber cannot back-pressure the [agent](./agents.md) loop or database writes.

## Related

- [Sessions](./sessions.md), [Messages](./messages.md), [File History](./file-history.md), [Permissions](./permissions.md) – all publish events.
- [Agents](./agents.md) – agent lifecycle events are published.
- [Workspaces](./workspaces.md) – event routing is scoped per workspace.
