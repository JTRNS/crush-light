package filepathext

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmartIsAbs(t *testing.T) {
	t.Parallel()

	rooted := filepath.Join(string(os.PathSeparator), "tmp", "child.txt")

	require.True(t, SmartIsAbs(rooted))
	require.False(t, SmartIsAbs(filepath.Join("tmp", "child.txt")))
}

func TestSmartJoin(t *testing.T) {
	t.Parallel()

	base := filepath.Join("home", "user")
	rooted := filepath.Join(string(os.PathSeparator), "tmp", "child.txt")

	require.Equal(t, filepath.Join(base, "child.txt"), SmartJoin(base, "child.txt"))
	require.Equal(t, rooted, SmartJoin(base, rooted))
}
