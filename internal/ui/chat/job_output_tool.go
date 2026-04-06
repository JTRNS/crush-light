package chat

import (
	"cmp"
	"encoding/json"

	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/styles"
)

// -----------------------------------------------------------------------------
// Job Output Tool
// -----------------------------------------------------------------------------

// JobOutputToolMessageItem is a message item for job_output tool calls.
type JobOutputToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*JobOutputToolMessageItem)(nil)

// NewJobOutputToolMessageItem creates a new [JobOutputToolMessageItem].
func NewJobOutputToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &JobOutputToolRenderContext{}, canceled)
}

// JobOutputToolRenderContext renders job_output tool messages.
type JobOutputToolRenderContext struct{}

// RenderTool implements the [ToolRenderer] interface.
func (j *JobOutputToolRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	if opts.IsPending() {
		return pendingTool(sty, "Job", opts.Anim, opts.Compact)
	}

	var params tools.JobOutputParams
	if err := json.Unmarshal([]byte(opts.ToolCall.Input), &params); err != nil {
		return toolErrorContent(sty, &message.ToolResult{Content: "Invalid parameters"}, cappedWidth)
	}

	var description string
	if opts.HasResult() && opts.Result.Metadata != "" {
		var meta tools.JobOutputResponseMetadata
		if err := json.Unmarshal([]byte(opts.Result.Metadata), &meta); err == nil {
			description = cmp.Or(meta.Description, meta.Command)
		}
	}

	content := ""
	if opts.HasResult() {
		content = opts.Result.Content
	}
	return renderJobTool(sty, opts, cappedWidth, "Output", params.ShellID, description, content)
}
