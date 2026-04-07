# Agents

An **agent** is a named personality that pairs an LLM [model](./models.md) with a set of [tools](./tools.md), a [system prompt](./system-prompts.md), and a behavioral profile. Crush ships two built-in agents — **Coder** and **Task** — and exposes a runtime that manages their execution within [sessions](./sessions.md).

## Built-in agents

### Coder

The primary agent. It uses the **large** [model](./models.md) and has access to every enabled [tool](./tools.md), including write-capable ones (`edit`, `multiedit`, `write`, `bash`). It is the agent users interact with in both the TUI and non-interactive mode.

### Task

A read-only research agent. It also uses the **large** [model](./models.md) but is restricted to search and read [tools](./tools.md): `glob`, `grep`, `ls`, `sourcegraph`, and `view`. The coder agent can spawn task agents via the `agent` [tool](./tools.md) for parallel information gathering.

Agent definitions live in [configuration](./configuration.md) under `agents` and are set up at startup by `config.SetupAgents()` (`internal/config/config.go`).

## Architecture

The agent subsystem has two main runtime components:

### Coordinator

The **coordinator** (`internal/agent/coordinator.go`) is the top-level orchestrator:

- Creates and configures `SessionAgent` instances.
- Resolves [models](./models.md) and merges call options from [provider](./providers.md) config.
- Builds the [tool](./tools.md) set for each agent type.
- Handles OAuth token refresh on 401 responses.
- Exposes `Run()`, `Cancel()`, `UpdateModels()`, and prompt-queue management.

### SessionAgent

The **session agent** (`internal/agent/agent.go`) runs the actual conversation loop for a single [session](./sessions.md):

1. Accepts a prompt (or dequeues one).
2. Loads [message](./messages.md) history for the session.
3. Creates a user [message](./messages.md).
4. Streams the LLM response, executing [tool](./tools.md) calls as they arrive.
5. Accumulates token usage and updates the [session](./sessions.md).
6. Checks whether [auto-summarization](#auto-summarization) is needed.
7. Repeats until the model stops or the loop is cancelled.

## Prompt queue

When a [session](./sessions.md) is already processing a prompt, additional prompts are queued in memory. After the current run completes, queued prompts are drained in order. This prevents request loss during rapid user input.

## Title generation

On the first user [message](./messages.md) in a session, the agent automatically generates a short title using the **small** [model](./models.md). This runs in a lightweight child [session](./sessions.md) (`title-<parentSessionID>`) so it does not pollute the main conversation.

## Auto-summarization

Long conversations can exceed a [model's](./models.md) context window. The agent detects this when:

- The context exceeds 200 000 tokens, **or**
- More than 80 % of the context window is consumed.

When triggered, the **small** [model](./models.md) compresses the conversation history into a summary [message](./messages.md) (`is_summary_message = true`) following the template in `internal/agent/templates/summary.md`. The summary replaces prior history for future turns while originals remain in the database.

Auto-summarization can be disabled via `options.disable_auto_summarize` in [configuration](./configuration.md).

## Loop detection

To prevent an agent from getting stuck in an infinite tool-call cycle, `hasRepeatedToolCalls()` (`internal/agent/loop_detection.go`) examines the last 10 steps. If any tool signature (SHA-256 of name + input + output) repeats five or more times, the agent stops.

## Sub-agents

The `agent` [tool](./tools.md) (`internal/agent/agent_tool.go`) launches a task agent as a sub-agent with its own child [session](./sessions.md). Sub-agents run with a restricted [tool](./tools.md) set (glob, grep, ls, view) and are stateless — they execute once and return results to the parent agent. Multiple sub-agents can run in parallel.

## Execution flow

```
Coordinator.Run(sessionID, prompt)
  → resolve model + provider
  → refresh OAuth if needed
  → SessionAgent.Run(SessionAgentCall)
      → load history
      → create user message
      → fantasy.Agent.Stream(...)
          → LLM turn → tool calls → tool results → next turn
      → update session usage
      → check summarization
      → drain prompt queue
```

The `fantasy` library (`charm.land/fantasy`) handles the low-level agent loop, streaming, and [provider](./providers.md) protocol translation.

## Related

- [Sessions](./sessions.md) – the conversational scope agents operate in.
- [Tools](./tools.md) – capabilities available to agents.
- [Models](./models.md) – LLM selection.
- [System Prompts](./system-prompts.md) – behavioral instructions.
- [Permissions](./permissions.md) – tool authorization.
- [Messages](./messages.md) – conversation turns produced by agents.
