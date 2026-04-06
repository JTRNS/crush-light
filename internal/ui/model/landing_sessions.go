package model

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/dialog"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/util"
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

type landingSessionsMode uint8

const (
	landingSessionsModeNormal   landingSessionsMode = iota
	landingSessionsModeDeleting                     // awaiting y/n delete confirmation
	landingSessionsModeUpdating                     // inline rename in progress
)

// landingSessions is an inline, keyboard-navigable session list embedded in
// the landing view main area.
type landingSessions struct {
	com      *common.Common
	list     *list.FilterableList
	sessions []session.Session
	mode     landingSessionsMode
	keyMap   struct {
		Next          key.Binding
		Previous      key.Binding
		Select        key.Binding
		Rename        key.Binding
		Delete        key.Binding
		ConfirmRename key.Binding
		CancelRename  key.Binding
		ConfirmDelete key.Binding
		CancelDelete  key.Binding
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
	l.keyMap.Rename = key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "rename"),
	)
	l.keyMap.Delete = key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("ctrl+x", "delete"),
	)
	l.keyMap.ConfirmRename = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	)
	l.keyMap.CancelRename = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	)
	l.keyMap.ConfirmDelete = key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "confirm"),
	)
	l.keyMap.CancelDelete = key.NewBinding(
		key.WithKeys("n", "esc"),
		key.WithHelp("n", "cancel"),
	)
	return l
}

// Len returns the number of sessions currently in the list.
func (l *landingSessions) Len() int {
	return l.list.Len()
}

// SetSessions replaces the list contents with the provided sessions.
func (l *landingSessions) SetSessions(sessions []session.Session) {
	l.sessions = sessions
	l.mode = landingSessionsModeNormal
	l.rebuildItems()
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

// Update handles a key press. It returns a tea.Cmd when an async operation
// (session load, rename, delete) needs to be executed.
func (l *landingSessions) Update(msg tea.KeyPressMsg) tea.Cmd {
	switch l.mode {
	case landingSessionsModeDeleting:
		switch {
		case key.Matches(msg, l.keyMap.ConfirmDelete):
			return l.confirmDelete()
		case key.Matches(msg, l.keyMap.CancelDelete):
			l.mode = landingSessionsModeNormal
			l.rebuildItems()
		}

	case landingSessionsModeUpdating:
		switch {
		case key.Matches(msg, l.keyMap.ConfirmRename):
			return l.confirmRename()
		case key.Matches(msg, l.keyMap.CancelRename):
			l.mode = landingSessionsModeNormal
			l.rebuildItems()
		default:
			// Forward all other keys to the selected item's inline text input.
			if item := l.list.SelectedItem(); item != nil {
				if si, ok := item.(*dialog.SessionItem); ok {
					return si.HandleInput(msg)
				}
			}
		}

	default: // landingSessionsModeNormal
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
		case key.Matches(msg, l.keyMap.Rename):
			l.mode = landingSessionsModeUpdating
			l.rebuildItems()
		case key.Matches(msg, l.keyMap.Delete):
			l.mode = landingSessionsModeDeleting
			l.rebuildItems()
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

	var header string
	switch l.mode {
	case landingSessionsModeDeleting:
		header = t.Dialog.Sessions.DeletingMessage.Width(width).Render("Delete this session? (y/n)")
	case landingSessionsModeUpdating:
		header = t.Dialog.Sessions.RenamingingMessage.Width(width).Render("Rename session (enter to confirm, esc to cancel)")
	default:
		header = t.Muted.Width(width).Render("Recent sessions")
	}

	headerHeight := lipgloss.Height(header) + 1 // +1 for blank separator line.
	l.list.SetSize(width, max(0, height-headerHeight))
	return lipgloss.JoinVertical(lipgloss.Left, header, "", l.list.Render())
}

// rebuildItems refreshes the list items to match the current mode and sessions.
func (l *landingSessions) rebuildItems() {
	selected := l.list.Selected()
	var items []list.FilterableItem
	switch l.mode {
	case landingSessionsModeDeleting:
		items = dialog.NewSessionItemsDeleting(l.com.Styles, l.sessions...)
	case landingSessionsModeUpdating:
		items = dialog.NewSessionItemsUpdating(l.com.Styles, l.sessions...)
	default:
		items = dialog.NewSessionItems(l.com.Styles, l.sessions...)
	}
	l.list.SetItems(items...)
	if selected >= 0 && selected < l.list.Len() {
		l.list.SetSelected(selected)
		l.list.ScrollToSelected()
	}
}

// confirmDelete removes the selected session and dispatches the delete command.
func (l *landingSessions) confirmDelete() tea.Cmd {
	item := l.list.SelectedItem()
	si, ok := item.(*dialog.SessionItem)
	if !ok {
		return nil
	}
	id := si.ID()
	l.mode = landingSessionsModeNormal
	l.removeSession(id)
	l.rebuildItems()
	return func() tea.Msg {
		if err := l.com.Workspace.DeleteSession(context.TODO(), id); err != nil {
			return util.NewErrorMsg(err)
		}
		return nil
	}
}

// confirmRename applies the new title and dispatches the save command.
func (l *landingSessions) confirmRename() tea.Cmd {
	item := l.list.SelectedItem()
	si, ok := item.(*dialog.SessionItem)
	if !ok {
		return nil
	}
	newTitle := strings.TrimSpace(si.InputValue())
	if newTitle == "" {
		l.mode = landingSessionsModeNormal
		l.rebuildItems()
		return nil
	}
	id := si.ID()
	l.mode = landingSessionsModeNormal
	// Update the local copy so the list re-renders immediately.
	for i, s := range l.sessions {
		if s.ID == id {
			l.sessions[i].Title = newTitle
			break
		}
	}
	l.rebuildItems()
	sess := si.Session
	sess.Title = newTitle
	return func() tea.Msg {
		if _, err := l.com.Workspace.SaveSession(context.TODO(), sess); err != nil {
			return util.NewErrorMsg(err)
		}
		return nil
	}
}

// removeSession removes the session with the given id from the local slice.
func (l *landingSessions) removeSession(id string) {
	filtered := l.sessions[:0]
	for _, s := range l.sessions {
		if s.ID != id {
			filtered = append(filtered, s)
		}
	}
	l.sessions = filtered
}
