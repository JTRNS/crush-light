# Messages

A **message** is a single turn in a [session](./sessions.md) conversation. Every user input, assistant response, and [tool](./tools.md) result is stored as a message with a role and a list of [content parts](#content-parts).

## What a message contains

At the data-model layer (`internal/message/message.go`), a message stores:

- `id` – unique identifier.
- `session_id` – foreign key to the owning [session](./sessions.md).
- `role` – one of `user`, `assistant`, `system`, or `tool`.
- `parts` – JSON array of polymorphic [content parts](#content-parts).
- `model` – the [model](./models.md) name that produced the message (assistant messages only).
- `provider` – the [provider](./providers.md) that served the model.
- `created_at` / `updated_at` – Unix timestamps in milliseconds.
- `finished_at` – timestamp when streaming completed (assistant messages).
- `is_summary_message` – flag indicating this message is a [summarization](#summarization-messages) snapshot.

Schema references:

- Table and triggers: `internal/db/migrations/20250424200609_initial.sql`
- Provider/model columns: `internal/db/migrations/20250627000000_add_provider_to_messages.sql`
- Summary flag: `internal/db/migrations/20250810000000_add_is_summary_message.sql`
- Message queries: `internal/db/sql/messages.sql`

## Content parts

The `parts` field is a JSON array where each element is a typed wrapper. The discriminator field `type` selects the concrete shape:

| Type | Struct | Purpose |
|------|--------|---------|
| `text` | `TextContent` | Plain text produced by user or assistant. |
| `reasoning` | `ReasoningContent` | Model chain-of-thought with an optional signature for caching. |
| `image_url` | `ImageURLContent` | Image reference (URL + detail level). |
| `binary` | `BinaryContent` | Raw binary data with MIME type (screenshots, generated files). |
| `tool_call` | `ToolCall` | A [tool](./tools.md) invocation request from the assistant. |
| `tool_result` | `ToolResult` | The output returned by a [tool](./tools.md) after execution. |
| `finish` | `Finish` | Signals why the turn ended (stop, tool use, length, error). |

Content parts are serialized with a wrapper pattern for discriminated unions. This design lets a single message carry mixed content — for example, an assistant turn that contains reasoning, text, and several tool calls.

Implementation: `internal/message/parts.go`, `internal/message/json.go`.

## Roles

| Role | Origin | Notes |
|------|--------|-------|
| `user` | Human input or queued prompt | Always receives an automatic `Finish` part. |
| `assistant` | LLM response | May stream incrementally via `AppendContent()` / `AppendReasoningContent()`. |
| `system` | Injected by the [agent](./agents.md) | Carries [system prompts](./system-prompts.md) and [context files](./context-files.md). |
| `tool` | [Tool](./tools.md) execution results | Created after each tool call completes. |

## Streaming

Assistant messages are built incrementally during generation:

1. A message is created with an empty parts list.
2. As the LLM streams tokens, `AppendContent()` and `AppendReasoningContent()` grow the relevant part in place.
3. Tool call inputs arrive via `AppendToolCallInput()`.
4. When the turn finishes, `finished_at` is set and a `Finish` part is appended.

Each mutation triggers a [pub/sub](./pubsub.md) update event so the UI can render partial content in real time.

## Summarization messages

When [auto-summarization](./agents.md#auto-summarization) triggers, the [agent](./agents.md) condenses the conversation history into a summary message with `is_summary_message = true`. This message replaces prior history for context-window purposes while the originals remain in the database.

## Conversion

Messages convert bidirectionally between the internal representation and the `fantasy.Message` format used by [providers](./providers.md):

- `ToAIMessage()` – internal → provider.
- `FromAIMessage()` – provider → internal.

Text file [attachments](#attachments) are inlined into the system prompt as XML-wrapped content blocks.

## Attachments

Users can attach files or images to a prompt. Attachments are modeled as:

```go
type Attachment struct {
    FilePath string
    Content  string
    MIMEType string
}
```

Text attachments become inline context. Image attachments become `BinaryContent` parts (when the [model](./models.md) supports images).

## Service

`message.Service` (`internal/message/service.go`) provides CRUD and event streaming:

- `Create` / `Update` / `Delete` / `DeleteSessionMessages`
- `Get` / `List` / `ListUserMessages` / `ListAllUserMessages`
- Embeds [`pubsub.Broker[Message]`](./pubsub.md) for real-time event publishing.

## Database triggers

Two triggers on the `messages` table keep the parent [session](./sessions.md) `message_count` in sync:

- `update_session_message_count_on_insert` – increments on INSERT.
- `update_session_message_count_on_delete` – decrements on DELETE.

All messages cascade-delete when their session is removed.
