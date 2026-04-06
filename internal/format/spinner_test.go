package format

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/ui/spinner"
	"github.com/stretchr/testify/require"
)

func TestNewSpinner(t *testing.T) {
	t.Parallel()

	s := NewSpinner(context.Background(), func() {}, spinner.Settings{ID: "spinner-test"})

	require.NotNil(t, s)
	require.NotNil(t, s.prog)
	require.NotNil(t, s.done)
}

func TestModelUpdate_QuitKeysCancelAndQuit(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{name: "ctrl+c", key: tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}},
		{name: "esc", key: tea.KeyPressMsg{Code: tea.KeyEsc}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cancelled := false
			m := model{
				cancel:  func() { cancelled = true },
				spinner: spinner.New(spinner.Settings{ID: "spinner-test"}),
			}

			updated, cmd := m.Update(tc.key)

			require.True(t, cancelled)
			require.IsType(t, model{}, updated)
			require.NotNil(t, cmd)
			require.IsType(t, tea.QuitMsg{}, cmd())
		})
	}
}

func TestModelUpdate_NonQuitKeyDoesNothing(t *testing.T) {
	t.Parallel()

	cancelled := false
	m := model{
		cancel:  func() { cancelled = true },
		spinner: spinner.New(spinner.Settings{ID: "spinner-test"}),
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})

	require.False(t, cancelled)
	require.IsType(t, model{}, updated)
	require.Nil(t, cmd)
}
