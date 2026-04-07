# Shared Vocabulary Index

## [Sessions](./sessions.md)
A session is the top-level conversation container that ties together one chat thread and its persisted metadata (title, usage/cost counters, timestamps, and message count). It is also the scope key used to associate related artifacts like messages, file history, read-file tracking, and in-flight/queued agent work.

## [Messages](./messages.md)
A message is a single turn in a session conversation. Every user input, assistant response, and tool result is stored as a message with a role (`user`, `assistant`, `system`, `tool`) and a list of typed content parts (text, reasoning, tool calls, tool results, images, finish signals).

## [Agents](./agents.md)
An agent is a named personality that pairs an LLM model with a set of tools, a system prompt, and a behavioral profile. Crush ships two built-in agents — **Coder** (full tool access, large model) and **Task** (read-only tools, used for parallel research). The agent runtime includes a coordinator, session agents, prompt queuing, and loop detection.

## [Tools](./tools.md)
A tool is a callable action that an agent can invoke during a conversation. Tools cover file operations (view, edit, write), search (glob, grep), shell execution (bash, job management, tmux), web access (fetch, sourcegraph), code intelligence (LSP diagnostics, references), and sub-agent delegation. Each tool has a Go implementation and a markdown description file.

## [Providers](./providers.md)
A provider is a configured connection to an LLM API service. Each provider has a protocol type (OpenAI, Anthropic, Google, Azure, Bedrock, Vertex AI, OpenRouter, or OpenAI-compatible), an API endpoint, authentication credentials, and a list of available models. Crush supports multiple providers simultaneously.

## [Models](./models.md)
A model is a specific LLM identified by a provider and a model ID. Crush uses two model slots — **large** (main conversation, code generation) and **small** (title generation, auto-summarization) — to balance capability against cost. Models can be switched at runtime.

## [Workspaces](./workspaces.md)
A workspace is an isolated runtime instance that binds a project directory to its own database, configuration, agent, and service layer. Each workspace operates independently. Data lives in `.crush/` by default. Workspaces can be accessed locally (in-process) or remotely (via the HTTP server).

## [Configuration](./configuration.md)
Configuration controls every aspect of Crush's behavior and is loaded from `crush.json` files. Multiple files are deep-merged across scopes (global → project → workspace) with increasing priority. It defines providers, models, agents, tools, permissions, and general options.

## [Context Files](./context-files.md)
Context files are project-specific markdown documents (e.g., `AGENTS.md`, `CRUSH.md`, `CLAUDE.md`) that Crush reads at startup and injects into the system prompt. They let project maintainers provide instructions, conventions, and domain knowledge that every agent session inherits automatically.

## [System Prompts](./system-prompts.md)
A system prompt is the foundational instruction set that tells an agent how to behave. It is assembled from Go templates, context files, tool descriptions, skill definitions, and optional provider prefixes. Templates define critical rules, communication style, workflow phases, and editing guidelines.

## [Skills](./skills.md)
A skill is a reusable, self-contained instruction set defined in a `SKILL.md` file following the Agent Skills open standard. Skills are discovered automatically from global and project directories, summarized in the system prompt, and activated on demand when the agent reads the skill file via the view tool.

## [Permissions](./permissions.md)
Permissions gate whether an agent is allowed to execute a tool action with side effects. The system supports one-time grants, persistent grants, session-level auto-approval, and global skip. Currently all requests are auto-approved, but the infrastructure supports interactive approval via the UI.

## [File History](./file-history.md)
File history tracks every version of every file modified during a session. It serves as an undo log, a diff source for the UI sidebar, and a safety mechanism — write-capable tools verify files have not changed since last read before applying edits. A companion file tracker records when each file was last read.

## [Pub/Sub](./pubsub.md)
Pub/sub is Crush's internal event system built on typed generic brokers. Services that mutate state (sessions, messages, file history, permissions) publish events. The TUI and HTTP event stream subscribe to these events for real-time updates. Non-blocking publish prevents slow consumers from stalling producers.

## [Shell](./shell.md)
The shell is Crush's command execution environment, built on a pure-Go POSIX interpreter (`mvdan.cc/sh`) for cross-platform consistency. It provides command blocking for security, background job management (up to 50 concurrent jobs), and streaming output capture. The bash, job_output, and job_kill tools delegate to this layer.
