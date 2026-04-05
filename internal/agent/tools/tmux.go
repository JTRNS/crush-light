package tools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/shell"
)

const TmuxToolName = "tmux"

//go:embed tmux.md
var tmuxDescription []byte

type TmuxParams struct {
	Operation   string            `json:"operation" description:"Operation to perform: create, send_keys, capture, list_sessions, list_panes, kill_session"`
	SessionName string            `json:"session_name,omitempty" description:"Name of the tmux session"`
	Target      string            `json:"target,omitempty" description:"Target pane or window (e.g., '0', '%0', '1.%0')"`
	Keys        string            `json:"keys,omitempty" description:"Keys to send (use this for send_keys)."`
	Command     string            `json:"command,omitempty" description:"Command to run (used with create operation)"`
	WorkingDir  string            `json:"working_dir,omitempty" description:"Working directory for the command. Defaults to current directory."`
	PaneCount   int               `json:"pane_count,omitempty" description:"Number of initial panes to create horizontally (used with create). Default: 1."`
	ExtraArgs   map[string]string `json:"extra_args,omitempty" description:"Additional key-value pairs for specialized use. Use commandN for commands to run in additional panes (e.g., command1, command2)."`
}

func quoteShell(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func formatError(action, cmd string, execErr error, stderr *string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "tmux %s failed", action)
	if cmd != "" {
		fmt.Fprintf(&b, " (command: %q)", cmd)
	}
	b.WriteString("\n")
	if stderr != nil {
		fmt.Fprintf(&b, "stderr: %s\n", *stderr)
	}
	if execErr != nil {
		fmt.Fprintf(&b, "error: %s", execErr)
	}
	return b.String()
}

func tmuxExec(ctx context.Context, cmd string) (string, error) {
	sh := shell.NewShell(&shell.Options{
		WorkingDir: "/tmp",
	})
	out, errStr, err := sh.Exec(ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("%s", formatError("", cmd, err, &errStr))
	}
	return out, nil
}

func NewTmuxTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		TmuxToolName,
		string(tmuxDescription),
		func(ctx context.Context, params TmuxParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Operation == "" {
				return fantasy.NewTextErrorResponse("missing required field: operation"), nil
			}

			workingDir := params.WorkingDir
			if workingDir == "" {
				workingDir = "/tmp"
			}

			switch params.Operation {
			case "create":
				if params.SessionName == "" {
					return fantasy.NewTextErrorResponse("create requires session_name"), nil
				}
				if params.Command == "" {
					return fantasy.NewTextErrorResponse("create requires command"), nil
				}
				// Kill existing session with same name if it exists (ignore errors).
				tmuxExec(ctx, fmt.Sprintf("tmux kill-session -t %s 2>/dev/null", quoteShell(params.SessionName)))

				paneCount := params.PaneCount
				if paneCount <= 0 {
					paneCount = 1
				}

				var cmd string
				if paneCount == 1 {
					cmd = fmt.Sprintf("tmux new-session -d -s %s -c %s %s", quoteShell(params.SessionName), quoteShell(workingDir), quoteShell(params.Command))
				} else {
					cmd = fmt.Sprintf("tmux new-session -d -s %s -c %s %s", quoteShell(params.SessionName), quoteShell(workingDir), quoteShell("sleep 0"))
					for i := 1; i < paneCount; i++ {
						cmd += fmt.Sprintf(" && tmux split-window -h -t %s", quoteShell(params.SessionName+":0"))
					}
					cmd += fmt.Sprintf(" && tmux select-layout -t %s even-horizontal",
						quoteShell(params.SessionName+":0"))
					// Now send commands to each pane.
					if len(params.ExtraArgs) > 0 {
						for i := 0; i < paneCount; i++ {
							var paneCmd string
							if i == 0 {
								paneCmd = params.Command
							} else {
								key := fmt.Sprintf("command%d", i)
								paneCmd = params.ExtraArgs[key]
							}
							if paneCmd != "" {
								escaped := strings.ReplaceAll(paneCmd, "'", "'\\''")
								cmd += fmt.Sprintf(" && tmux send-keys -t %s '%s' Enter", quoteShell(fmt.Sprintf("%s:0.%d", params.SessionName, i)), escaped)
							}
						}
					} else {
						// Send the main command to the first pane.
						escaped := strings.ReplaceAll(params.Command, "'", "'\\''")
						cmd += fmt.Sprintf(" && tmux send-keys -t %s 'sleep 0 && %s' Enter", quoteShell(params.SessionName+":0"), escaped)
					}
				}

				_, err := tmuxExec(ctx, cmd)
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to create session %q: %v", params.SessionName, err)), nil
				}

				// Get pane info.
				paneInfo, _ := tmuxExec(ctx, fmt.Sprintf("tmux list-panes -t %s -F '#{pane_id} #{pane_index} #{pane_current_command}' 2>/dev/null", quoteShell(params.SessionName)))

				var result strings.Builder
				fmt.Fprintf(&result, "tmux session %q created successfully", params.SessionName)
				if workingDir != "" {
					fmt.Fprintf(&result, " with working directory %q", workingDir)
				}
				result.WriteString("\n")
				if paneInfo != "" {
					result.WriteString("\nActive panes:\n")
					result.WriteString(paneInfo)
				}
				result.WriteString("\n\nUse send_keys to interact with the session, capture to read output.")
				return fantasy.NewTextResponse(result.String()), nil

			case "send_keys":
				if params.SessionName == "" {
					return fantasy.NewTextErrorResponse("send_keys requires session_name"), nil
				}
				if params.Keys == "" {
					return fantasy.NewTextErrorResponse("send_keys requires keys"), nil
				}

				escapedKeys := strings.ReplaceAll(params.Keys, "'", "'\\''")
				target := fmt.Sprintf("%s", params.SessionName)
				if params.Target != "" {
					target = fmt.Sprintf("%s:%s", params.SessionName, params.Target)
				}

				cmd := fmt.Sprintf("tmux send-keys -t %s '%s'", quoteShell(target), escapedKeys)
				out, err := tmuxExec(ctx, cmd)
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to send keys to session %q: %v (%s)", params.SessionName, err, out)), nil
				}

				return fantasy.NewTextResponse(fmt.Sprintf("Keys sent to session %q (target: %s)\n\nUse capture to read the output.",
					params.SessionName, target)), nil

			case "capture":
				if params.SessionName == "" {
					return fantasy.NewTextErrorResponse("capture requires session_name"), nil
				}

				target := fmt.Sprintf("%s", params.SessionName)
				if params.Target != "" {
					target = fmt.Sprintf("%s:%s", params.SessionName, params.Target)
				}

				lines := params.ExtraArgs["lines"]
				if lines == "" {
					lines = "40"
				}

				cmd := fmt.Sprintf("tmux capture-pane -t %s -p -S -%s 2>/dev/null", quoteShell(target), lines)
				out, err := tmuxExec(ctx, cmd)
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to capture session %q: %v", params.SessionName, err)), nil
				}

				if strings.TrimSpace(out) == "" {
					out = "(empty pane - may need more time for output to appear)"
				}

				return fantasy.NewTextResponse(fmt.Sprintf("Captured output from session %q (target: %s):\n\n```\n%s\n```",
					params.SessionName, target, out)), nil

			case "list_sessions":
				out, err := tmuxExec(ctx, "tmux list-sessions -F '#{session_name}: #{session_windows} windows, #{session_created} created' 2>/dev/null")
				if err != nil {
					return fantasy.NewTextErrorResponse("no active tmux sessions found"), nil
				}
				return fantasy.NewTextResponse("Active tmux sessions:\n\n" + out), nil

			case "list_panes":
				if params.SessionName == "" {
					return fantasy.NewTextErrorResponse("list_panes requires session_name"), nil
				}

				out, err := tmuxExec(ctx, fmt.Sprintf("tmux list-panes -t %s -F 'pane #{pane_index} (id: #{pane_id}): #{pane_current_command}' 2>/dev/null", quoteShell(params.SessionName)))
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("session %q not found or has no panes", params.SessionName)), nil
				}
				return fantasy.NewTextResponse(fmt.Sprintf("Panes in session %q:\n\n%s", params.SessionName, out)), nil

			case "kill_session":
				if params.SessionName == "" {
					return fantasy.NewTextErrorResponse("kill_session requires session_name"), nil
				}

				out, err := tmuxExec(ctx, fmt.Sprintf("tmux kill-session -t %s 2>/dev/null", quoteShell(params.SessionName)))
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to kill session %q: %v (%s)", params.SessionName, err, out)), nil
				}
				return fantasy.NewTextResponse(fmt.Sprintf("tmux session %q terminated", params.SessionName)), nil

			default:
				return fantasy.NewTextErrorResponse(fmt.Sprintf("unknown operation: %q. Valid operations: create, send_keys, capture, list_sessions, list_panes, kill_session",
					params.Operation)), nil
			}
		})
}
