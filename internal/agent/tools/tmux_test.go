package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"charm.land/fantasy"
)

func runTmuxTool(t *testing.T, tool fantasy.AgentTool, ctx context.Context, params TmuxParams) fantasy.ToolResponse {
	t.Helper()

	input, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	call := fantasy.ToolCall{
		ID:    "test-call",
		Name:  TmuxToolName,
		Input: string(input),
	}

	resp, err := tool.Run(ctx, call)
	if err != nil {
		t.Fatalf("tool run error: %v", err)
	}
	return resp
}

func TestTmuxTool_CreateAndKill(t *testing.T) {
	t.Parallel()

	tool := NewTmuxTool()
	ctx := context.Background()

	// Create a session.
	createResp := runTmuxTool(t, tool, ctx, TmuxParams{
		Operation:   "create",
		SessionName: "tmux-test-create",
		Command:     "sleep 300",
	})

	text := createResp.Content
	if !strings.Contains(text, "created successfully") {
		t.Errorf("expected 'created successfully' in response, got: %s", text)
	}

	// Verify session exists via list.
	listResp := runTmuxTool(t, tool, ctx, TmuxParams{
		Operation: "list_sessions",
	})
	if !strings.Contains(listResp.Content, "tmux-test-create") {
		t.Errorf("expected 'tmux-test-create' in sessions list, got: %s", listResp.Content)
	}

	// Kill the session.
	killResp := runTmuxTool(t, tool, ctx, TmuxParams{
		Operation:   "kill_session",
		SessionName: "tmux-test-create",
	})
	if !strings.Contains(killResp.Content, "terminated") {
		t.Errorf("expected 'terminated' in response, got: %s", killResp.Content)
	}
}

func TestTmuxTool_Validation(t *testing.T) {
	t.Parallel()

	tool := NewTmuxTool()
	ctx := context.Background()

	tests := []struct {
		name   string
		params TmuxParams
		want   string
	}{
		{
			name:   "missing operation",
			params: TmuxParams{},
			want:   "missing required field: operation",
		},
		{
			name:   "create missing session_name",
			params: TmuxParams{Operation: "create", Command: "echo hi"},
			want:   "create requires session_name",
		},
		{
			name:   "create missing command",
			params: TmuxParams{Operation: "create", SessionName: "test"},
			want:   "create requires command",
		},
		{
			name:   "send_keys missing session_name",
			params: TmuxParams{Operation: "send_keys", Keys: "ls"},
			want:   "send_keys requires session_name",
		},
		{
			name:   "send_keys missing keys",
			params: TmuxParams{Operation: "send_keys", SessionName: "test"},
			want:   "send_keys requires keys",
		},
		{
			name:   "capture missing session_name",
			params: TmuxParams{Operation: "capture"},
			want:   "capture requires session_name",
		},
		{
			name:   "list_panes missing session_name",
			params: TmuxParams{Operation: "list_panes"},
			want:   "list_panes requires session_name",
		},
		{
			name:   "kill_session missing session_name",
			params: TmuxParams{Operation: "kill_session"},
			want:   "kill_session requires session_name",
		},
		{
			name:   "unknown operation",
			params: TmuxParams{Operation: "foobar"},
			want:   "unknown operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := runTmuxTool(t, tool, ctx, tt.params)
			text := resp.Content
			if !strings.Contains(text, tt.want) {
				t.Errorf("expected %q in response, got: %s", tt.want, text)
			}
		})
	}
}

func TestTmuxTool_SendKeysAndCapture(t *testing.T) {
	t.Parallel()

	tool := NewTmuxTool()
	ctx := context.Background()

	// Create a session.
	runTmuxTool(t, tool, ctx, TmuxParams{
		Operation:   "create",
		SessionName: "tmux-test-io",
		Command:     "sleep 300",
	})

	// send_keys should send only the provided keys.
	sendResp := runTmuxTool(t, tool, ctx, TmuxParams{
		Operation:   "send_keys",
		SessionName: "tmux-test-io",
		Keys:        "echo hello-no-enter",
	})
	if !strings.Contains(sendResp.Content, "Keys sent") {
		t.Errorf("expected 'Keys sent' in response, got: %s", sendResp.Content)
	}

	// Capture should show something.
	capResp := runTmuxTool(t, tool, ctx, TmuxParams{
		Operation:   "capture",
		SessionName: "tmux-test-io",
	})
	text := capResp.Content
	if !strings.Contains(text, "Captured output") {
		t.Errorf("expected 'Captured output' in response, got: %s", text)
	}

	// List panes should show pane info.
	paneResp := runTmuxTool(t, tool, ctx, TmuxParams{
		Operation:   "list_panes",
		SessionName: "tmux-test-io",
	})
	if !strings.Contains(paneResp.Content, "pane") {
		t.Errorf("expected 'pane' in response, got: %s", paneResp.Content)
	}

	// Clean up.
	runTmuxTool(t, tool, ctx, TmuxParams{
		Operation:   "kill_session",
		SessionName: "tmux-test-io",
	})
}
