package model

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/dialog"
	"github.com/charmbracelet/crush/internal/ui/list"
)

// selectSessionMsg is emitted when the user selects a session from the landing
// sessions list.
type selectSessionMsg struct {
	sessionID string
}

// landingSessionsLoadedMsg is sent when sessions have been fetched for the
// landing view list.
type landingSessionsLoadedMsg struct {
	sessions []session.Session
}

// landingSessions is an inline, keyboard-navigable session list embedded in
// the landing view main area.
type landingSessions struct {
	com    *common.Common
	list   *list.FilterableList
	keyMap struct {
		Next     key.Binding
		Previous key.Binding
		Select   key.Binding
	}
}

func newLandingSessions(com *common.Common) *landingSessions {
	l := &landingSessions{com: com}
	l.list = list.NewFilterableList()
	l.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n", "j"),
		key.WithHelp("↓", "next"),
	)
	l.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p", "k"),
		key.WithHelp("↑", "previous"),
	)
	l.keyMap.Select = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "resume"),
	)
	return l
}

// Len returns the number of sessions currently in the list.
func (l *landingSessions) Len() int {
	return l.list.Len()
}

// SetSessions replaces the list contents with the provided sessions.
func (l *landingSessions) SetSessions(sessions []session.Session) {
	items := dialog.NewSessionItems(l.com.Styles, sessions...)
	l.list.SetItems(items...)
	if l.list.Len() > 0 {
		l.list.SelectFirst()
		l.list.ScrollToTop()
	}
}

// Focus gives the list keyboard focus so items render in their focused style.
func (l *landingSessions) Focus() {
	l.list.Focus()
}

// Blur removes keyboard focus from the list.
func (l *landingSessions) Blur() {
	l.list.Blur()
}

// Update handles a key press. It returns a tea.Cmd that emits a
// selectSessionMsg when the user confirms a selection.
func (l *landingSessions) Update(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, l.keyMap.Previous):
		if l.list.IsSelectedFirst() {
			l.list.SelectLast()
		} else {
			l.list.SelectPrev()
		}
		l.list.ScrollToSelected()
	case key.Matches(msg, l.keyMap.Next):
		if l.list.IsSelectedLast() {
			l.list.SelectFirst()
		} else {
			l.list.SelectNext()
		}
		l.list.ScrollToSelected()
	case key.Matches(msg, l.keyMap.Select):
		if item := l.list.SelectedItem(); item != nil {
			if si, ok := item.(*dialog.SessionItem); ok {
				id := si.ID()
				return func() tea.Msg {
					return selectSessionMsg{sessionID: id}
				}
			}
		}
	}
	return nil
}

// View renders the session list into the given width and height.
func (l *landingSessions) View(width, height int) string {
	t := l.com.Styles
	if l.list.Len() == 0 {
		return t.Muted.Width(width).PaddingTop(1).Render("No previous sessions")
	}
	header := t.Muted.Width(width).Render("Recent sessions")
	headerHeight := lipgloss.Height(header) + 1 // +1 for blank separator line.
	l.list.SetSize(width, max(0, height-headerHeight))
	return lipgloss.JoinVertical(lipgloss.Left, header, "", l.list.Render())
}
