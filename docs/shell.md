# Shell

The **shell** is Crush's command execution environment. It provides cross-platform POSIX shell interpretation, security-scoped command blocking, and background job management. The [bash tool](./tools.md) and related job [tools](./tools.md) delegate all execution to this layer.

## Execution model

Crush does not fork a system shell. Instead, it uses `mvdan.cc/sh` — a pure-Go POSIX shell interpreter — to parse and execute commands in-process. This ensures consistent behavior across Linux, macOS, and Windows without requiring Bash to be installed.

Each `Shell` instance (`internal/shell/shell.go`) maintains:

| State | Purpose |
|-------|---------|
| `cwd` | Working directory, updated by `cd` within commands. |
| `env` | Environment variables, including automatic entries (see below). |
| `blockFuncs` | Chain of [command blockers](#command-blocking) for security. |

Commands run via `Exec()` return stdout/stderr as streaming output. The shell is mutex-protected so commands execute serially within a single instance.

## Automatic environment

Every shell instance injects these variables:

| Variable | Value |
|----------|-------|
| `CRUSH` | `1` |
| `AGENT` | `crush` |
| `AI_AGENT` | `crush` |

These let scripts and tools detect they are running inside a Crush [agent](./agents.md) session.

## Command blocking

The shell supports pluggable block functions (`BlockFunc`) that inspect the command line before execution:

| Blocker | What it blocks |
|---------|---------------|
| `CommandsBlocker` | Specific top-level commands (e.g., `ssh`, `chrome`, `firefox`, `lynx`). |
| `ArgumentsBlocker` | Specific subcommands, flags, or arguments on a given command (e.g., `npm install -g`, `pip install --user`). |

When a blocked command is detected, execution is rejected before the process starts. The [bash tool](./tools.md) configures blockers for 23 interactive commands and 9 package manager patterns.

## Background jobs

The `BackgroundShellManager` (`internal/shell/background.go`) manages long-running processes:

| Aspect | Detail |
|--------|--------|
| Limit | Maximum 50 concurrent background jobs. |
| Retention | Completed jobs are kept for 8 hours before automatic cleanup. |
| Output | Stdout and stderr are captured in thread-safe ring buffers. |
| Lifecycle | Jobs are started with `Start()`, monitored with `Get()`, and terminated with `Kill()`. |
| Identification | Each job gets a unique `shell_id` used by the [job_output](./tools.md) and [job_kill](./tools.md) tools. |

The [bash tool](./tools.md) supports an `auto_background_after` parameter (default 60 seconds). If a foreground command runs longer than this threshold, it is automatically promoted to a background job and its `shell_id` is returned.

## Cross-platform notes

On Windows, `mvdan.cc/sh/moreinterp/coreutils` provides Unix utilities (cat, grep, etc.) so POSIX commands work without WSL or Git Bash. Path separators and line endings are normalized by the interpreter.

## Related

- [Tools](./tools.md) – `bash`, `job_output`, `job_kill`, and `tmux` tools use the shell.
- [Agents](./agents.md) – agents execute commands through tools that delegate to the shell.
- [Permissions](./permissions.md) – command blocking is a complementary security layer.
