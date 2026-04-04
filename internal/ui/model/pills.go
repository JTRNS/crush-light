package model

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/ui/styles"
)

// pillStyle returns the appropriate style for a pill based on focus state.
func pillStyle(focused, panelFocused bool, t *styles.Styles) lipgloss.Style {
	if !panelFocused || focused {
		return t.Pills.Focused
	}
	return t.Pills.Blurred
}

const (
	// pillHeightWithBorder is the height of a pill including its border.
	pillHeightWithBorder = 3
	// maxQueueDisplayLength is the maximum length of a queue item in the list.
	maxQueueDisplayLength = 60
)

// pillSection represents which section of the pills panel is focused.
type pillSection int

const (
	pillSectionQueue pillSection = iota
)

// queuePill renders the queue count pill with gradient triangles.
func queuePill(queue int, focused, panelFocused bool, t *styles.Styles) string {
	if queue <= 0 {
		return ""
	}
	triangles := styles.ForegroundGrad(t, "▶▶▶▶▶▶▶▶▶", false, t.RedDark, t.Secondary)
	if queue < len(triangles) {
		triangles = triangles[:queue]
	}

	text := t.Base.Render(fmt.Sprintf("%d Queued", queue))
	content := fmt.Sprintf("%s %s", strings.Join(triangles, ""), text)
	return pillStyle(focused, panelFocused, t).Render(content)
}

// queueList renders the expanded queue items list.
func queueList(queueItems []string, t *styles.Styles) string {
	if len(queueItems) == 0 {
		return ""
	}

	var lines []string
	for _, item := range queueItems {
		text := item
		if len(text) > maxQueueDisplayLength {
			text = text[:maxQueueDisplayLength-1] + "…"
		}
		prefix := t.Pills.QueueItemPrefix.Render() + " "
		lines = append(lines, prefix+t.Muted.Render(text))
	}

	return strings.Join(lines, "\n")
}

// togglePillsExpanded toggles the pills panel expansion state.
func (m *UI) togglePillsExpanded() tea.Cmd {
	if !m.hasSession() {
		return nil
	}
	if m.promptQueue <= 0 {
		return nil
	}
	m.pillsExpanded = !m.pillsExpanded
	if m.pillsExpanded {
		m.focusedPillSection = pillSectionQueue
	}
	m.updateLayoutAndSize()

	// Make sure to follow scroll if follow is enabled when toggling pills.
	if m.chat.Follow() {
		m.chat.ScrollToBottom()
	}

	return nil
}

// switchPillSection is a no-op since there is only one section now.
func (m *UI) switchPillSection(_ int) tea.Cmd {
	return nil
}

// pillsAreaHeight calculates the total height needed for the pills area.
func (m *UI) pillsAreaHeight() int {
	if !m.hasSession() {
		return 0
	}
	hasQueue := m.promptQueue > 0
	if !hasQueue {
		return 0
	}

	pillsAreaHeight := pillHeightWithBorder
	if m.pillsExpanded && hasQueue {
		pillsAreaHeight += m.promptQueue
	}
	return pillsAreaHeight
}

// renderPills renders the pills panel and stores it in m.pillsView.
func (m *UI) renderPills() {
	m.pillsView = ""
	if !m.hasSession() {
		return
	}

	width := m.layout.pills.Dx()
	if width <= 0 {
		return
	}

	paddingLeft := 3
	hasQueue := m.promptQueue > 0

	if !hasQueue {
		return
	}

	t := m.com.Styles
	queueFocused := m.pillsExpanded && m.focusedPillSection == pillSectionQueue

	var pills []string
	if hasQueue {
		pills = append(pills, queuePill(m.promptQueue, queueFocused, m.pillsExpanded, t))
	}

	var expandedList string
	if m.pillsExpanded && queueFocused && hasQueue {
		if m.com.Workspace.AgentIsReady() {
			queueItems := m.com.Workspace.AgentQueuedPromptsList(m.session.ID)
			expandedList = queueList(queueItems, t)
		}
	}

	if len(pills) == 0 {
		return
	}

	pillsRow := lipgloss.JoinHorizontal(lipgloss.Top, pills...)

	helpDesc := "open"
	if m.pillsExpanded {
		helpDesc = "close"
	}
	helpKey := t.Pills.HelpKey.Render("ctrl+t")
	helpText := t.Pills.HelpText.Render(helpDesc)
	helpHint := lipgloss.JoinHorizontal(lipgloss.Center, helpKey, " ", helpText)
	pillsRow = lipgloss.JoinHorizontal(lipgloss.Center, pillsRow, " ", helpHint)

	pillsArea := pillsRow
	if expandedList != "" {
		pillsArea = lipgloss.JoinVertical(lipgloss.Left, pillsRow, expandedList)
	}

	m.pillsView = t.Pills.Area.MaxWidth(width).PaddingLeft(paddingLeft).Render(pillsArea)
}
