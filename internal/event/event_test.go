package event

// These tests verify that the Error function correctly handles various
// scenarios. These tests will not log anything.

import (
	"reflect"
	"testing"
)

func TestError(t *testing.T) {
	t.Run("returns early when error is nil", func(t *testing.T) {
		Error(nil)
	})

	t.Run("accepts non-nil values without panicking", func(t *testing.T) {
		Error("test error", "key", "value")
		Error("some error")
		Error(newDefaultTestError("runtime error"), "key", "value")
	})
}

func TestPairsToProps(t *testing.T) {
	t.Run("sets valid key value pairs", func(t *testing.T) {
		got := pairsToProps("foo", "bar", "count", 3)
		want := Properties{
			"foo":   "bar",
			"count": 3,
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("pairsToProps() = %#v, want %#v", got, want)
		}
	})

	t.Run("returns empty properties for odd pairs", func(t *testing.T) {
		got := pairsToProps("foo", "bar", "count")
		if len(got) != 0 {
			t.Fatalf("pairsToProps() should return empty properties, got %#v", got)
		}
	})

	t.Run("ignores non-string key and continues", func(t *testing.T) {
		got := pairsToProps(123, "bad", "ok", true)
		want := Properties{"ok": true}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("pairsToProps() = %#v, want %#v", got, want)
		}
	})
}

// newDefaultTestError creates a test error that mimics runtime panic
// errors. This helps us testing that the Error function can handle various
// error types, including those that might be passed from a panic recovery
// scenario.
func newDefaultTestError(s string) error {
	return testError(s)
}

type testError string

func (e testError) Error() string {
	return string(e)
}
