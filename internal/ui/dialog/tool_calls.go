package dialog

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/styles"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/sahilm/fuzzy"
)

const (
	// ToolCallsID is the identifier for the tool calls dialog.
	ToolCallsID = "tool_calls"
)

// ToolCalls displays tool calls for the current session.
type ToolCalls struct {
	com   *common.Common
	help  help.Model
	list  *list.FilterableList
	input textinput.Model

	keyMap struct {
		Select   key.Binding
		Next     key.Binding
		Previous key.Binding
		UpDown   key.Binding
		Close    key.Binding
	}
}

// ToolCallItem is a condensed, filterable row for a single tool call.
type ToolCallItem struct {
	id         string
	name       string
	input      string
	createdAt  int64
	isFinished bool
	position   int
	t          *styles.Styles
	m          fuzzy.Match
	cache      map[int]string
	focused    bool
}

var (
	_ Dialog   = (*ToolCalls)(nil)
	_ ListItem = (*ToolCallItem)(nil)
)

// NewToolCalls creates a new tool calls dialog.
func NewToolCalls(com *common.Common, messages []message.Message) (*ToolCalls, error) {
	tc := &ToolCalls{com: com}

	helpModel := help.New()
	helpModel.Styles = com.Styles.DialogHelpStyles()
	tc.help = helpModel

	tc.list = list.NewFilterableList()
	tc.list.Focus()
	tc.list.SetSelected(0)

	tc.input = textinput.New()
	tc.input.SetVirtualCursor(false)
	tc.input.Placeholder = "Type to filter"
	tc.input.SetStyles(com.Styles.TextInput)
	tc.input.Focus()

	tc.keyMap.Select = key.NewBinding(
		key.WithKeys("enter", "ctrl+y"),
		key.WithHelp("enter", "open in chat"),
	)
	tc.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓", "next item"),
	)
	tc.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑", "previous item"),
	)
	tc.keyMap.UpDown = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "choose"),
	)
	tc.keyMap.Close = CloseKey

	items := toolCallItems(com.Styles, messages)
	tc.list.SetItems(items...)
	tc.list.SetSelected(0)
	tc.list.ScrollToTop()

	return tc, nil
}

// ID implements Dialog.
func (t *ToolCalls) ID() string {
	return ToolCallsID
}

// HandleMsg implements [Dialog].
func (t *ToolCalls) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, t.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, t.keyMap.Select):
			if selectedItem, ok := t.list.SelectedItem().(*ToolCallItem); ok && selectedItem.position > 0 {
				return ActionOpenToolCall{ToolCallID: selectedItem.id}
			}
			return ActionClose{}
		case key.Matches(msg, t.keyMap.Previous):
			t.list.Focus()
			if t.list.IsSelectedFirst() {
				t.list.SelectLast()
			} else {
				t.list.SelectPrev()
			}
			t.list.ScrollToSelected()
		case key.Matches(msg, t.keyMap.Next):
			t.list.Focus()
			if t.list.IsSelectedLast() {
				t.list.SelectFirst()
			} else {
				t.list.SelectNext()
			}
			t.list.ScrollToSelected()
		default:
			var cmd tea.Cmd
			t.input, cmd = t.input.Update(msg)
			value := t.input.Value()
			t.list.SetFilter(value)
			t.list.ScrollToTop()
			t.list.SetSelected(0)
			return ActionCmd{cmd}
		}
	}
	return nil
}

// Cursor returns the cursor position relative to the dialog.
func (t *ToolCalls) Cursor() *tea.Cursor {
	return InputCursor(t.com.Styles, t.input.Cursor())
}

// Draw implements [Dialog].
func (t *ToolCalls) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	sty := t.com.Styles
	width := max(0, min(defaultDialogMaxWidth, area.Dx()-sty.Dialog.View.GetHorizontalBorderSize()))
	height := max(0, min(defaultDialogHeight, area.Dy()-sty.Dialog.View.GetVerticalBorderSize()))
	innerWidth := width - sty.Dialog.View.GetHorizontalFrameSize()
	heightOffset := sty.Dialog.Title.GetVerticalFrameSize() + titleContentHeight +
		sty.Dialog.InputPrompt.GetVerticalFrameSize() + inputContentHeight +
		sty.Dialog.HelpView.GetVerticalFrameSize() +
		sty.Dialog.View.GetVerticalFrameSize()

	t.input.SetWidth(max(0, innerWidth-sty.Dialog.InputPrompt.GetHorizontalFrameSize()-1))
	t.list.SetSize(innerWidth, height-heightOffset)
	t.help.SetWidth(innerWidth)

	rc := NewRenderContext(sty, width)
	rc.Title = "Tool Calls"
	inputView := sty.Dialog.InputPrompt.Render(t.input.View())
	rc.AddPart(inputView)
	listView := sty.Dialog.List.Height(t.list.Height()).Render(t.list.Render())
	rc.AddPart(listView)
	rc.Help = t.help.View(t)

	view := rc.Render()
	cur := t.Cursor()
	DrawCenterCursor(scr, area, view, cur)
	return cur
}

// ShortHelp implements [help.KeyMap].
func (t *ToolCalls) ShortHelp() []key.Binding {
	return []key.Binding{
		t.keyMap.UpDown,
		t.keyMap.Select,
		t.keyMap.Close,
	}
}

// FullHelp implements [help.KeyMap].
func (t *ToolCalls) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{t.keyMap.Select, t.keyMap.Next, t.keyMap.Previous},
		{t.keyMap.Close},
	}
}

func toolCallItems(sty *styles.Styles, messages []message.Message) []list.FilterableItem {
	items := []list.FilterableItem{}
	position := 1
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		calls := msg.ToolCalls()
		for j := len(calls) - 1; j >= 0; j-- {
			call := calls[j]
			items = append(items, &ToolCallItem{
				id:         call.ID,
				name:       call.Name,
				input:      call.Input,
				createdAt:  msg.CreatedAt,
				isFinished: call.Finished,
				position:   position,
				t:          sty,
			})
			position++
		}
	}
	if len(items) == 0 {
		items = append(items, &ToolCallItem{
			id:       "none",
			name:     "No tool calls",
			input:    "",
			createdAt: 0,
			position: 0,
			t:        sty,
		})
	}
	return items
}

// Filter returns a filterable summary string.
func (t *ToolCallItem) Filter() string {
	return fmt.Sprintf("%s %s", t.name, t.input)
}

// ID returns the unique identifier of the item.
func (t *ToolCallItem) ID() string {
	return t.id
}

// SetFocused sets focused state.
func (t *ToolCallItem) SetFocused(focused bool) {
	if t.focused != focused {
		t.cache = nil
	}
	t.focused = focused
}

// SetMatch sets fuzzy match data.
func (t *ToolCallItem) SetMatch(m fuzzy.Match) {
	t.cache = nil
	t.m = m
}

// Render renders the condensed row.
func (t *ToolCallItem) Render(width int) string {
	title := t.name
	if t.input != "" {
		title = fmt.Sprintf("%s  %s", t.name, t.input)
	}
	info := ""
	if t.position > 0 {
		status := "pending"
		if t.isFinished {
			status = "OK"
		}
		if t.createdAt > 0 {
			info = fmt.Sprintf("#%d · %s · %s", t.position, status, formatToolCallTime(time.Unix(t.createdAt, 0)))
		} else {
			info = fmt.Sprintf("#%d · %s", t.position, status)
		}
	}
	styles := ListItemStyles{
		ItemBlurred:     t.t.Dialog.NormalItem,
		ItemFocused:     t.t.Dialog.SelectedItem,
		InfoTextBlurred: t.t.Subtle,
		InfoTextFocused: t.t.Base,
	}
	return renderItem(styles, title, info, t.focused, width, t.cache, &t.m)
}

func formatToolCallTime(ts time.Time) string {
	now := time.Now()
	if ts.Year() == now.Year() && ts.YearDay() == now.YearDay() {
		return ts.Format("15:04:05")
	}
	days := int(now.Sub(ts).Hours() / 24)
	days = max(days, 1)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
