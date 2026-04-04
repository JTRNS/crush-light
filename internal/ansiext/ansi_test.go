package ansiext

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestEscape_ReplacesControlCharacters(t *testing.T) {
	t.Parallel()

	input := "A\tB\n" + string(rune(ansi.DEL)) + "\x00"

	require.Equal(t, "A␉B␊␡␀", Escape(input))
}

func TestEscape_PreservesPrintableText(t *testing.T) {
	t.Parallel()

	require.Equal(t, "plain-é-text", Escape("plain-é-text"))
}
