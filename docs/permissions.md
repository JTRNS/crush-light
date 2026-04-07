# Permissions

**Permissions** gate whether an [agent](./agents.md) is allowed to execute a [tool](./tools.md) action that may have side effects. The permission system provides the infrastructure for interactive approval, although the current implementation auto-approves all requests.

## How permissions work

1. A [tool](./tools.md) calls `permission.Service.Request()` with details about the action (tool name, description, path, parameters).
2. The request is published via [pub/sub](./pubsub.md) so the UI can display it.
3. The service returns a grant/deny decision.
4. If denied, the tool receives `ErrorPermissionDenied` and reports the denial to the [agent](./agents.md).

In the current implementation, `Request()` always returns `true` (auto-approval). The request event is still published for UI visibility.

Implementation: `internal/permission/permission.go`.

## Permission types

| Type | Scope | Behavior |
|------|-------|----------|
| **One-time grant** | Single request | Approves one specific action. |
| **Persistent grant** | Rest of [session](./sessions.md) | Approves all future matching actions in the session. |
| **Session auto-approve** | Entire [session](./sessions.md) | Skips permission checks for all tools. Set automatically in non-interactive mode. |
| **Global skip** | All sessions | Disables permission requests entirely (`SetSkipRequests`). |

## Permission request fields

| Field | Purpose |
|-------|---------|
| `session_id` | Owning [session](./sessions.md). |
| `tool_call_id` | The [tool call](./messages.md#content-parts) that triggered the request. |
| `tool_name` | Name of the [tool](./tools.md) (e.g., `bash`, `edit`). |
| `description` | Human-readable summary of what the action does. |
| `action` | Specific action type within the tool. |
| `path` | File or resource path affected. |
| `params` | Full tool parameters for inspection. |

## Notifications

After a decision, a `PermissionNotification` is published on a separate [pub/sub](./pubsub.md) broker. The notification carries `tool_call_id`, `granted`, and `denied` flags. The UI subscribes to these to update its display.

## Configuration

[Permission rules](./configuration.md) can be defined in `crush.json` under the `permissions` key. The `permission.Service` supports command-level blocking via `BlockFunc` chains that can reject requests based on:

- Command name.
- Subcommand.
- Flags.
- Arguments.

## Related

- [Tools](./tools.md) – request permission before executing side effects.
- [Agents](./agents.md) – receive permission denials as tool errors.
- [Sessions](./sessions.md) – permission scope is per-session.
- [Pub/Sub](./pubsub.md) – permission events are published and consumed.
- [Configuration](./configuration.md) – permission rules are declared here.
