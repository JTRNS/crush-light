package stringext

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCapitalize(t *testing.T) {
	t.Parallel()

	require.Equal(t, "Hello World", Capitalize("hello world"))
}

func TestNormalizeSpace(t *testing.T) {
	t.Parallel()

	content := " \r\n\tfirst line\r\nsecond\tline\t\r\n "

	require.Equal(t, "first line\nsecond    line", NormalizeSpace(content))
}
