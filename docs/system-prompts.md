# System Prompts

A **system prompt** is the foundational instruction set that tells an [agent](./agents.md) how to behave. Crush assembles system prompts from Go templates, [context files](./context-files.md), [tool](./tools.md) descriptions, [skill](./skills.md) definitions, and optional [provider](./providers.md) prefixes.

## Templates

System prompt templates live in `internal/agent/templates/` and use Go's `text/template` syntax. The template is executed with runtime data injected as template variables.

### Coder prompt (`coder.md.tpl`)

The main prompt for the [coder agent](./agents.md#coder). It defines:

- **Critical rules** – override-everything constraints (read before editing, test after changes, security, autonomy).
- **Communication style** – concise responses, match user language, no preamble.
- **Workflow** – phases: search/read → execute changes → verify → finish.
- **Decision making** – autonomous by default; stop only when truly ambiguous.
- **Editing rules** – exact whitespace matching, uniqueness requirements.
- **Error handling** – iterative troubleshooting with multiple remediation strategies.
- **Code conventions** – check existing style, use existing libraries, match patterns.

### Task prompt (`task.md.tpl`)

A simpler prompt for the [task agent](./agents.md#task), focused on search and information gathering.

### Initialization prompt (`initialize.md.tpl`)

Used during [project initialization](./context-files.md#project-initialization) to generate a [context file](./context-files.md). It instructs the agent to analyze the codebase and document build commands, architecture, conventions, and gotchas.

### Title prompt (`title.md`)

Generates a short (≤ 50 character) conversation title from the first user [message](./messages.md). Uses the same language as the user input.

### Summary prompt (`summary.md`)

Used by [auto-summarization](./agents.md#auto-summarization) to compress conversation history. The summary must be self-contained because it becomes the only context available when resuming the [session](./sessions.md). Required sections: current state, files & changes, technical context, strategy, and next steps.

## Template variables

Templates receive a data struct with fields including:

| Variable | Content |
|----------|---------|
| `{{.ContextFiles}}` | Concatenated [context file](./context-files.md) contents. |
| `{{.Skills}}` | XML representation of available [skills](./skills.md). |
| `{{.OS}}` | Operating system (GOOS). |
| `{{.CWD}}` | Current working directory. |

The prompt builder reads matching files and injects their content before the template is sent to the [model](./models.md).

## Tool descriptions

Each [tool](./tools.md) has a companion `.md` file in `internal/agent/tools/` that describes its purpose, parameters, and usage guidelines. These descriptions are included in the system prompt alongside the JSON schema that the LLM uses to format tool calls.

Some tools also contribute behavioral templates (e.g., `bash.tpl` provides detailed execution guidelines for the bash tool).

## Provider prefix

[Providers](./providers.md) can define a `system_prompt_prefix` that is prepended to the assembled system prompt. This is used for provider-specific safety policies or behavioral adjustments (e.g., Azure OpenAI content filters).

## Assembly order

The final system prompt is assembled roughly as:

1. Provider prefix (if any).
2. Template output (critical rules, style, workflow).
3. Context files.
4. Skill definitions.
5. Tool descriptions (handled by the `fantasy` library at the protocol level).

## Related

- [Agents](./agents.md) – each agent type has its own system prompt template.
- [Context Files](./context-files.md) – project-specific content injected into prompts.
- [Tools](./tools.md) – tool descriptions become part of the prompt.
- [Skills](./skills.md) – skill instructions are included in the prompt.
- [Providers](./providers.md) – may contribute a prefix.
- [Models](./models.md) – the prompt is sent to the selected model.
