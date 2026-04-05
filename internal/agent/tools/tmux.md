Manages tmux terminal sessions for running long-lived or interactive processes without blocking.

<purpose>
Prevents blocking on dev servers, watches, tail -f, TUI apps, interactive CLIs, or any program that never exits.
The agent should ALWAYS use this tool for such commands instead of the bash tool.
</purpose>

<concepts>
- Session: A named tmux instance that can contain multiple windows/panes
- Pane: A terminal within a session that runs its own process
- Detached: The session runs in the background; the agent controls it via commands
</concepts>

<operations>
- create: Create a new detached session and run a command inside it
- send_keys: Send keystrokes to a specific pane
- capture: Capture the visible output from a pane
- list_sessions: List all active sessions
- list_panes: List panes within a session
- kill_session: Terminate an entire session and all its processes
</operations>

<guidelines>
- Always prefer tmux over bash for commands that start a dev server, watcher, tail -f, or any TUI app
- Use descriptive session names (e.g., "dev-server", "test-watch", "logs")
- After send_keys, wait ~1-2 seconds before capture to let output appear
- When running a server, verify it started by capturing pane output
- Sessions persist independently — a created session can be found and reused later
</guidelines>

<warning>
NEVER use the bash tool to run commands like: npm start, npm run dev, python -m http.server, tail -f, htop, vim, or anything interactive.
Use this tool instead. The bash tool will block forever on non-terminating commands.
</warning>
