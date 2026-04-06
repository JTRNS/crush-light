package chat

import (
	"cmp"
	"encoding/json"

	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/styles"
)

// -----------------------------------------------------------------------------
// Job Kill Tool
// -----------------------------------------------------------------------------

// JobKillToolMessageItem is a message item for job_kill tool calls.
type JobKillToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*JobKillToolMessageItem)(nil)

// NewJobKillToolMessageItem creates a new [JobKillToolMessageItem].
func NewJobKillToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &JobKillToolRenderContext{}, canceled)
}

// JobKillToolRenderContext renders job_kill tool messages.
type JobKillToolRenderContext struct{}

// RenderTool implements the [ToolRenderer] interface.
func (j *JobKillToolRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	if opts.IsPending() {
		return pendingTool(sty, "Job", opts.Anim, opts.Compact)
	}

	var params tools.JobKillParams
	if err := json.Unmarshal([]byte(opts.ToolCall.Input), &params); err != nil {
		return toolErrorContent(sty, &message.ToolResult{Content: "Invalid parameters"}, cappedWidth)
	}

	var description string
	if opts.HasResult() && opts.Result.Metadata != "" {
		var meta tools.JobKillResponseMetadata
		if err := json.Unmarshal([]byte(opts.Result.Metadata), &meta); err == nil {
			description = cmp.Or(meta.Description, meta.Command)
		}
	}

	content := ""
	if opts.HasResult() {
		content = opts.Result.Content
	}
	return renderJobTool(sty, opts, cappedWidth, "Kill", params.ShellID, description, content)
}
