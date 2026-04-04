package diff

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateDiff_TrimmedFilenameAndChangeCounts(t *testing.T) {
	t.Parallel()

	diff, additions, removals := GenerateDiff(
		"alpha\nbeta\n",
		"alpha\ngamma\n",
		"/tmp/example.txt",
	)

	require.Contains(t, diff, "--- a/tmp/example.txt")
	require.Contains(t, diff, "+++ b/tmp/example.txt")
	require.Contains(t, diff, "-beta")
	require.Contains(t, diff, "+gamma")
	require.Equal(t, 1, additions)
	require.Equal(t, 1, removals)
}

func TestGenerateDiff_AddOnlyFile(t *testing.T) {
	t.Parallel()

	diff, additions, removals := GenerateDiff("", "one\ntwo\n", "notes.txt")

	require.Contains(t, diff, "--- a/notes.txt")
	require.Contains(t, diff, "+++ b/notes.txt")
	require.Contains(t, diff, "+one")
	require.Contains(t, diff, "+two")
	require.Equal(t, 2, additions)
	require.Zero(t, removals)
}

func TestGenerateDiff_DeleteOnlyFile(t *testing.T) {
	t.Parallel()

	diff, additions, removals := GenerateDiff("one\ntwo\n", "", "notes.txt")

	require.Contains(t, diff, "--- a/notes.txt")
	require.Contains(t, diff, "+++ b/notes.txt")
	require.Contains(t, diff, "-one")
	require.Contains(t, diff, "-two")
	require.Zero(t, additions)
	require.Equal(t, 2, removals)
}
